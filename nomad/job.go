package nomad

import (
	"errors"
	nomad "github.com/hashicorp/nomad/api"
)

// ParseJob takes a hcl job file and converts it to json.
func (c *DefaultClient) ParseJob(hcl string) (*nomad.Job, error) {
	var job *nomad.Job
	var err error

	for retries := 0; retries <= c.httpRetryAttempts && err == nil; retries++ {
		job, err = c.Client.Jobs().ParseHCL(hcl, false)
	}

	if err != nil || *job.ID == "" {
		return nil, err
	}

	return job, nil
}

// UpdateJob registers the given job with nomad.
func (c *DefaultClient) UpdateJob(job *nomad.Job) (*nomad.JobRegisterResponse, error) {
	var validation *nomad.JobValidateResponse
	var err error

	for retries := 0; retries <= c.httpRetryAttempts && err == nil; retries++ {
		validation, _, err = c.Client.Jobs().Validate(job, nil)
	}

	if err != nil {
		return nil, err
	}

	if len(validation.ValidationErrors) > 0 {
		return nil, errors.New(validation.Error)
	}

	var registered *nomad.JobRegisterResponse

	for retries := 0; retries <= c.httpRetryAttempts && err == nil; retries++ {
		registered, _, err = c.Client.Jobs().Register(job, nil)
	}

	if err != nil {
		return nil, err
	}

	return registered, nil
}

// GetLatestDeployment returns the most recent deployment for a job.
func (c *DefaultClient) GetLatestDeployment(name string) (*nomad.Deployment, error) {
	var deployment *nomad.Deployment
	var err error

	for retries := 0; retries <= c.httpRetryAttempts && err == nil; retries++ {
		deployment, _, err = c.Client.Jobs().LatestDeployment(name, nil)
	}

	if err != nil {
		return nil, err
	}

	return deployment, nil
}

// StopJob stops a given job.
func (c *DefaultClient) StopJob(ID string, purge bool) error {
	var err error

	for retries := 0; retries <= c.httpRetryAttempts && err == nil; retries++ {
		_, _, err = c.Client.Jobs().Deregister(ID, false, nil)
	}

	if err != nil {
		return err
	}

	return nil
}

// ReadJob reads a job by id.
func (c *DefaultClient) ReadJob(ID string) (*nomad.Job, error) {
	var job *nomad.Job
	var err error

	for retries := 0; retries <= c.httpRetryAttempts && err == nil; retries++ {
		job, _, err = c.Client.Jobs().Info(ID, nil)
	}

	if err != nil {
		return nil, err
	}

	return job, nil
}
