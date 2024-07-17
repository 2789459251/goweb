package register

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type EtcdRegister struct {
	etcdClient *clientv3.Client
}

func CreateEtcdCli(option Option) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   option.Endpoints,   //节点
		DialTimeout: option.DialTimeout, //超过5秒钟连不上超时
	})
	return cli, err
}

/*注册服务*/
func RegisterEtcdService(cli *clientv3.Client, serviceName string, host string, port int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := cli.Put(ctx, serviceName, fmt.Sprintf("%s:%d", host, port))
	defer cancel()
	return err
}

/*获取服务*/
func GetEtcdValue(cli *clientv3.Client, serviceName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	v, err := cli.Get(ctx, serviceName)
	defer cancel()
	kvs := v.Kvs
	if len(kvs) == 0 {
		return "", fmt.Errorf("service %s not exist", serviceName)
	}
	return string(kvs[0].Value), err
}

func (r *EtcdRegister) CreateCli(option Option) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   option.Endpoints,   //节点
		DialTimeout: option.DialTimeout, //超过5秒钟连不上超时
	})
	r.etcdClient = cli
	return err
}
func (r *EtcdRegister) RegisterService(serviceName string, host string, port int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := r.etcdClient.Put(ctx, serviceName, fmt.Sprintf("%s:%d", host, port))
	defer cancel()
	return err
}

// 只支持一个serveiceName一个值
func (r *EtcdRegister) GetValue(serviceName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	v, err := r.etcdClient.Get(ctx, serviceName)
	defer cancel()
	kvs := v.Kvs
	if len(kvs) == 0 {
		return "", fmt.Errorf("service %s not exist", serviceName)
	}
	return string(kvs[0].Value), err
}
func (r *EtcdRegister) Close() error {
	return r.etcdClient.Close()
}
