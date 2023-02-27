package argparse

import "context"

type Subcommand interface {
	Name() string
	Description() string
	Run(ctx context.Context, args []string) error
}

type genericSubcommand struct {
	nameStr        string
	descriptionStr string
	runFunc        func(context.Context, []string) error
}

func NewSubcommand(
	name string,
	description string,
	runFunc func(context.Context, []string) error,
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

func (s *genericSubcommand) Run(ctx context.Context, args []string) error {
	return s.runFunc(ctx, args)
}
