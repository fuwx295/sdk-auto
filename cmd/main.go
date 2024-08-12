package main

import (
	"flag"
	"sdk-auto/pkg/config"
	"sdk-auto/pkg/httpserver"

	"github.com/spf13/viper"
)

func main() {
	configPath := flag.String("config", "sdk-auto.yml", "Configuration file")
	flag.Parse()

	viper := viper.New()
	viper.SetConfigFile(*configPath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	autoCfg := &config.AutoConfig{}
	viper.UnmarshalKey("auto", autoCfg)

	httpserver.StartHttpServer(autoCfg)
}
