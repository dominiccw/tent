package command

import (
	"os"
	"testing"

	filet "github.com/Flaque/filet"
	config "github.com/PM-Connect/tent/config"
	"github.com/PM-Connect/tent/nomad"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
)

type testNomadClient struct {
	// Call Counts
	ReadDeploymentCallCount      int
	ReadEvaluationCallCount      int
	ParseJobCallCount            int
	UpdateJobCallCount           int
	GetLatestDeploymentCallCount int
	StopJobCallCount             int
	ReadJobCallCount             int

	// Return Values
	ReadDeploymentReturnValue      nomad.Deployment
	ReadEvaluationReturnValue      nomad.Evaluation
	ParseJobReturnValue            string
	ParseJobReturnID               string
	UpdateJobReturnValue           nomad.UpdateJobResponse
	GetLatestDeploymentReturnValue nomad.Deployment
	ReadJobReturnValue             nomad.ReadJobResponse
}

func (c *testNomadClient) ReadDeployment(ID string) (nomad.Deployment, error) {
	c.ReadDeploymentCallCount++

	return c.ReadDeploymentReturnValue, nil
}

func (c *testNomadClient) ReadEvaluation(ID string) (nomad.Evaluation, error) {
	c.ReadEvaluationCallCount++

	return c.ReadEvaluationReturnValue, nil
}

func (c *testNomadClient) ParseJob(hcl string) (string, string, error) {
	c.ParseJobCallCount++

	return c.ParseJobReturnValue, c.ParseJobReturnID, nil
}

func (c *testNomadClient) UpdateJob(name string, data string) (nomad.UpdateJobResponse, error) {
	c.UpdateJobCallCount++

	return c.UpdateJobReturnValue, nil
}

func (c *testNomadClient) GetLatestDeployment(name string) (nomad.Deployment, error) {
	c.GetLatestDeploymentCallCount++

	return c.GetLatestDeploymentReturnValue, nil
}

func (c *testNomadClient) StopJob(ID string, purge bool) error {
	c.StopJobCallCount++

	return nil
}

func (c *testNomadClient) ReadJob(ID string) (nomad.ReadJobResponse, error) {
	c.ReadJobCallCount++

	return c.ReadJobReturnValue, nil
}

func TestParseNomadFile(t *testing.T) {
	result, err := parseNomadFile(
		"job \"[!job_name!]\" { group \"[!name!]\" count = [!group_size!] { task \"[!deployment_name!]\" { config { image = \"[!image_web!]\" } } } }",
		"service",
		"deployment",
		config.Deployment{
			Builds: map[string]config.Build{
				"web": config.Build{
					RegistryURL: "some-registry.com",
					Name:        "test",
					DeployTag:   "latest",
				},
			},
			StartInstances: 2,
		},
		map[string]int{},
		config.Environment{},
	)

	assert.Nil(t, err)
	assert.Equal(t, "job \"service-deployment\" { group \"service\" count = 2 { task \"deployment\" { config { image = \"some-registry.com/test:latest\" } } } }", result)
}

func TestParseNomadFileWithoutStartInstances(t *testing.T) {
	result, err := parseNomadFile(
		"job \"[!job_name!]\" { group \"[!name!]\" count = [!group_deployment_size!] { task \"[!deployment_name!]\" { config { image = \"[!image_web!]\" } } } }",
		"service",
		"deployment",
		config.Deployment{
			Builds: map[string]config.Build{
				"web": config.Build{
					RegistryURL: "some-registry.com",
					Name:        "test",
					DeployTag:   "latest",
				},
			},
		},
		map[string]int{},
		config.Environment{},
	)

	assert.Nil(t, err)
	assert.Equal(t, "job \"service-deployment\" { group \"service\" count = 2 { task \"deployment\" { config { image = \"some-registry.com/test:latest\" } } } }", result)
}

func TestParseNomadFileWithGroupSizes(t *testing.T) {
	result, err := parseNomadFile(
		"job \"[!job_name!]\" { group \"[!name!]\" count = [!group_deployment_size!] { task \"[!deployment_name!]\" { config { image = \"[!image_web!]\" } } } }",
		"service",
		"deployment",
		config.Deployment{
			Builds: map[string]config.Build{
				"web": config.Build{
					RegistryURL: "some-registry.com",
					Name:        "test",
					DeployTag:   "latest",
				},
			},
		},
		map[string]int{"deployment": 4},
		config.Environment{},
	)

	assert.Nil(t, err)
	assert.Equal(t, "job \"service-deployment\" { group \"service\" count = 4 { task \"deployment\" { config { image = \"some-registry.com/test:latest\" } } } }", result)
}

func TestLoadNomadFile(t *testing.T) {
	defer filet.CleanUp(t)

	var data = `
    job "test" {}
	`

	filet.File(t, "test.nomad", data)

	result, err := loadNomadFile("test.nomad")

	assert.Nil(t, err)
	assert.Equal(t, data, result)
}

func TestLoadNomadFileWithMissingFile(t *testing.T) {
	result, err := loadNomadFile("my-file-somewhere.nomad")

	assert.NotNil(t, err)
	assert.Empty(t, result)
}

func TestGenerateNomadFileNameWithNoConfiguredFile(t *testing.T) {
	fileName := generateNomadFileName("", "my-job")

	assert.Equal(t, "my-job.nomad", fileName)
}

func TestGenerateNomadFileNameWithConfiguredFile(t *testing.T) {
	fileName := generateNomadFileName("my-file.nomad", "my-job")

	assert.Equal(t, "my-file.nomad", fileName)
}

func TestGenerateNomadURLWithoutTrailingSlash(t *testing.T) {
	url := generateNomadURL("http://example.com")

	assert.Equal(t, "http://example.com", url)
}

func TestGenerateNomadURLWithTrailingSlash(t *testing.T) {
	url := generateNomadURL("http://example.com/")

	assert.Equal(t, "http://example.com", url)
}

func TestGenerateJobNameWithoutServiceName(t *testing.T) {
	name := generateJobName("", "my-app", "web")

	assert.Equal(t, "my-app-web", name)
}

func TestGenerateJobNameWithServiceName(t *testing.T) {
	name := generateJobName("my-service", "my-app", "web")

	assert.Equal(t, "my-service", name)
}

func TestDeploy(t *testing.T) {
	defer filet.CleanUp(t)

	deployCommand := DeployCommand{
		Meta: Meta{
			UI: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
			Config: config.Config{
				Deployments: map[string]config.Deployment{
					"test": config.Deployment{
						NomadFile: "test2.nomad",
					},
				},
			},
		},
	}

	var data = `
    job "test" {
		datacenters = ["dc1"]
		type = "service"

		group "app" {
			count = 2

			task "web" {
				driver = "docker"
			}
		}
	}
	`

	filet.File(t, "test2.nomad", data)

	nomadClient := testNomadClient{
		ReadDeploymentReturnValue: nomad.Deployment{
			Status: "successful",
		},
		ReadEvaluationReturnValue: nomad.Evaluation{
			Status: "complete",
		},
		ParseJobReturnID: "some-job",
		GetLatestDeploymentReturnValue: nomad.Deployment{
			Status: "running",
		},
	}

	var errorCount int

	deployCommand.deploy("test", deployCommand.Meta.Config.Deployments["test"], true, &errorCount, &nomadClient, config.Environment{})

	assert.Equal(t, 0, errorCount)
	assert.Equal(t, 1, nomadClient.ReadDeploymentCallCount)
	assert.Equal(t, 1, nomadClient.ReadEvaluationCallCount)
	assert.Equal(t, 1, nomadClient.ParseJobCallCount)
	assert.Equal(t, 1, nomadClient.UpdateJobCallCount)
	assert.Equal(t, 1, nomadClient.GetLatestDeploymentCallCount)
	assert.Equal(t, 0, nomadClient.StopJobCallCount)
	assert.Equal(t, 1, nomadClient.ReadJobCallCount)
}
