package daemon

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/github/git-bundle-server/internal/common"
)

const launchTemplate string = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key><string>{{.Label}}</string>
    <key>Program</key><string>{{.Program}}</string>
    <key>StandardOutPath</key><string>{{.StdOut}}</string>
    <key>StandardErrorPath</key><string>{{.StdErr}}</string>
  </dict>
</plist>
`

const domainFormat string = "gui/%s"

const LaunchdServiceNotFoundErrorCode int = 113

type launchdConfig struct {
	DaemonConfig
	StdOut string
	StdErr string
}

type launchd struct {
	user       common.UserProvider
	cmdExec    common.CommandExecutor
	fileSystem common.FileSystem
}

func NewLaunchdProvider(
	u common.UserProvider,
	c common.CommandExecutor,
	fs common.FileSystem,
) DaemonProvider {
	return &launchd{
		user:       u,
		cmdExec:    c,
		fileSystem: fs,
	}
}

func (l *launchd) isBootstrapped(serviceTarget string) (bool, error) {
	// run 'launchctl print' on given service target to see if it exists
	exitCode, err := l.cmdExec.Run("launchctl", "print", serviceTarget)
	if err != nil {
		return false, err
	}

	if exitCode == 0 {
		return true, nil
	} else if exitCode == LaunchdServiceNotFoundErrorCode {
		return false, nil
	} else {
		return false, fmt.Errorf("could not determine if service '%s' is bootstrapped: "+
			"'launchctl print' exited with status '%d'", serviceTarget, exitCode)
	}
}

func (l *launchd) bootstrapFile(domain string, filename string) error {
	// run 'launchctl bootstrap' on given domain & file
	exitCode, err := l.cmdExec.Run("launchctl", "bootstrap", domain, filename)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("'launchctl bootstrap' exited with status %d", exitCode)
	}

	return nil
}

func (l *launchd) bootoutFile(domain string, filename string) error {
	// run 'launchctl bootout' on given domain & file
	exitCode, err := l.cmdExec.Run("launchctl", "bootout", domain, filename)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("'launchctl bootout' exited with status %d", exitCode)
	}

	return nil
}

func (l *launchd) Create(config *DaemonConfig, force bool) error {
	// Add launchd-specific config
	lConfig := &launchdConfig{
		DaemonConfig: *config,
		StdOut:       "/dev/null",
		StdErr:       "/dev/null",
	}

	// Generate the configuration
	var newPlist bytes.Buffer
	t, err := template.New(config.Label).Parse(launchTemplate)
	if err != nil {
		return fmt.Errorf("unable to generate launchd configuration: %w", err)
	}
	t.Execute(&newPlist, lConfig)

	// Check the existing file - if it's the same as the new content, do not overwrite
	user, err := l.user.CurrentUser()
	if err != nil {
		return fmt.Errorf("could not get current user for launchd service: %w", err)
	}

	filename := filepath.Join(user.HomeDir, "Library", "LaunchAgents", fmt.Sprintf("%s.plist", config.Label))
	domainTarget := fmt.Sprintf(domainFormat, user.Uid)
	serviceTarget := fmt.Sprintf("%s/%s", domainTarget, config.Label)

	alreadyLoaded, err := l.isBootstrapped(serviceTarget)
	if err != nil {
		return err
	}

	// First, verify whether the file exists
	// TODO: only overwrite file if file contents have changed
	fileExists, err := l.fileSystem.FileExists(filename)
	if err != nil {
		return fmt.Errorf("could not determine whether plist '%s' exists: %w", filename, err)
	}

	if alreadyLoaded && !fileExists {
		// Abort on corrupted configuration
		return fmt.Errorf("service target '%s' is bootstrapped, but its plist doesn't exist", serviceTarget)
	}

	if !force && alreadyLoaded {
		// Not forcing a refresh of the file, so we do nothing
		return nil
	}

	// Otherwise, write & bootstrap the file
	if alreadyLoaded {
		// Unload the old file, if necessary
		l.bootoutFile(domainTarget, filename)
	}

	if !fileExists || force {
		err = l.fileSystem.WriteFile(filename, newPlist.Bytes())
		if err != nil {
			return fmt.Errorf("unable to overwrite plist file: %w", err)
		}
	}

	err = l.bootstrapFile(domainTarget, filename)
	if err != nil {
		return fmt.Errorf("could not bootstrap daemon process '%s': %w", config.Label, err)
	}

	return nil
}

func (l *launchd) Start(label string) error {
	user, err := l.user.CurrentUser()
	if err != nil {
		return fmt.Errorf("could not get current user for launchd service: %w", err)
	}

	domainTarget := fmt.Sprintf(domainFormat, user.Uid)
	serviceTarget := fmt.Sprintf("%s/%s", domainTarget, label)
	exitCode, err := l.cmdExec.Run("launchctl", "kickstart", serviceTarget)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("'launchctl kickstart' exited with status %d", exitCode)
	}

	return nil
}

func (l *launchd) Stop(label string) error {
	user, err := l.user.CurrentUser()
	if err != nil {
		return fmt.Errorf("could not get current user for launchd service: %w", err)
	}

	domainTarget := fmt.Sprintf(domainFormat, user.Uid)
	serviceTarget := fmt.Sprintf("%s/%s", domainTarget, label)
	exitCode, err := l.cmdExec.Run("launchctl", "kill", "SIGINT", serviceTarget)
	if err != nil {
		return err
	}

	// Don't throw an error if the service hasn't been bootstrapped
	if exitCode != 0 && exitCode != LaunchdServiceNotFoundErrorCode {
		return fmt.Errorf("'launchctl kill' exited with status %d", exitCode)
	}

	return nil
}
