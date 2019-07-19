package nomad

import (
	nomad "github.com/hashicorp/nomad/api"
)

// ReadDeployment returns the data for a given deployment id.
func (c *DefaultClient) ReadDeployment(ID string) (*nomad.Deployment, error) {
	var deployment *nomad.Deployment
	var err error

	for retries := 0; retries <= c.httpRetryAttempts && err == nil; retries++ {
		deployment, _, err = c.Client.Deployments().Info(ID, nil)
	}

	if err != nil {
		return &nomad.Deployment{}, err
	}

	return deployment, nil
}
