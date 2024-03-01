package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

var Cfg = &Config{}

type Config struct {
	Server struct {
		Name      string `yaml:"name"`
		Testnet   bool   `yaml:"testnet"`
		RpcListen string `yaml:"rpc_listen"`
	} `yaml:"server"`
	Chain struct {
		Url      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"chain"`
	DB struct {
		Mysql   Mysql `yaml:"mysql"`
		Indexer Mysql `yaml:"indexer"`
	} `yaml:"db"`
}

type Mysql struct {
	Addr     string `yaml:"addr"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

func Init(configPath string) error {
	configFile, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer configFile.Close()
	if err := yaml.NewDecoder(configFile).Decode(Cfg); err != nil {
		return err
	}
	return nil
}
