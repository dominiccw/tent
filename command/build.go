package command

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	config "github.com/PM-Connect/tent/config"
	"github.com/PM-Connect/tent/docker"
)

// BuildCommand runs the build to prepare the project for deployment.
type BuildCommand struct {
	Meta
}

// Help displays help output for the command.
func (c *BuildCommand) Help() string {
	helpText := `
Usage: tent build

    Build is used to build the project ready for deployment.

General Options:

    ` + generalOptionsUsage() + `
    `

	return strings.TrimSpace(helpText)
}

// Synopsis displays the command synopsis.
func (c *BuildCommand) Synopsis() string { return "Build the project according to the config." }

// Name returns the name of the command.
func (c *BuildCommand) Name() string { return "build" }

// Run starts the build procedure.
func (c *BuildCommand) Run(args []string) int {
	var verbose bool

	flags := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	flags.BoolVar(&verbose, "verbose", false, "Turn on verbose output.")
	flags.Parse(args)

	args = flags.Args()

	var concurrency int

	if c.Config.Concurrent {
		concurrency = 5
	} else {
		concurrency = 1
	}

	c.UI.Output(fmt.Sprintf("===> Running up to %d builds concurrently.", concurrency))

	sem := make(chan bool, concurrency)

	errorCount := 0

	for _, deployment := range c.Config.Deployments {
		for key, build := range deployment.Builds {
			sem <- true
			go func(key string, build config.Build, verbose bool, errorCount *int) {
				defer func() { <-sem }()
				c.build(key, build, verbose, c.makeBuilder(), errorCount)
			}(key, build, verbose, &errorCount)
		}
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

// Create a docker builder to use.
func (c *BuildCommand) makeBuilder() docker.Docker {
	return new(docker.DefaultDocker)
}

// Build the configured image and push to the configured tags.
func (c *BuildCommand) build(name string, build config.Build, verbose bool, builder docker.Docker, errorCount *int) {
	c.UI.Output(fmt.Sprintf("===> [%s] Starting build.", name))

	if len(build.Script) > 0 {
		c.UI.Output(fmt.Sprintf("===> [%s] Running build script: %s", name, build.Script))

		args := []string{build.Script}

		cmd := exec.Command("bash", args...)

		out, err := cmd.CombinedOutput()

		if err != nil {
			c.UI.Error(fmt.Sprintf("===> [%s] Error running script %s: %s", name, build.Script, err))
			*errorCount++
			return
		}

		if verbose {
			lines := strings.Split(string(out), "\n")

			for _, line := range lines {
				fmt.Println(fmt.Sprintf("===> [%s]    ", name) + line)
			}
		}

		c.UI.Info(fmt.Sprintf("===> [%s] Completed build and push process.", name))

		return
	}

	var tagsToBuild []string

	if len(build.Tags) == 0 {
		tagsToBuild = []string{"latest"}
	} else {
		tagsToBuild = build.Tags
	}

	tags := buildTags(build.RegistryURL, build.Name, tagsToBuild)

	err := builder.BuildImage(name, build.Context, tags, build.Target, tags[len(tags)-1], build.File, verbose)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] Failed building image: %s", name, err))
		*errorCount++
		return
	}

	c.UI.Info(fmt.Sprintf("===> [%s] Finished build.", name))

	if build.Push {
		for _, tag := range tags {
			c.UI.Output(fmt.Sprintf("===> [%s] Pushing tag: %s", name, tag))
			err := builder.PushImage(name, tag, verbose)

			if err != nil {
				c.UI.Error(fmt.Sprintf("===> [%s] Failed pushing the tag %s, did you log in? (docker login)", name, tag))
				*errorCount++
			}
		}
	}

	c.UI.Info(fmt.Sprintf("===> [%s] Completed build and push process.", name))
}

// BuildTags combines the list of tags into a list of tags including the repository and the image name.
func buildTags(registryURL string, imageName string, tags []string) []string {
	completeTags := []string{}

	if len(tags) == 0 {
		tags = []string{"latest"}
	}

	for _, tag := range tags {
		completeTags = append(completeTags, BuildTag(registryURL, imageName, tag))
	}

	return completeTags
}

// BuildTag builds a single tag.
func BuildTag(registryURL string, imageName string, tag string) string {
	if len(tag) == 0 {
		tag = "latest"
	}

	if len(registryURL) > 0 && !strings.HasSuffix(registryURL, "/") {
		registryURL = registryURL + "/"
	}

	return fmt.Sprintf("%s%s:%s", registryURL, imageName, tag)
}
