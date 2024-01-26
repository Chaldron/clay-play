package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DbConn   string `yaml:"db_conn"`
	Port     int    `yaml:"port"`
    FbAppId string `yaml:"fb_app_id"`
    FbSecret string `yaml:"fb_secret"`
}

func ReadFile(src string) (*Config, error) {
	b, err := os.ReadFile(src)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = yaml.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
