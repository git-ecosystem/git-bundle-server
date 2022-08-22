package main

import (
	"fmt"
)

type Update struct{}

func (Update) subcommand() string {
	return "update"
}

func (Update) run(args []string) error {
	fmt.Printf("Found Update method!\n")

	for _, arg := range args {
		fmt.Printf("%s\n", arg)
	}

	return nil
}
