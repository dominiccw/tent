package command

import (
	"flag"
	"fmt"
	"strings"

	config "labs.pmsystem.co.uk/devops/tent/config"
	nomad "labs.pmsystem.co.uk/devops/tent/nomad"
)

// DestroyCommand runs the build to prepare the project for deployment.
type DestroyCommand struct {
	Meta
}

// Help displays help output for the command.
func (c *DestroyCommand) Help() string {
	helpText := `
Usage: tent destroy [-env=] [-purge] [-force]

	Destroy is used to build the project ready for deployment.
	
	-env=
		Specify the environment configuration to use.
	-purge
		Forces garbage collection of the job within nomad.

General Options:

    ` + generalOptionsUsage() + `
    `

	return strings.TrimSpace(helpText)
}

// Synopsis displays the command synopsis.
func (c *DestroyCommand) Synopsis() string { return "Destroy the project according to the config." }

// Name returns the name of the command.
func (c *DestroyCommand) Name() string { return "build" }

// Run starts the build procedure.
func (c *DestroyCommand) Run(args []string) int {
	var verbose bool
	var environment string
	var purge bool
	var force bool

	flags := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	flags.BoolVar(&verbose, "verbose", false, "Turn on verbose output.")
	flags.BoolVar(&purge, "purge", false, "Purge the job on nomad immediately.")
	flags.BoolVar(&force, "force", false, "Force the descruction and to not ask for confirmation.")
	flags.StringVar(&environment, "env", "production", "Specify the environment to use.")
	flags.Parse(args)

	envConfig := c.Config.Environments[environment]

	args = flags.Args()

	if environment == "production" {
		c.UI.Warn("You are running using the Production environment!")
	}

	if !force {
		result, _ := c.UI.Ask("Are you sure? [Y|n]")

		if result != "Y" && result != "y" {
			return 0
		}
	}

	nomadURL := generateNomadURL(envConfig.NomadURL)

	nomadClient := nomad.DefaultClient{
		Address: nomadURL,
	}

	var concurrency int

	if c.Config.Concurrent {
		concurrency = 5
	} else {
		concurrency = 1
	}

	sem := make(chan bool, concurrency)

	errorCount := 0

	for name, deployment := range c.Config.Deployments {
		sem <- true
		go func(name string, deployment config.Deployment, verbose bool, errorCount *int, nomadClient nomad.Client) {
			defer func() { <-sem }()
			c.destroy(name, deployment, envConfig, purge, verbose, errorCount, nomadClient)
		}(name, deployment, verbose, &errorCount, &nomadClient)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	if errorCount > 0 {
		c.UI.Error("Exiting with errors.")
		return 1
	}

	return 0
}

func (c *DestroyCommand) destroy(name string, deployment config.Deployment, environment config.Environment, purge bool, verbose bool, errorCount *int, nomadClient nomad.Client) {
	jobName := fmt.Sprintf("%s-%s", c.Config.Name, name)

	c.UI.Output(fmt.Sprintf("===> [%s] Stopping job.", name))

	err := nomadClient.StopJob(jobName, false)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] Error stopping job %s: %s", name, jobName, err))
		*errorCount++
		return
	}

	c.UI.Info(fmt.Sprintf("===> [%s] Successfully stopped job: %s", name, jobName))
}
