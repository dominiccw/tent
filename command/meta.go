package command

import (
	"github.com/mitchellh/cli"
	"github.com/pm-connect/tent/config"
)

// Meta contains the meta options for functionally for nearly every command.
type Meta struct {
	Config config.Config
	UI     cli.Ui
}
