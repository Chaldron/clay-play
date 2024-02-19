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
	if len(os.Args) < 2 {
		return fmt.Errorf("specify a program")
	}

	programArgs := os.Args[2:]
	appProgram := newAppProgram(programArgs)
	migrationProgram := newMigrationProgram(programArgs)

	programs := []program{
		appProgram,
		migrationProgram,
	}

	input := os.Args[1]
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
