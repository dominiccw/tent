package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/a8m/envsubst"
	validator "gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

// Environment configuration.
type Environment struct {
	NomadURL  string            `yaml:"nomad_url" validate:"required,url"`
	Variables map[string]string `yaml:"variables"`
}

// Build configuration.
type Build struct {
	Context     string   `yaml:"context"`
	RegistryURL string   `yaml:"registry_url"`
	Name        string   `yaml:"name" validate:"omitempty,min=3"`
	Tags        []string `yaml:"tags"`
	Push        bool     `yaml:"push"`
	Target      string   `yaml:"target" validate:"omitempty,alphanum"`
	File        string   `yaml:"file" validate:"omitempty,file"`
	DeployTag   string   `yaml:"deploy_tag"`
	Script      string   `yaml:"script"`
}

// Deployment Configuration.
type Deployment struct {
	Builds         map[string]Build  `yaml:"builds" validate:"dive"`
	NomadFile      string            `yaml:"nomad_file"`
	StartInstances int               `yaml:"start_instances" validate:"omitempty,min=1,max=10"`
	Variables      map[string]string `yaml:"variables"`
	ServiceName    string            `yaml:"service_name" validate:"omitempty,min=3"`
}

// Config for the overall setup.
type Config struct {
	Name         string                 `yaml:"name" validate:"required,min=3"`
	Concurrent   bool                   `yaml:"concurrent"`
	Environments map[string]Environment `yaml:"environments" validate:"required,dive"`
	Deployments  map[string]Deployment  `yaml:"deployments" validate:"required,dive"`
}

// LoadFromFile generates the config from a given yaml file.
func LoadFromFile(file string) (Config, error) {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		return Config{}, err
	}

	config, err := parseConfig(data)

	return config, err
}

func parseConfig(data []byte) (Config, error) {
	config := Config{}

	err := yaml.Unmarshal(data, &config)

	if err != nil {
		return config, err
	}

	tmpName, _ := envsubst.String(config.Name)
	config.Name = tmpName

	for k, env := range config.Environments {
		var x = env
		tmpNomadURL, _ := envsubst.String(env.NomadURL)
		x.NomadURL = tmpNomadURL

		newVariables := x.Variables

		for key, value := range x.Variables {
			newValue, _ := envsubst.String(value)

			newVariables[key] = newValue
		}

		x.Variables = newVariables

		config.Environments[k] = x
	}

	for k, dep := range config.Deployments {
		var x = dep

		for key, build := range x.Builds {
			var b = build

			tmpRegistryURL, _ := envsubst.String(b.RegistryURL)
			b.RegistryURL = tmpRegistryURL

			tmpBuildName, _ := envsubst.String(b.Name)
			b.Name = strings.ToLower(tmpBuildName)

			tmpTarget, _ := envsubst.String(b.Target)
			b.Target = tmpTarget

			tmpDeployTag, _ := envsubst.String(b.DeployTag)
			b.DeployTag = strings.ToLower(strings.Replace(tmpDeployTag, "/", "-", -1))

			var newTags []string

			for _, tag := range b.Tags {
				tmpTag, _ := envsubst.String(tag)
				newTags = append(newTags, strings.ToLower(strings.Replace(tmpTag, "/", "-", -1)))
			}

			if len(b.Script) > 0 {
				b.Script, _ = filepath.Abs(b.Script)
			}

			b.Tags = newTags

			x.Builds[key] = b
		}

		x.NomadFile, _ = envsubst.String(x.NomadFile)

		if len(x.NomadFile) > 0 {
			x.NomadFile, _ = filepath.Abs(x.NomadFile)
		}

		x.ServiceName, _ = envsubst.String(x.ServiceName)

		newVariables := x.Variables

		for key, value := range x.Variables {
			newValue, _ := envsubst.String(value)

			newVariables[key] = newValue
		}

		x.Variables = newVariables

		config.Deployments[k] = x
	}

	validate := *validator.New()

	err = validate.Struct(config)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			return config, fmt.Errorf("expected field '%s' validation to match '%s' but got value '%s'", err.StructField(), err.Tag(), err.Value())
		}
	}

	for _, dep := range config.Deployments {
		for _, build := range dep.Builds {
			if len(build.Script) == 0 {
				err = validate.Var(build.Name, "required,min=3")

				if err != nil {
					for _, err := range err.(validator.ValidationErrors) {
						return config, fmt.Errorf("expected field '%s' validation to match '%s' but got value '%s'", err.StructField(), err.Tag(), err.Value())
					}
				}

				err = validate.Var(build.DeployTag, "required,min=1")

				if err != nil {
					for _, err := range err.(validator.ValidationErrors) {
						return config, fmt.Errorf("expected field '%s' validation to match '%s' but got value '%s'", err.StructField(), err.Tag(), err.Value())
					}
				}
			}
		}
	}

	return config, err
}
