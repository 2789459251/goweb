package register

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strconv"
	"time"
)

type EtcdRegister struct {
	etcdClient *clientv3.Client
	balancer   *balancer
	resolver   *resolver
}

func (r *EtcdRegister) CreateCli(option Option) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   option.Endpoints,
		DialTimeout: option.DialTimeout,
	})
	r.etcdClient = cli
	r.balancer = &balancer{
		cache: make(map[string]*etcdManager),
	}
	r.resolver = &resolver{}
	return err
}

func (r *EtcdRegister) RegisterService(serviceName string, host string, port int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var ttl int64 = 10
	lease, err := r.etcdClient.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	em := NewManager(r.etcdClient, serviceName)
	r.balancer.cache[serviceName] = em
	endpoint := endPoint{
		addr:        fmt.Sprintf("%s:%d", host, port),
		serviceName: serviceName,
	}
	if err := em.AddEndPoint(ctx, serviceName, endpoint, lease); err != nil {
		return err
	}

	go r.balancer.watch(ctx, serviceName)

	for {
		select {
		case <-time.After(5 * time.Second):
			resp, err := r.etcdClient.KeepAliveOnce(ctx, lease.ID)
			if err != nil {
				return err
			}
			fmt.Printf("keep alive resp:%+v\n", resp)
		case <-ctx.Done():
			return errors.New("register service failed")
		}
	}
}

func (r *EtcdRegister) GetValue(serviceName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	endpoint, err := r.balancer.getEndPoint(ctx, serviceName)
	if err != nil {
		return "", err
	}

	return endpoint.addr, nil
}

func (r *EtcdRegister) Close() error {
	return r.etcdClient.Close()
}

type endPoint struct {
	addr        string
	serviceName string
	id          string
	healthy     bool
}

type etcdManager struct {
	client    *clientv3.Client
	target    string
	endPoints []*endPoint
}

type balancer struct {
	cache map[string]*etcdManager
	next  int
}

func (b *balancer) getEndPoint(ctx context.Context, serviceName string) (*endPoint, error) {
	em, ok := b.cache[serviceName]
	if !ok {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	var healthyEndPoints []*endPoint
	for _, ep := range em.endPoints {
		if ep.healthy {
			healthyEndPoints = append(healthyEndPoints, ep)
		}
	}

	if len(healthyEndPoints) == 0 {
		return nil, fmt.Errorf("no healthy endpoint available for service %s", serviceName)
	}

	index := b.next % len(healthyEndPoints)
	b.next = (b.next + 1) % len(healthyEndPoints)
	return healthyEndPoints[index], nil
}

func (b *balancer) remove(serviceName string, endpoint *endPoint) {
	em, ok := b.cache[serviceName]
	if !ok {
		return
	}

	for i, ep := range em.endPoints {
		if ep == endpoint {
			em.endPoints = append(em.endPoints[:i], em.endPoints[i+1:]...)
			break
		}
	}
}

func (b *balancer) watch(ctx context.Context, serviceName string) {
	for {
		select {
		case <-time.Tick(time.Second):
			// 监听服务实例状态变化,并更新 balancer 的缓存
		}
	}
}

type resolver struct {
}

func (em *etcdManager) AddEndPoint(ctx context.Context, serviceName string, point endPoint, lease *clientv3.LeaseGrantResponse) error {
	if em.target != serviceName {
		return errors.New(fmt.Sprintf("manager_target %s and service %s not match", em.target, serviceName))
	}
	point.id = strconv.FormatInt(int64(lease.ID), 10)
	em.endPoints = append(em.endPoints, &point)
	endpoint_, err := json.Marshal(em.endPoints)
	if err != nil {
		return err
	}
	ops := []clientv3.Op{clientv3.OpPut(serviceName, string(endpoint_), clientv3.WithLease(lease.ID))}
	em.client.KV.Txn(ctx).Then(ops...).Commit()
	return nil
}
func NewManager(cli *clientv3.Client, target string) *etcdManager {
	return &etcdManager{
		client:    cli,
		target:    target,
		endPoints: make([]*endPoint, 0),
	}
}
