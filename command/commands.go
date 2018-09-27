package command

import (
	"os"
	"strings"

	"github.com/mitchellh/cli"
	config "labs.pmsystem.co.uk/devops/tent/config"
)

func generalOptionsUsage() string {
	helpText := `
    -verbose
        Enables verbose logging.
    `

	return strings.TrimSpace(helpText)
}

// Commands creates all of the possible commands that can be run.
func Commands(conf config.Config) map[string]cli.CommandFactory {
	meta := Meta{
		Config: conf,
	}

	meta.UI = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	meta.UI = &cli.ColoredUi{
		Ui:         meta.UI,
		ErrorColor: cli.UiColorRed,
		WarnColor:  cli.UiColorYellow,
		InfoColor:  cli.UiColorGreen,
	}

	return map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &BuildCommand{
				Meta: meta,
			}, nil
		},
		"deploy": func() (cli.Command, error) {
			return &DeployCommand{
				Meta: meta,
			}, nil
		},
		"destroy": func() (cli.Command, error) {
			return &DestroyCommand{
				Meta: meta,
			}, nil
		},
	}
}
