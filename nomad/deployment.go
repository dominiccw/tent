package nomad

import (
	nomad "github.com/hashicorp/nomad/api"
)

// ReadDeployment returns the data for a given deployment id.
func (c *DefaultClient) ReadDeployment(ID string) (*nomad.Deployment, error) {
	client, err := nomad.NewClient(&nomad.Config{
		Address: c.Address,
	})

	if err != nil {
		return &nomad.Deployment{}, err
	}

	deployment, _, err := client.Deployments().Info(ID, nil)

	if err != nil {
		return &nomad.Deployment{}, err
	}

	return deployment, nil
}
