package cmd

import (
	"io"

	"github.com/github/git-bundle-server/internal/utils"
)

type settingType int

const (
	StdinKey settingType = iota
	StdoutKey
	StderrKey
	EnvKey
)

type Setting utils.KeyValue[settingType, any]

func Stdin(stdin io.Reader) Setting {
	return Setting{
		StdinKey,
		stdin,
	}
}

func Stdout(stdout io.Writer) Setting {
	return Setting{
		StdoutKey,
		stdout,
	}
}

func Stderr(stderr io.Writer) Setting {
	return Setting{
		StderrKey,
		stderr,
	}
}

func Env(env []string) Setting {
	return Setting{
		EnvKey,
		env,
	}
}
