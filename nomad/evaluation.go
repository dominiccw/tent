package nomad

import (
	nomad "github.com/hashicorp/nomad/api"
)

// ReadEvaluation reads the requested evaluation.
func (c *DefaultClient) ReadEvaluation(ID string) (*nomad.Evaluation, error) {
	eval, _, err := c.Client.Evaluations().Info(ID, nil)

	if err != nil {
		return &nomad.Evaluation{}, err
	}

	return eval, nil
}
