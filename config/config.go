package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DbConn               string `yaml:"db_conn" env:"DB_CONN,required"`
	Port                 int    `yaml:"port" env:"PORT,required"`
	DefaultAdminPassword string `yaml:"default_admin_password" env:"DEFAULT_ADMIN_PASSWORD,required"`
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

func ReadEnv() (*Config, error) {
	conf := &Config{}
	err := env.Parse(conf)

	if err != nil {
		return nil, err
	}

	return conf, nil
}

func LoadFromCommandLineArgs(argv []string) (*Config, error) {
	flagSet := flag.NewFlagSet("app", flag.ExitOnError)
	configFilePath := flagSet.String("c", "./config.yaml", "path to config file")
	useConfigFromEnv := flagSet.Bool("e", false, "use config from environment")

	err := flagSet.Parse(argv)
	if err != nil {
		return nil, err
	}

	if *useConfigFromEnv {
		return ReadEnv()
	} else {
		return ReadFile(*configFilePath)
	}
}
