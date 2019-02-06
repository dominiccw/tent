package nomad

import (
	nomad "github.com/hashicorp/nomad/api"
)

// Client interface.
type Client interface {
	// Deployment
	ReadDeployment(ID string) (*nomad.Deployment, error)

	// Evaluation
	ReadEvaluation(ID string) (*nomad.Evaluation, error)

	// Job
	ParseJob(hcl string) (*nomad.Job, error)
	UpdateJob(*nomad.Job) (*nomad.JobRegisterResponse, error)
	GetLatestDeployment(name string) (*nomad.Deployment, error)
	StopJob(ID string, purge bool) error
	ReadJob(ID string) (*nomad.Job, error)
}

// DefaultClient is the default nomad client.
type DefaultClient struct {
	Address string
	Client  *nomad.Client
}

// NewDefaultClient creates a new client for the given address.
func NewDefaultClient(addr string) (*DefaultClient, error) {
	client, err := nomad.NewClient(&nomad.Config{
		Address: addr,
	})

	if err != nil {
		return nil, err
	}

	return &DefaultClient{
		Address: addr,
		Client:  client,
	}, nil
}
