package argparse

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/github/git-bundle-server/internal/log"
)

// For consistency with 'flag', use 2 as the usage-related error code
const usageExitCode int = 2

type positionalArg struct {
	name        string
	description string
	value       interface{}
}

type argParser struct {
	// State
	isTopLevel bool
	parsed     bool
	argOffset  int

	// Pre-parsing
	subcommands    map[string]Subcommand
	positionalArgs []*positionalArg

	// Post-parsing
	selectedSubcommand Subcommand

	logger log.TraceLogger
	flag.FlagSet
}

func NewArgParser(logger log.TraceLogger, usageString string) *argParser {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	a := &argParser{
		isTopLevel:  false,
		parsed:      false,
		argOffset:   0,
		subcommands: make(map[string]Subcommand),
		logger:      logger,
		FlagSet:     *flagSet,
	}

	a.FlagSet.Usage = func() {
		out := a.FlagSet.Output()
		fmt.Fprintf(out, "usage: %s\n\n", usageString)

		// Print flags (if any)
		flagCount := 0
		a.FlagSet.VisitAll(func(f *flag.Flag) { flagCount++ })
		if flagCount > 0 {
			fmt.Fprintln(out, "Flags:")
			a.FlagSet.PrintDefaults()
			fmt.Fprint(out, "\n")
		}

		// Print subcommands (if any)
		if len(a.subcommands) > 0 {
			if a.isTopLevel {
				fmt.Fprintln(out, "Commands:")
			} else {
				fmt.Fprintln(out, "Subcommands:")
			}
			a.printSubcommands()
			fmt.Fprint(out, "\n")
		}
	}

	return a
}

func (a *argParser) SetIsTopLevel(isTopLevel bool) {
	a.isTopLevel = isTopLevel
}

func (a *argParser) printSubcommands() {
	out := a.FlagSet.Output()
	for _, subcommand := range a.subcommands {
		fmt.Fprintf(out, "  %s\n    \t%s\n",
			subcommand.Name(),
			strings.ReplaceAll(strings.TrimSpace(subcommand.Description()), "\n", "\n    \t"),
		)
	}
}

func (a *argParser) Subcommand(subcommand Subcommand) {
	a.subcommands[subcommand.Name()] = subcommand
}

func (a *argParser) PositionalStringVar(name string, description string, arg *string) {
	a.positionalArgs = append(a.positionalArgs, &positionalArg{
		name:        name,
		description: description,
		value:       arg,
	})
}

func (a *argParser) PositionalString(name string, description string) *string {
	arg := new(string)
	a.PositionalStringVar(name, description, arg)
	return arg
}

func (a *argParser) PositionalListVar(name string, description string, arg *[]string) {
	a.positionalArgs = append(a.positionalArgs, &positionalArg{
		name:        name,
		description: description,
		value:       arg,
	})
}

func (a *argParser) PositionalList(name string, description string) *[]string {
	arg := &[]string{}
	a.PositionalListVar(name, description, arg)
	return arg
}

func (a *argParser) Parse(ctx context.Context, args []string) {
	if a.parsed {
		// Do nothing if we've already parsed args
		return
	}

	// Validate
	if len(a.subcommands) > 0 && len(a.positionalArgs) > 0 {
		panic("cannot mix subcommands and positional args")
	}
	for i, positionalArg := range a.positionalArgs {
		if i < len(a.positionalArgs)-1 {
			// Only the last positional arg can be a list
			_, isList := positionalArg.value.(*[]string)
			if isList {
				panic("only the last positional arg can be a list type")
			}
		}
	}

	err := a.FlagSet.Parse(args)
	if err != nil {
		// The error was already printed (via a.FlagSet.Usage()), so we
		// just need to exit
		a.logger.Error(ctx, err)
		a.logger.Exit(ctx, usageExitCode)
	}

	if len(a.subcommands) > 0 {
		// Parse subcommand, if applicable
		if a.FlagSet.NArg() == 0 {
			a.Usage(ctx, "Please specify a subcommand")
		}

		subcommand, exists := a.subcommands[a.FlagSet.Arg(0)]
		if !exists {
			a.Usage(ctx, "Invalid subcommand '%s'", a.FlagSet.Arg(0))
		} else {
			a.selectedSubcommand = subcommand
			a.argOffset++
		}
	} else {
		// Handle positional args
		for _, arg := range a.positionalArgs {
			// First, try single string case
			sPtr, isStr := arg.value.(*string)
			if isStr {
				if a.NArg() == 0 {
					a.Usage(ctx, "No value specified for required argument '%s'", arg.name)
				}

				*sPtr = a.Arg(0)
				a.argOffset++
				continue
			}

			// Next, try list case
			lPtr, isList := arg.value.(*[]string)
			if isList {
				*lPtr = a.Args()
				a.argOffset += a.NArg()
				break
			}

			panic("Positional arg has invalid type")
		}

		if a.NArg() != 0 {
			// If not using subcommands, all args should be accounted for
			// Exit with usage if not
			a.Usage(ctx, "Unused arguments specified: %s", strings.Join(a.Args(), " "))
		}
	}

	a.parsed = true
}

func (a *argParser) Arg(index int) string {
	return a.FlagSet.Arg(index + a.argOffset)
}

func (a *argParser) Args() []string {
	return a.FlagSet.Args()[a.argOffset:]
}

func (a *argParser) NArg() int {
	if a.FlagSet.NArg() <= a.argOffset {
		return 0
	} else {
		return a.FlagSet.NArg() - a.argOffset
	}
}

func (a *argParser) InvokeSubcommand(ctx context.Context) error {
	if !a.parsed || a.selectedSubcommand == nil {
		panic("subcommand has not been parsed")
	}

	if a.isTopLevel {
		a.logger.LogCommand(ctx, a.selectedSubcommand.Name())
	}

	return a.selectedSubcommand.Run(ctx, a.Args())
}

func (a *argParser) Usage(ctx context.Context, errFmt string, args ...any) {
	fmt.Fprintf(a.FlagSet.Output(), errFmt+"\n", args...)
	a.FlagSet.Usage()

	a.logger.Errorf(ctx, errFmt, args...)
	a.logger.Exit(ctx, usageExitCode)
}
