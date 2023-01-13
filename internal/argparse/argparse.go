package argparse

import (
	"flag"
	"fmt"
	"os"
)

type argParser struct {
	flag.FlagSet
}

func NewArgParser(usageString string) *argParser {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	a := &argParser{
		FlagSet: *flagSet,
	}

	a.FlagSet.Usage = func() {
		out := a.FlagSet.Output()
		fmt.Fprintf(out, "usage: %s\n\n", usageString)
	}

	return a
}

func (a *argParser) Parse(args []string) {
	err := a.FlagSet.Parse(args)
	if err != nil {
		panic("argParser FlagSet error handling should be 'ExitOnError', but error encountered")
	}
}

func (a *argParser) Usage(errFmt string, args ...any) {
	fmt.Fprintf(a.FlagSet.Output(), errFmt+"\n", args...)
	a.FlagSet.Usage()

	// Exit with error code 2 to match flag.Parse() behavior
	os.Exit(2)
}
