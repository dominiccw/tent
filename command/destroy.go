package command

import (
	"flag"
	"fmt"
	"strings"

	config "github.com/PM-Connect/tent/config"
	nomad "github.com/PM-Connect/tent/nomad"
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

	if envConfig.NomadURL == "" {
		c.UI.Error(fmt.Sprintf("Unable to find any environment config for environment: %s", environment))
		return 1
	}

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
	c.UI.Output(fmt.Sprintf("===> [%s] Starting destruction process.", name))

	if verbose {
		c.UI.Output(fmt.Sprintf("===> [%s] Loading nomad file: %s.", name, deployment.NomadFile))
	}

	jobName := generateJobName(deployment.ServiceName, c.Config.Name, name)

	existingJob, err := nomadClient.ReadJob(jobName)

	groupSizes := map[string]int{}

	if err == nil {
		for _, group := range existingJob.TaskGroups {
			groupSizes[group.Name] = group.Count
		}
	}

	nomadFile := generateNomadFileName(deployment.NomadFile, jobName)

	nomadFileContents, err := loadNomadFile(nomadFile)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] %s", name, err))
		*errorCount++
		return
	}

	if verbose {
		c.UI.Output(fmt.Sprintf("===> [%s] Parsing nomad file and doing variable replacement: %s.", name, deployment.NomadFile))
	}

	parsedFile, err := parseNomadFile(nomadFileContents, c.Config.Name, name, deployment, groupSizes, environment)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] %s", name, err))
		*errorCount++
		return
	}

	if verbose {
		c.UI.Output(fmt.Sprintf("===> [%s] Converting job file to json for job: %s.", name, c.Config.Name))
	}

	_, id, err := nomadClient.ParseJob(parsedFile)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] %s", name, err))
		*errorCount++
		return
	}

	c.UI.Output(fmt.Sprintf("===> [%s] Stopping job.", name))

	err = nomadClient.StopJob(id, false)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] Error stopping job %s: %s", name, id, err))
		*errorCount++
		return
	}

	c.UI.Info(fmt.Sprintf("===> [%s] Successfully stopped job: %s", name, id))
}
