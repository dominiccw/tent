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
}

func NewDefaultClient(addr string) *DefaultClient {
	return &DefaultClient{
		Address: addr,
	}
}
