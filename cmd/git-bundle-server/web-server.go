package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/github/git-bundle-server/cmd/utils"
	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/daemon"
	"github.com/github/git-bundle-server/internal/log"
)

type webServerCmd struct {
	logger    log.TraceLogger
	container *utils.DependencyContainer
}

func NewWebServerCommand(logger log.TraceLogger, container *utils.DependencyContainer) argparse.Subcommand {
	return &webServerCmd{
		logger:    logger,
		container: container,
	}
}

func (webServerCmd) Name() string {
	return "web-server"
}

func (webServerCmd) Description() string {
	return `Manage the web server hosting bundle content`
}

func (w *webServerCmd) getDaemonConfig(ctx context.Context) (*daemon.DaemonConfig, error) {
	// Find git-bundle-web-server
	fileSystem := utils.GetDependency[common.FileSystem](ctx, w.container)
	programPath, err := fileSystem.GetLocalExecutable("git-bundle-web-server")
	if err != nil {
		return nil, w.logger.Error(ctx, err)
	}

	return &daemon.DaemonConfig{
		Label:       "com.github.gitbundleserver",
		Description: "Web server hosting Git Bundle Server content",
		Program:     programPath,
	}, nil
}

func (w *webServerCmd) startServer(ctx context.Context, args []string) error {
	// Parse subcommand arguments
	parser := argparse.NewArgParser(w.logger, "git-bundle-server web-server start [-f|--force]")

	// Args for 'git-bundle-server web-server start'
	force := parser.Bool("force", false, "Force reconfiguration of the web server daemon")
	parser.BoolVar(force, "f", false, "Alias of --force")

	// Arguments passed through to 'git-bundle-web-server'
	webServerFlags, validate := utils.WebServerFlags(parser)
	webServerFlags.VisitAll(func(f *flag.Flag) {
		parser.Var(f.Value, f.Name, fmt.Sprintf("[Web server] %s", f.Usage))
	})

	parser.Parse(ctx, args)
	validate(ctx)

	d := utils.GetDependency[daemon.DaemonProvider](ctx, w.container)

	config, err := w.getDaemonConfig(ctx)
	if err != nil {
		return w.logger.Error(ctx, err)
	}

	// Configure flags
	loopErr := error(nil)
	parser.Visit(func(f *flag.Flag) {
		if webServerFlags.Lookup(f.Name) != nil {
			value := f.Value.String()
			if f.Name == "cert" || f.Name == "key" {
				// Need the absolute value of the path
				value, err = filepath.Abs(value)
				if err != nil {
					if loopErr == nil {
						// NEEDSWORK: Only report the first error because Go
						// doesn't like it when you manually chain errors :(
						// Luckily, this is slated to change in v1.20, per
						// https://tip.golang.org/doc/go1.20#errors
						loopErr = fmt.Errorf("could not get absolute path of '%s': %w", f.Name, err)
					}
					return
				}
			}
			config.Arguments = append(config.Arguments, fmt.Sprintf("--%s", f.Name), value)
		}
	})
	if loopErr != nil {
		// Error happened in 'Visit'
		return w.logger.Error(ctx, loopErr)
	}

	err = d.Create(ctx, config, *force)
	if err != nil {
		return w.logger.Error(ctx, err)
	}

	err = d.Start(ctx, config.Label)
	if err != nil {
		return w.logger.Error(ctx, err)
	}

	return nil
}

func (w *webServerCmd) stopServer(ctx context.Context, args []string) error {
	// Parse subcommand arguments
	parser := argparse.NewArgParser(w.logger, "git-bundle-server web-server stop [--remove]")
	remove := parser.Bool("remove", false, "Remove the web server daemon configuration from the system after stopping")
	parser.Parse(ctx, args)

	d := utils.GetDependency[daemon.DaemonProvider](ctx, w.container)

	config, err := w.getDaemonConfig(ctx)
	if err != nil {
		return w.logger.Error(ctx, err)
	}

	err = d.Stop(ctx, config.Label)
	if err != nil {
		return w.logger.Error(ctx, err)
	}

	if *remove {
		err = d.Remove(ctx, config.Label)
		if err != nil {
			return w.logger.Error(ctx, err)
		}
	}

	return nil
}

func (w *webServerCmd) Run(ctx context.Context, args []string) error {
	// Parse command arguments
	parser := argparse.NewArgParser(w.logger, "git-bundle-server web-server (start|stop) <options>")
	parser.Subcommand(argparse.NewSubcommand("start", "Start the web server", w.startServer))
	parser.Subcommand(argparse.NewSubcommand("stop", "Stop the web server", w.stopServer))
	parser.Parse(ctx, args)

	return parser.InvokeSubcommand(ctx)
}
