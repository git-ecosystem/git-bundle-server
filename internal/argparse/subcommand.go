package argparse

type Subcommand interface {
	Name() string
	Description() string
	Run(args []string) error
}

type genericSubcommand struct {
	nameStr        string
	descriptionStr string
	runFunc        func([]string) error
}

func NewSubcommand(
	name string,
	description string,
	runFunc func([]string) error,
) *genericSubcommand {
	return &genericSubcommand{
		nameStr:        name,
		descriptionStr: description,
		runFunc:        runFunc,
	}
}

func (s *genericSubcommand) Name() string {
	return s.nameStr
}

func (s *genericSubcommand) Description() string {
	return s.descriptionStr
}

func (s *genericSubcommand) Run(args []string) error {
	return s.runFunc(args)
}
