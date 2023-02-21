package daemon

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"path/filepath"

	"github.com/github/git-bundle-server/internal/common"
	"github.com/github/git-bundle-server/internal/utils"
)

type xmlItem struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type xmlArray struct {
	XMLName  xml.Name
	Elements []interface{} `xml:",any"`
}

type plist struct {
	XMLName xml.Name `xml:"plist"`
	Version string   `xml:"version,attr"`
	Config  xmlArray `xml:"dict"`
}

const plistHeader string = `<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`

func xmlName(name string) xml.Name {
	return xml.Name{Local: name}
}

func (p *plist) addKeyValue(key string, value any) {
	p.Config.Elements = append(p.Config.Elements, xmlItem{XMLName: xmlName("key"), Value: key})
	switch value := value.(type) {
	case string:
		p.Config.Elements = append(p.Config.Elements, xmlItem{XMLName: xmlName("string"), Value: value})
	case []string:
		p.Config.Elements = append(p.Config.Elements,
			xmlArray{
				XMLName: xmlName("array"),
				Elements: utils.Map(value, func(e string) interface{} {
					return xmlItem{XMLName: xmlName("string"), Value: e}
				}),
			},
		)
	default:
		panic("Invalid value type in 'addKeyValue'")
	}
}

const domainFormat string = "user/%s"

const LaunchdNoSuchProcessErrorCode int = 3
const LaunchdServiceNotFoundErrorCode int = 113

type launchdConfig struct {
	DaemonConfig
	LimitLoadToSessionType string
	StdOut                 string
	StdErr                 string
}

func (c *launchdConfig) toPlist() *plist {
	p := &plist{
		Version: "1.0",
		Config:  xmlArray{Elements: []interface{}{}},
	}
	p.addKeyValue("Label", c.Label)
	p.addKeyValue("Program", c.Program)
	p.addKeyValue("LimitLoadToSessionType", c.LimitLoadToSessionType)
	p.addKeyValue("StandardOutPath", c.StdOut)
	p.addKeyValue("StandardErrorPath", c.StdErr)

	// IMPORTANT!!!
	// You must explicitly set the first argument to the executable path
	// because 'ProgramArguments' maps directly 'argv' in 'execvp'. The
	// programs calling this library likely will, by convention, assume the
	// first element of 'argv' is the executing program.
	// See https://www.unix.com/man-page/osx/5/launchd.plist/ and
	// https://man7.org/linux/man-pages/man3/execvp.3.html for more details.
	args := make([]string, len(c.Arguments)+1)
	args[0] = c.Program
	copy(args[1:], c.Arguments[:])
	p.addKeyValue("ProgramArguments", args)

	return p
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

func (l *launchd) isBootstrapped(ctx context.Context, serviceTarget string) (bool, error) {
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

func (l *launchd) bootstrapFile(ctx context.Context, domain string, filename string) error {
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

func (l *launchd) bootout(ctx context.Context, serviceTarget string) (bool, error) {
	// run 'launchctl bootout' on given service target
	exitCode, err := l.cmdExec.Run("launchctl", "bootout", serviceTarget)
	if err != nil {
		return false, err
	}

	if exitCode == 0 {
		return true, nil
	} else if exitCode == LaunchdNoSuchProcessErrorCode {
		return false, nil
	} else {
		return false, fmt.Errorf("'launchctl bootout' failed with status %d", exitCode)
	}
}

func (l *launchd) Create(ctx context.Context, config *DaemonConfig, force bool) error {
	// Add launchd-specific config
	lConfig := &launchdConfig{
		DaemonConfig:           *config,
		LimitLoadToSessionType: "Background",
		StdOut:                 "/dev/null",
		StdErr:                 "/dev/null",
	}

	// Generate the configuration
	var newPlist bytes.Buffer
	newPlist.WriteString(xml.Header)
	newPlist.WriteString(plistHeader)
	encoder := xml.NewEncoder(&newPlist)
	encoder.Indent("", "  ")
	err := encoder.Encode(lConfig.toPlist())
	if err != nil {
		return fmt.Errorf("could not encode plist: %w", err)
	}

	// Check the existing file - if it's the same as the new content, do not overwrite
	user, err := l.user.CurrentUser()
	if err != nil {
		return fmt.Errorf("could not get current user for launchd service: %w", err)
	}

	filename := filepath.Join(user.HomeDir, "Library", "LaunchAgents", fmt.Sprintf("%s.plist", config.Label))
	domainTarget := fmt.Sprintf(domainFormat, user.Uid)
	serviceTarget := fmt.Sprintf("%s/%s", domainTarget, config.Label)

	alreadyLoaded, err := l.isBootstrapped(ctx, serviceTarget)
	if err != nil {
		return err
	}

	fileExists, err := l.fileSystem.FileExists(filename)
	if err != nil {
		return fmt.Errorf("could not determine whether plist '%s' exists: %w", filename, err)
	}

	// If not forcing re-configuration & the service configuration is valid,
	// do nothing
	if !force && alreadyLoaded && fileExists {
		return nil
	}

	// Unload the service so we can reconfigure & reload
	if alreadyLoaded {
		_, err = l.bootout(ctx, serviceTarget)
		if err != nil {
			return fmt.Errorf("could not bootout daemon process '%s': %w", config.Label, err)
		}
	}

	// Rewrite the plist, if needed
	if !fileExists || force {
		// TODO: only overwrite file if file contents have changed
		err = l.fileSystem.WriteFile(filename, newPlist.Bytes())
		if err != nil {
			return fmt.Errorf("unable to write plist file: %w", err)
		}
	}

	err = l.bootstrapFile(ctx, domainTarget, filename)
	if err != nil {
		return fmt.Errorf("could not bootstrap daemon process '%s': %w", config.Label, err)
	}

	return nil
}

func (l *launchd) Start(ctx context.Context, label string) error {
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

func (l *launchd) Stop(ctx context.Context, label string) error {
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
	if exitCode != 0 &&
		exitCode != LaunchdServiceNotFoundErrorCode &&
		exitCode != LaunchdNoSuchProcessErrorCode {
		return fmt.Errorf("'launchctl kill' exited with status %d", exitCode)
	}

	return nil
}

func (l *launchd) Remove(ctx context.Context, label string) error {
	user, err := l.user.CurrentUser()
	if err != nil {
		return fmt.Errorf("could not get current user for launchd service: %w", err)
	}

	filename := filepath.Join(user.HomeDir, "Library", "LaunchAgents", fmt.Sprintf("%s.plist", label))
	domainTarget := fmt.Sprintf(domainFormat, user.Uid)
	serviceTarget := fmt.Sprintf("%s/%s", domainTarget, label)

	_, err = l.bootout(ctx, serviceTarget)
	if err != nil {
		return fmt.Errorf("could not remove daemon process '%s': %w", label, err)
	}

	_, err = l.fileSystem.DeleteFile(filename)
	if err != nil {
		return fmt.Errorf("could not delete launchd plist: %w", err)
	}

	return nil
}
