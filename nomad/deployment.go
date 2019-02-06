package nomad

import (
	nomad "github.com/hashicorp/nomad/api"
)

// ReadDeployment returns the data for a given deployment id.
func (c *DefaultClient) ReadDeployment(ID string) (*nomad.Deployment, error) {
	deployment, _, err := c.Client.Deployments().Info(ID, nil)

	if err != nil {
		return &nomad.Deployment{}, err
	}

	return deployment, nil
}
