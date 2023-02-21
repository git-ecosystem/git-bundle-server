package utils

import (
	"context"
	"flag"
	"fmt"
	"strconv"
)

// Helpers

// Forward declaration (kinda) of argParser.
// 'argparse.argParser' is private, but we want to be able to pass instances
// of it to functions, so we need to define an interface that includes the
// functions we want to call from the parser.
type argParser interface {
	Lookup(name string) *flag.Flag
	Usage(ctx context.Context, errFmt string, args ...any)
}

func GetFlagValue[T any](parser argParser, name string) T {
	flagVal := parser.Lookup(name)
	if flagVal == nil {
		panic(fmt.Sprintf("flag '--%s' is undefined", name))
	}

	flagGetter, ok := flagVal.Value.(flag.Getter)
	if !ok {
		panic(fmt.Sprintf("flag '--%s' is invalid (does not implement flag.Getter)", name))
	}

	value, ok := flagGetter.Get().(T)
	if !ok {
		panic(fmt.Sprintf("flag '--%s' is invalid (cannot cast to appropriate type)", name))
	}

	return value
}

// Sets of flags shared between multiple commands/programs

func WebServerFlags(parser argParser) (*flag.FlagSet, func(context.Context)) {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	port := f.String("port", "8080", "The port on which the server should be hosted")
	cert := f.String("cert", "", "The path to the X.509 SSL certificate file to use in securely hosting the server")
	key := f.String("key", "", "The path to the certificate's private key")

	// Function to call for additional arg validation (may exit with 'Usage()')
	validationFunc := func(ctx context.Context) {
		p, err := strconv.Atoi(*port)
		if err != nil || p < 0 || p > 65535 {
			parser.Usage(ctx, "Invalid port '%s'.", *port)
		}
		if (*cert == "") != (*key == "") {
			parser.Usage(ctx, "Both '--cert' and '--key' are needed to specify SSL configuration.")
		}
	}

	return f, validationFunc
}
