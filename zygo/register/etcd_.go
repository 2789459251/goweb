package register

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
)

func (e *EtcdRegister) DiscoverService(serviceName string) ([]string, error) {
	// 使用 etcdClient 从 etcd 获取服务列表
	// 返回服务的地址列表
	ctx := context.Background()
	key := fmt.Sprintf("/service/%s", serviceName)
	resp, err := e.etcdClient.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var addrs []string
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}

	return addrs, nil
}
func (e *EtcdRegister) registerService(serviceName, host string, port int) error {
	// 使用 etcdClient 向 etcd 注册服务
	// 例如：设置一个 key 为 "/service/{serviceName}/{host}_{port}"
	ctx := context.Background()
	key := fmt.Sprintf("/service/%s/%s_%d", serviceName, host, port)
	value := fmt.Sprintf("%s:%d", host, port)
	lease := clientv3.NewLease(e.etcdClient)

	// Grant a lease for 10 seconds
	leaseResp, err := lease.Grant(ctx, 10)
	if err != nil {
		return err
	}
	// 开始租约续期的 keepalive 循环
	keepAliveChan, err := lease.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		// 处理启动 keepalive 时的错误
		return err
	}

	// 使用 defer 关键字确保在函数退出前关闭连接
	defer func() {
		// 确保关闭 keepalive 通道，这将停止续期租约
		close(keepAliveChan)
	}()

	// 在一个单独的 goroutine 中处理 keepalive 响应
	go func() {
		for {
			select {
			case <-ctx.Done():
				// 如果上下文完成或取消，则退出 goroutine
				return
			case kaResp, ok := <-keepAliveChan:
				if !ok {
					// 如果 keepAliveChan 通道关闭，可能是由于客户端与 etcd 服务器的连接断开
					log.Printf("keepalive channel closed for lease ID: %d", leaseResp.ID)
					return
				}
				// 这里没有错误处理，因为 LeaseKeepAliveResponse 没有 Err 字段
				// 但是您可以根据实际情况记录租约的 TTL 或其他信息
				log.Printf("lease ID: %d, TTL: %d", kaResp.ID, kaResp.TTL)
			}
		}
	}()

	// Put the key with the lease
	_, err = e.etcdClient.Put(ctx, key, value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	return nil
}
