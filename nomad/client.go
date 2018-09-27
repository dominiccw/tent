package nomad

import (
	"net/http"
	"time"
)

// Client interface.
type Client interface {
	// Deployment
	ReadDeployment(ID string) (Deployment, error)

	// Evaluation
	ReadEvaluation(ID string) (Evaluation, error)

	// Job
	ParseJob(hcl string) (string, string, error)
	UpdateJob(name string, data string) (UpdateJobResponse, error)
	GetLatestDeployment(name string) (Deployment, error)
	StopJob(ID string, purge bool) error
	ReadJob(ID string) (ReadJobResponse, error)
}

// DefaultClient is the default nomad client.
type DefaultClient struct {
	Address string
}

var nomadClient = &http.Client{
	Timeout: time.Second * 10,
}
