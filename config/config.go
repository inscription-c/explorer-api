package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

var Cfg = &Config{}

type Config struct {
	Server struct {
		Name        string `yaml:"name"`
		Testnet     bool   `yaml:"testnet"`
		RpcListen   string `yaml:"rpc_listen"`
		EnablePProf bool   `yaml:"pprof"`
		Prometheus  bool   `yaml:"prometheus"`
	} `yaml:"server"`
	Chain struct {
		Url         string `yaml:"url"`
		Username    string `yaml:"username"`
		Password    string `yaml:"password"`
		StartHeight uint32 `yaml:"start_height"`
	} `yaml:"chain"`
	DB struct {
		Mysql   Mysql `yaml:"mysql"`
		Indexer Mysql `yaml:"indexer"`
	} `yaml:"db"`
	Sentry struct {
		Dsn              string  `yaml:"dsn"`
		TracesSampleRate float64 `yaml:"traces_sample_rate"`
	} `yaml:"sentry"`
	Origins []string `yaml:"origins"`
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
