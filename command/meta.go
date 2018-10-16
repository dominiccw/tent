package command

import (
	"github.com/pm-connect/tent/config"
	"github.com/mitchellh/cli"
)

// Meta contains the meta options for functionaly for neraly every command.
type Meta struct {
	Config config.Config
	UI     cli.Ui
}
