package main

import (
	"log"
	"os"
)

func main() {
	cmds := all()

	if len(os.Args) < 2 {
		log.Fatal("usage: git-bundle-server <command> [<options>]\n")
		return
	}

	for i := 0; i < len(cmds); i++ {
		if cmds[i].subcommand() == os.Args[1] {
			err := cmds[i].run(os.Args[2:])
			if err != nil {
				log.Fatal("Failed with error: ", err)
			}
		}
	}
}
