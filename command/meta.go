package command

import (
	"github.com/mitchellh/cli"
	"github.com/pm-connect/tent/config"
)

// Meta contains the meta options for functionaly for neraly every command.
type Meta struct {
	Config config.Config
	UI     cli.Ui
}
