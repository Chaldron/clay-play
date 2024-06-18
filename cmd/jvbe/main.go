package main

import (
	"fmt"
	"os"
)

type program interface {
	name() string
	parse() error
	run() error
}

func run() error {
	var programArgs []string
	var input string
	if len(os.Args) < 2 {
		input = "app"
	} else {
		programArgs = os.Args[2:]
		input = os.Args[1]
	}

	appProgram := newAppProgram(programArgs)
	migrationProgram := newMigrationProgram(programArgs)

	programs := []program{
		appProgram,
		migrationProgram,
	}

	for _, prog := range programs {
		if input == prog.name() {
			err := prog.parse()
			if err != nil {
				return err
			}
			return prog.run()
		}
	}

	return fmt.Errorf("unknown program: %s", input)
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
