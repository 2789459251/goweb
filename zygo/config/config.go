package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"os"
	"web/zygo/mylog"
)

var Conf = &MyConfig{
	logger: mylog.Default(),
}

type MyConfig struct {
	logger   *mylog.Logger
	Log      map[string]any
	Pool     map[string]any
	Mysql    map[string]any
	Template map[string]any
}

func init() {
	loadToml()
}

// 加载配置文件
func loadToml() {
	configFile := flag.String("conf", "conf/app.toml", "app config file")
	flag.Parse()
	if _, err := os.Stat(*configFile); err != nil {
		Conf.logger.Info("conf/app.toml file not load，because not exist")
		return
	}
	_, err := toml.DecodeFile(*configFile, Conf)
	if err != nil {
		Conf.logger.Info("conf/app.toml decode fail check format")
		return
	}
}
