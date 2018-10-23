package command

import (
	"os"
	"testing"
	"time"

	filet "github.com/Flaque/filet"
	config "github.com/pm-connect/tent/config"
	"github.com/pm-connect/tent/nomad"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockNomadClient struct {
	mock.Mock
}

func (c *mockNomadClient) ReadDeployment(ID string) (nomad.Deployment, error) {
	args := c.Called(ID)
	return args.Get(0).(nomad.Deployment), args.Error(1)
}

func (c *mockNomadClient) ReadEvaluation(ID string) (nomad.Evaluation, error) {
	args := c.Called(ID)
	return args.Get(0).(nomad.Evaluation), args.Error(1)
}

func (c *mockNomadClient) ParseJob(hcl string) (string, string, error) {
	args := c.Called(hcl)
	return args.String(0), args.String(1), args.Error(2)
}

func (c *mockNomadClient) UpdateJob(name string, data string) (nomad.UpdateJobResponse, error) {
	args := c.Called(name, data)
	return args.Get(0).(nomad.UpdateJobResponse), args.Error(1)
}

func (c *mockNomadClient) GetLatestDeployment(name string) (nomad.Deployment, error) {
	args := c.Called(name)
	return args.Get(0).(nomad.Deployment), args.Error(1)
}

func (c *mockNomadClient) StopJob(ID string, purge bool) error {
	args := c.Called(ID, purge)
	return args.Error(0)
}

func (c *mockNomadClient) ReadJob(ID string) (nomad.ReadJobResponse, error) {
	args := c.Called(ID)
	return args.Get(0).(nomad.ReadJobResponse), args.Error(1)
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
				Name: "app",
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

	nomadClient := new(mockNomadClient)

	nomadClient.On("ParseJob", data).Return("{\"job\": {}}", "job-id", nil).Once()
	nomadClient.On("ReadJob", "job-id").Return(nomad.ReadJobResponse{}, nil).Once()
	nomadClient.On("ParseJob", data).Return("{\"job\": {}}", "job-id", nil).Once()
	nomadClient.On("UpdateJob", "job-id", "{\"job\": {}}").Return(nomad.UpdateJobResponse{EvalID: "eval-id"}, nil).Once()
	nomadClient.On("ReadJob", "job-id").Return(nomad.ReadJobResponse{Type: "service"}, nil).Once()
	nomadClient.On("ReadEvaluation", "eval-id").Return(nomad.Evaluation{Status: "pending"}, nil).Twice()
	nomadClient.On("ReadEvaluation", "eval-id").Return(nomad.Evaluation{Status: "complete"}, nil).Once()
	nomadClient.On("GetLatestDeployment", "job-id").Return(nomad.Deployment{ID: "deployment-id", Status: "running"}, nil).Once()
	nomadClient.On("ReadDeployment", "deployment-id").Return(nomad.Deployment{
		ID:     "deployment-id",
		Status: "running",
		TaskGroups: map[string]nomad.DeploymentTaskGroup{
			"web": {
				HealthyAllocs:   2,
				UnhealthyAllocs: 0,
				DesiredTotal:    2,
			},
		},
	}, nil).Once()
	nomadClient.On("ReadDeployment", "deployment-id").Return(nomad.Deployment{ID: "deployment-id", Status: "successful"}, nil).Once()

	var errorCount int

	evaluationNotCompleteSleep = time.Millisecond * 1
	healthyMatchesDesiredSleep = time.Millisecond * 1
	healthyGreaterThanZeroSleep = time.Millisecond * 1
	healthyIsZeroSleep = time.Millisecond * 1

	deployCommand.deploy("test", deployCommand.Meta.Config.Deployments["test"], true, &errorCount, nomadClient, config.Environment{})

	nomadClient.AssertExpectations(t)
	assert.Equal(t, 0, errorCount)
}

func TestDeployForJobWithNoEvaluationReturned(t *testing.T) {
	defer filet.CleanUp(t)

	deployCommand := DeployCommand{
		Meta: Meta{
			UI: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
			Config: config.Config{
				Name: "app",
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
		type = "batch"

		group "app" {
			count = 2

			task "web" {
				driver = "docker"
			}
		}
	}
	`

	filet.File(t, "test2.nomad", data)

	nomadClient := new(mockNomadClient)

	nomadClient.On("ParseJob", data).Return("{\"job\": {}}", "job-id", nil).Once()
	nomadClient.On("ReadJob", "job-id").Return(nomad.ReadJobResponse{}, nil).Once()
	nomadClient.On("ParseJob", data).Return("{\"job\": {}}", "job-id", nil).Once()
	nomadClient.On("UpdateJob", "job-id", "{\"job\": {}}").Return(nomad.UpdateJobResponse{EvalID: ""}, nil).Once()
	nomadClient.On("ReadJob", "job-id").Return(nomad.ReadJobResponse{Type: "batch"}, nil).Once()

	var errorCount int

	deployCommand.deploy("test", deployCommand.Meta.Config.Deployments["test"], true, &errorCount, nomadClient, config.Environment{})

	nomadClient.AssertExpectations(t)
	assert.Equal(t, 0, errorCount)
}

func TestDeployForJobThatFails(t *testing.T) {
	defer filet.CleanUp(t)

	deployCommand := DeployCommand{
		Meta: Meta{
			UI: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
			Config: config.Config{
				Name: "app",
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

	nomadClient := new(mockNomadClient)

	nomadClient.On("ParseJob", data).Return("{\"job\": {}}", "job-id", nil).Once()
	nomadClient.On("ReadJob", "job-id").Return(nomad.ReadJobResponse{}, nil).Once()
	nomadClient.On("ParseJob", data).Return("{\"job\": {}}", "job-id", nil).Once()
	nomadClient.On("UpdateJob", "job-id", "{\"job\": {}}").Return(nomad.UpdateJobResponse{EvalID: "eval-id"}, nil).Once()
	nomadClient.On("ReadJob", "job-id").Return(nomad.ReadJobResponse{Type: "service"}, nil).Once()
	nomadClient.On("ReadEvaluation", "eval-id").Return(nomad.Evaluation{Status: "pending"}, nil).Twice()
	nomadClient.On("ReadEvaluation", "eval-id").Return(nomad.Evaluation{Status: "complete"}, nil).Once()
	nomadClient.On("GetLatestDeployment", "job-id").Return(nomad.Deployment{ID: "deployment-id", Status: "running"}, nil).Once()
	nomadClient.On("ReadDeployment", "deployment-id").Return(nomad.Deployment{
		ID:     "deployment-id",
		Status: "running",
		TaskGroups: map[string]nomad.DeploymentTaskGroup{
			"web": {
				HealthyAllocs:   2,
				UnhealthyAllocs: 0,
				DesiredTotal:    2,
			},
		},
	}, nil).Once()
	nomadClient.On("ReadDeployment", "deployment-id").Return(nomad.Deployment{ID: "deployment-id", Status: "failure"}, nil).Once()

	var errorCount int

	evaluationNotCompleteSleep = time.Millisecond * 1
	healthyMatchesDesiredSleep = time.Millisecond * 1
	healthyGreaterThanZeroSleep = time.Millisecond * 1
	healthyIsZeroSleep = time.Millisecond * 1

	deployCommand.deploy("test", deployCommand.Meta.Config.Deployments["test"], true, &errorCount, nomadClient, config.Environment{})

	nomadClient.AssertExpectations(t)
	assert.Equal(t, 1, errorCount)
}
