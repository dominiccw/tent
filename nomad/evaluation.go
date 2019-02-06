package nomad

import (
	nomad "github.com/hashicorp/nomad/api"
)

// ReadEvaluation reads the requested evaluation.
func (c *DefaultClient) ReadEvaluation(ID string) (*nomad.Evaluation, error) {
	client, err := nomad.NewClient(&nomad.Config{
		Address: c.Address,
	})

	if err != nil {
		return &nomad.Evaluation{}, err
	}

	eval, _, err := client.Evaluations().Info(ID, nil)

	if err != nil {
		return &nomad.Evaluation{}, err
	}

	return eval, nil
}
