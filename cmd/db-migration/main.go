package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/Chaldron/clay-play/config"
	"github.com/Chaldron/clay-play/db"
	"github.com/Chaldron/clay-play/logger"
	_ "github.com/mattn/go-sqlite3"
)

func run() error {
	log := logger.NewStdLogger()

	flagSet := flag.NewFlagSet("db-migration", flag.ExitOnError)
	configFilePath := flagSet.String("c", "./config.yaml", "path to config file")
	useConfigFromEnv := flagSet.Bool("e", false, "use config from environment")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	conf, err := config.LoadFromCommandLineArgs(*configFilePath, *useConfigFromEnv)
	if err != nil {
		return err
	}

	db, err := db.Connect(conf.DbConn, conf.DefaultAdminPassword, log)
	if err != nil {
		return err
	}

	action := flagSet.Arg(0)
	version := flagSet.Arg(1)

	switch action {
	case "create":
		return db.MigrationCreate(version)
	case "downgrade":
		v, err := strconv.ParseInt(version, 10, 64)
		if err != nil {
			return err
		}
		return db.MigrationDownTo(v)
	default:
		return errors.New("invalid action: " + action)
	}
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
