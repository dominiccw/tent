package config

import (
	"os"
	"path/filepath"
	"testing"

	filet "github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	var data = `
    name: my-job
    concurrent: true
    environments:
      staging:
        nomad_url: http://example.com
        variables:
          some_variable: test
      production:
        nomad_url: http://example.com/prod
    deployments:
      web:
        builds:
          app:
            context: .
            registry_url: http://example.com
            name: example
            tags:
              - my-tag
              - latest
              - latest/test
            push: true
            target: production
            deploy_tag: latest
        nomad_file: example.nomad
        start_instances: 2
        service_name: my-service
    `

	c, err := parseConfig([]byte(data))

	expectedNomadFilePath, _ := filepath.Abs("example.nomad")

	assert.Nil(t, err)
	assert.Equal(t, "my-job", c.Name)
	assert.True(t, c.Concurrent)
	assert.Equal(t, "http://example.com", c.Environments["staging"].NomadURL)
	assert.Equal(t, "http://example.com/prod", c.Environments["production"].NomadURL)
	assert.Equal(t, expectedNomadFilePath, c.Deployments["web"].NomadFile)
	assert.Equal(t, ".", c.Deployments["web"].Builds["app"].Context)
	assert.Equal(t, "http://example.com", c.Deployments["web"].Builds["app"].RegistryURL)
	assert.Equal(t, "example", c.Deployments["web"].Builds["app"].Name)
	assert.True(t, c.Deployments["web"].Builds["app"].Push)
	assert.Equal(t, "production", c.Deployments["web"].Builds["app"].Target)
	assert.Equal(t, "latest", c.Deployments["web"].Builds["app"].DeployTag)
	assert.ElementsMatch(t, []string{"my-tag", "latest", "latest-test"}, c.Deployments["web"].Builds["app"].Tags)
	assert.Equal(t, 2, c.Deployments["web"].StartInstances)
	assert.Equal(t, "my-service", c.Deployments["web"].ServiceName)
	assert.Equal(t, "test", c.Environments["staging"].Variables["some_variable"])
}

func TestParseMinimalConfig(t *testing.T) {
	var data = `
    name: test
    environments:
      production:
        nomad_url: http://example.com/prod
    deployments:
      web:
    `

	c, err := parseConfig([]byte(data))

	assert.Nil(t, err)
	assert.False(t, c.Concurrent)
	assert.Equal(t, "http://example.com/prod", c.Environments["production"].NomadURL)
	assert.Empty(t, c.Deployments["web"].NomadFile)
	assert.Empty(t, c.Deployments["web"].Builds)
	assert.Empty(t, c.Deployments["web"].Builds)
	assert.Empty(t, c.Deployments["web"].Builds)
	assert.Empty(t, c.Deployments["web"].Builds)
	assert.Empty(t, c.Deployments["web"].Builds)
}

func TestLoadFromFile(t *testing.T) {
	defer filet.CleanUp(t)

	var data = `
    name: test
    concurrent: true
    environments:
      staging:
        nomad_url: http://example.com
      production:
        nomad_url: http://example.com/prod
    deployments:
      web:
        build:
          app:
            context: .
            registry_url: http://example.com
            name: example
            tags:
              - my-tag
              - latest
            push: true
            deploy_tag: latest
        nomad_file: example.nomad
        start_instances: 2
    `

	filet.File(t, "tent.yaml", data)

	_, err := LoadFromFile("tent.yaml")

	assert.Nil(t, err)
}

func TestParseWithEnvironmentExpansion(t *testing.T) {
	var data = `
    name: test
    concurrent: true
    environments:
      staging:
        nomad_url: ${NOMAD_URL}
        variables:
          some_variable: ${SOME_VARIABLE}
      production:
        nomad_url: http://example.com/prod
    deployments:
      web:
        builds:
          app:
            context: .
            registry_url: ${REG_URL}
            name: ${IMAGE_NAME}
            tags:
              - ${NON_EXISTING:-$DEFAULT_TAG}
              - latest:${COMMIT_HASH}
            push: true
            target: ${NOT_HERE:-production}
            deploy_tag: latest
        nomad_file: ${NOMAD_FILE}
        start_instances: 2
    `

	os.Setenv("DEFAULT_TAG", "my-tag")
	os.Setenv("NOMAD_URL", "http://test-nomad-url")
	os.Setenv("REG_URL", "http://some-registry.com")
	os.Setenv("IMAGE_NAME", "example-image")
	os.Setenv("COMMIT_HASH", "101010999999")
	os.Setenv("DEPLOY_IMAGE", "somewhere/my-image:latest")
	os.Setenv("NOMAD_FILE", "test-nomad-file.nomad")
	os.Setenv("SOME_VARIABLE", "test")

	c, err := parseConfig([]byte(data))

	expectedNomadFilePath, _ := filepath.Abs("test-nomad-file.nomad")

	assert.Nil(t, err)
	assert.Equal(t, "http://test-nomad-url", c.Environments["staging"].NomadURL)
	assert.Equal(t, "http://some-registry.com", c.Deployments["web"].Builds["app"].RegistryURL)
	assert.Equal(t, "example-image", c.Deployments["web"].Builds["app"].Name)
	assert.Equal(t, "production", c.Deployments["web"].Builds["app"].Target)
	assert.ElementsMatch(t, c.Deployments["web"].Builds["app"].Tags, []string{"my-tag", "latest:101010999999"})
	assert.Equal(t, expectedNomadFilePath, c.Deployments["web"].NomadFile)
	assert.Equal(t, "test", c.Environments["staging"].Variables["some_variable"])
}

func TestConfigWithBuildScript(t *testing.T) {
	var data = `
    name: my-job
    concurrent: true
    environments:
      staging:
        nomad_url: http://example.com
      production:
        nomad_url: http://example.com/prod
    deployments:
      web:
        builds:
          app:
            script: ./example.sh
        nomad_file: example.nomad
        start_instances: 2
    `

	c, err := parseConfig([]byte(data))

	expectedFilePath, _ := filepath.Abs("./example.sh")

	assert.Nil(t, err)
	assert.Equal(t, expectedFilePath, c.Deployments["web"].Builds["app"].Script)
}
