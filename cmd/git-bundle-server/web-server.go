package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/github/git-bundle-server/internal/argparse"
	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/daemon"
)

type webServer struct {
	user       common.UserProvider
	cmdExec    common.CommandExecutor
	fileSystem common.FileSystem
}

func NewWebServerCommand() *webServer {
	// Create dependencies
	return &webServer{
		user:       common.NewUserProvider(),
		cmdExec:    common.NewCommandExecutor(),
		fileSystem: common.NewFileSystem(),
	}
}

func (webServer) Name() string {
	return "web-server"
}

func (webServer) Description() string {
	return `Manage the web server hosting bundle content`
}

func (w *webServer) getDaemonConfig() (*daemon.DaemonConfig, error) {
	// Find git-bundle-web-server
	// First, search for it on the path
	programPath, err := exec.LookPath("git-bundle-web-server")
	if err != nil {
		if errors.Is(err, exec.ErrDot) {
			// Result is a relative path
			programPath, err = filepath.Abs(programPath)
			if err != nil {
				return nil, fmt.Errorf("could not get absolute path to program: %w", err)
			}
		} else {
			// Fall back on looking for it in the same directory as the currently-running executable
			exePath, err := os.Executable()
			if err != nil {
				return nil, fmt.Errorf("failed to get path to current executable: %w", err)
			}
			exeDir := filepath.Dir(exePath)
			if err != nil {
				return nil, fmt.Errorf("failed to get parent dir of current executable: %w", err)
			}

			programPath = filepath.Join(exeDir, "git-bundle-web-server")
			programExists, err := w.fileSystem.FileExists(programPath)
			if err != nil {
				return nil, fmt.Errorf("could not determine whether path to 'git-bundle-web-server' exists: %w", err)
			} else if !programExists {
				return nil, fmt.Errorf("could not find path to 'git-bundle-web-server'")
			}
		}
	}

	return &daemon.DaemonConfig{
		Label:       "com.github.bundleserver",
		Description: "Web server hosting Git Bundle Server content",
		Program:     programPath,
	}, nil
}

func (w *webServer) startServer(args []string) error {
	// Parse subcommand arguments
	parser := argparse.NewArgParser("git-bundle-server web-server start [-f|--force]")
	force := parser.Bool("force", false, "Whether to force reconfiguration of the web server daemon")
	parser.BoolVar(force, "f", false, "Alias of --force")
	parser.Parse(args)

	d, err := daemon.NewDaemonProvider(w.user, w.cmdExec, w.fileSystem)
	if err != nil {
		return err
	}

	config, err := w.getDaemonConfig()
	if err != nil {
		return err
	}

	err = d.Create(config, *force)
	if err != nil {
		return err
	}

	err = d.Start(config.Label)
	if err != nil {
		return err
	}

	return nil
}

func (w *webServer) stopServer(args []string) error {
	// Parse subcommand arguments
	parser := argparse.NewArgParser("git-bundle-server web-server stop")
	parser.Parse(args)

	d, err := daemon.NewDaemonProvider(w.user, w.cmdExec, w.fileSystem)
	if err != nil {
		return err
	}

	config, err := w.getDaemonConfig()
	if err != nil {
		return err
	}

	err = d.Stop(config.Label)
	if err != nil {
		return err
	}

	return nil
}

func (w *webServer) Run(args []string) error {
	// Parse command arguments
	parser := argparse.NewArgParser("git-bundle-server web-server (start|stop) <options>")
	parser.Subcommand(argparse.NewSubcommand("start", "Start the web server", w.startServer))
	parser.Subcommand(argparse.NewSubcommand("stop", "Stop the web server", w.stopServer))
	parser.Parse(args)

	return parser.InvokeSubcommand()
}
