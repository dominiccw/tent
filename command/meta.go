package command

import (
	"github.com/mitchellh/cli"
	"labs.pmsystem.co.uk/devops/tent/config"
)

// Meta contains the meta options for functionaly for neraly every command.
type Meta struct {
	Config config.Config
	UI     cli.Ui
}
