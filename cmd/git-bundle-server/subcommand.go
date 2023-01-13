package main

type Subcommand interface {
	Name() string
	run(args []string) error
}

func all() []Subcommand {
	return []Subcommand{
		Delete{},
		Init{},
		Start{},
		Stop{},
		Update{},
		UpdateAll{},
	}
}
