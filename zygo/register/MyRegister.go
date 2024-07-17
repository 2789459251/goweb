package register

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"time"
)

type Option struct {
	Endpoints         []string      //节点
	DialTimeout       time.Duration //超时时间
	ServiceName       string
	Host              string
	Port              int
	NacosServerConfig []constant.ServerConfig
	NacosClientConfig *constant.ClientConfig
}

type MyRegister interface {
	CreateCli(option Option) error
	RegisterService(serviceName string, host string, port int) error
	GetValue(serviceName string) (string, error)
	Close() error
}
