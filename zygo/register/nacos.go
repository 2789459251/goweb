package register

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func CreateNacosClient() (naming_client.INamingClient, error) {

	// 创建clientConfig的另一种方式
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(""), //当namespace是public时，此处填空字符串。
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
	)

	// 至少一个ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "127.0.0.1",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}

	// 创建服务发现客户端的另一种方式 (推荐)
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return nil, err
	}
	return namingClient, nil
}

// 服务注册实例
func RegisterService(namingClient naming_client.INamingClient, serviceName string, host string, port uint64) error {
	_, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        port,
		ServiceName: serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "shanghai"},
		//ClusterName: "cluster-a", // 默认值DEFAULT
		//GroupName:   "group-a",   // 默认值DEFAULT_GROUP
	})
	return err

}

/*获取健康实例，加权轮训*/
func GetService(namingClient naming_client.INamingClient, serviceName string) (host string, port uint64, err error) {
	instance, err := namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		//GroupName:   "group-a",             // 默认值DEFAULT_GROUP
		//Clusters:    []string{"cluster-a"}, // 默认值DEFAULT
	})
	if err != nil {
		return "", uint64(0), err
	}
	return instance.Ip, instance.Port, nil
}
