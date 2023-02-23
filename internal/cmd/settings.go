package cmd

import (
	"io"

	"github.com/github/git-bundle-server/internal/utils"
)

type settingType int

const (
	stdinKey settingType = iota
	stdoutKey
	stderrKey
	envKey
)

type Setting utils.KeyValue[settingType, any]

func Stdin(stdin io.Reader) Setting {
	return Setting{
		stdinKey,
		stdin,
	}
}

func Stdout(stdout io.Writer) Setting {
	return Setting{
		stdoutKey,
		stdout,
	}
}

func Stderr(stderr io.Writer) Setting {
	return Setting{
		stderrKey,
		stderr,
	}
}

func Env(env []string) Setting {
	return Setting{
		envKey,
		env,
	}
}
