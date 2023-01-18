package argparse

type Subcommand interface {
	Name() string
	Description() string
	Run(args []string) error
}
