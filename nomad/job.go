package nomad

import (
	"errors"
	nomad "github.com/hashicorp/nomad/api"
)

// ParseJob takes a hcl job file and converts it to json.
func (c *DefaultClient) ParseJob(hcl string) (*nomad.Job, error) {
	client, err := nomad.NewClient(&nomad.Config{
		Address: c.Address,
	})

	if err != nil {
		return nil, err
	}

	job, err := client.Jobs().ParseHCL(hcl, false)

	if err != nil || *job.ID == "" {
		return nil, err
	}

	return job, nil
}

// UpdateJob registers the given job with nomad.
func (c *DefaultClient) UpdateJob(job *nomad.Job) (*nomad.JobRegisterResponse, error) {
	client, err := nomad.NewClient(&nomad.Config{
		Address: c.Address,
	})

	if err != nil {
		return nil, err
	}

	validation, _, err := client.Jobs().Validate(job, nil)

	if err != nil {
		return nil, err
	}

	if len(validation.ValidationErrors) > 0 {
		return nil, errors.New(validation.Error)
	}

	registered, _, err := client.Jobs().Register(job, nil)

	if err != nil {
		return nil, err
	}

	return registered, nil
}

// GetLatestDeployment returns the most recent deployment for a job.
func (c *DefaultClient) GetLatestDeployment(name string) (*nomad.Deployment, error) {
	client, err := nomad.NewClient(&nomad.Config{
		Address: c.Address,
	})

	if err != nil {
		return nil, err
	}

	deployment, _, err := client.Jobs().LatestDeployment(name, nil)

	if err != nil {
		return nil, err
	}

	return deployment, nil
}

// StopJob stops a given job.
func (c *DefaultClient) StopJob(ID string, purge bool) error {
	client, err := nomad.NewClient(&nomad.Config{
		Address: c.Address,
	})

	if err != nil {
		return err
	}

	_, _, err = client.Jobs().Deregister(ID, false, nil)

	if err != nil {
		return err
	}

	return nil
}

// ReadJob reads a job by id.
func (c *DefaultClient) ReadJob(ID string) (*nomad.Job, error) {
	client, err := nomad.NewClient(&nomad.Config{
		Address: c.Address,
	})

	if err != nil {
		return nil, err
	}

	job, _, err := client.Jobs().Info(ID, nil)

	if err != nil {
		return nil, err
	}

	return job, nil
}
