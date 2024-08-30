package config

import (
	"flag"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DbConn               string `yaml:"db_conn"`
	Port                 int    `yaml:"port"`
	DefaultAdminPassword string `yaml:"default_admin_password"`
}

func ReadFile(src string) (*Config, error) {
	bytes, err := os.ReadFile(src)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = yaml.Unmarshal(bytes, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func LoadFromCommandLineArgs(argv []string) (*Config, error) {
	flagSet := flag.NewFlagSet("app", flag.ExitOnError)
	configFilePath := flagSet.String("c", "./config.yaml", "path to config file")

	err := flagSet.Parse(argv)
	if err != nil {
		return nil, err
	}

	return ReadFile(*configFilePath)
}
