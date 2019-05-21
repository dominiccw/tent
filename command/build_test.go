package command

import (
	"os"
	"testing"

	"github.com/mitchellh/cli"
	config "github.com/pm-connect/tent/config"
	"github.com/pm-connect/tent/docker"
	"github.com/stretchr/testify/assert"
)

func TestBuildTagsWithSingleTag(t *testing.T) {
	tags := buildTags("test.registry.somewhere", "my-repo/my-image", []string{"latest"})

	assert.Equal(t, 1, len(tags))
	assert.ElementsMatch(t, []string{"test.registry.somewhere/my-repo/my-image:latest"}, tags)
}

func TestBuildTagsWithMultipleTags(t *testing.T) {
	tags := buildTags(
		"test.registry.somewhere",
		"my-repo/my-image",
		[]string{"latest", "v1", "master"},
	)

	assert.Equal(t, 3, len(tags))
	assert.ElementsMatch(
		t,
		[]string{"test.registry.somewhere/my-repo/my-image:latest", "test.registry.somewhere/my-repo/my-image:v1", "test.registry.somewhere/my-repo/my-image:master"},
		tags,
	)
}

func TestBuildTagsWithNoTags(t *testing.T) {
	tags := buildTags(
		"test.registry.somewhere",
		"my-repo/my-image",
		[]string{},
	)

	assert.Equal(t, 1, len(tags))
	assert.ElementsMatch(
		t,
		[]string{"test.registry.somewhere/my-repo/my-image:latest"},
		tags,
	)
}

func TestBuildForSingleTag(t *testing.T) {
	buildCommand := BuildCommand{
		Meta: Meta{
			UI: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
			Config: config.Config{
				Deployments: map[string]config.Deployment{
					"test": {
						Builds: map[string]config.Build{
							"app": {
								Context:     ".",
								RegistryURL: "some-registry.somewhere",
								Name:        "my-image",
								Tags:        []string{"latest"},
								Push:        true,
							},
						},
						NomadFile: "test",
					},
				},
			},
		},
	}

	docker := TestDocker{
		BuildImageCallCount: 0,
		PushImageCallCount:  0,
	}

	errorCount := 0

	buildCommand.build(
		"test",
		buildCommand.Meta.Config.Deployments["test"].Builds["app"],
		true,
		&docker,
		&errorCount,
	)

	assert.Equal(t, 1, docker.BuildImageCallCount)
	assert.Equal(t, 1, docker.PushImageCallCount)
	assert.Equal(t, 0, errorCount)
}

func TestBuildForMultipleTags(t *testing.T) {
	buildCommand := BuildCommand{
		Meta: Meta{
			UI: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
			Config: config.Config{
				Deployments: map[string]config.Deployment{
					"test": {
						Builds: map[string]config.Build{
							"app": {
								Context:     ".",
								RegistryURL: "some-registry.somewhere",
								Name:        "my-image",
								Tags:        []string{"latest", "master"},
								Push:        true,
							},
						},
						NomadFile: "test",
					},
				},
			},
		},
	}

	docker := TestDocker{
		BuildImageCallCount: 0,
		PushImageCallCount:  0,
	}

	errorCount := 0

	buildCommand.build(
		"test",
		buildCommand.Meta.Config.Deployments["test"].Builds["app"],
		true,
		&docker,
		&errorCount,
	)

	assert.Equal(t, 1, docker.BuildImageCallCount)
	assert.Equal(t, 2, docker.PushImageCallCount)
	assert.Equal(t, 0, errorCount)
}

func TestBuildForMultipleTagsWithoutPush(t *testing.T) {
	buildCommand := BuildCommand{
		Meta: Meta{
			UI: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
			Config: config.Config{
				Deployments: map[string]config.Deployment{
					"test": {
						Builds: map[string]config.Build{
							"app": {
								Context:     ".",
								RegistryURL: "some-registry.somewhere",
								Name:        "my-image",
								Tags:        []string{"latest", "master"},
								Push:        false,
							},
						},
						NomadFile: "test",
					},
				},
			},
		},
	}

	docker := TestDocker{
		BuildImageCallCount: 0,
		PushImageCallCount:  0,
	}

	errorCount := 0

	buildCommand.build(
		"test",
		buildCommand.Meta.Config.Deployments["test"].Builds["app"],
		true,
		&docker,
		&errorCount,
	)

	assert.Equal(t, 1, docker.BuildImageCallCount)
	assert.Equal(t, 0, docker.PushImageCallCount)
	assert.Equal(t, 0, errorCount)
}

func TestMakeBuilder(t *testing.T) {
	buildCommand := BuildCommand{}

	d := buildCommand.makeBuilder()

	assert.IsType(t, new(docker.DefaultDocker), d)
}

type TestDocker struct {
	BuildImageCallCount int
	PushImageCallCount  int
}

func (b *TestDocker) BuildImage(name string, context string, tags []string, buildArgs map[string]string, target string, cacheFrom string, file string, output bool) error {
	b.BuildImageCallCount++

	return nil
}

func (b *TestDocker) PushImage(name string, image string, output bool) error {
	b.PushImageCallCount++

	return nil
}
