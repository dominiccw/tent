package nomad

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type parseJobRequest struct {
	Canonicalize bool
	JobHCL       string
}

type parseJobResponse struct {
	ID string
}

// Job is a job.
type Job map[string]interface{}

type updateJobRequest struct {
	Job
}

// ReadJobResponse is the returned job when reading a job.
type ReadJobResponse struct {
	ID         string
	Name       string
	Periodic   Periodic
	TaskGroups []TaskGroup
	Type       string
}

// Periodic is the configuration for a periodic job.
type Periodic struct {
	Enabled         bool
	ProhibitOverlap bool
	Spec            string
	SpecType        string
	TimeZone        string
}

// TaskGroup is a task group within a job.
type TaskGroup struct {
	Name  string
	Count int
}

// UpdateJobResponse contains the state of the job that was created/updated.
type UpdateJobResponse struct {
	EvalID          string
	EvalCreateIndex int
	JobModifyIndex  int
}

// ParseJob takes a hcl job file and converts it to json.
func (c *DefaultClient) ParseJob(hcl string) (string, string, error) {
	requestBody := parseJobRequest{
		Canonicalize: true,
		JobHCL:       hcl,
	}

	requestContent, err := json.Marshal(requestBody)

	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/jobs/parse", c.Address), bytes.NewBuffer(requestContent))
	req.Header.Set("Content-Type", "application/json")

	response, err := nomadClient.Do(req)

	if err != nil {
		return "", "", err
	}

	defer response.Body.Close()

	buf, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", "", err
	}

	var job parseJobResponse

	err = json.Unmarshal(buf, &job)

	if job.ID == "" {
		return "", "", fmt.Errorf("invalid response received from nomad for /v1/jobs/parse \n %s", string(buf))
	}

	return string(buf), job.ID, err
}

// UpdateJob registers the given job with nomad.
func (c *DefaultClient) UpdateJob(name string, data string) (UpdateJobResponse, error) {
	var job Job

	json.Unmarshal([]byte(data), &job)

	requestBody := updateJobRequest{
		Job: job,
	}

	requestContent, err := json.Marshal(requestBody)

	if err != nil {
		return UpdateJobResponse{}, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/job/%s", c.Address, name), bytes.NewBuffer(requestContent))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return UpdateJobResponse{}, err
	}

	response, err := nomadClient.Do(req)

	if err != nil {
		return UpdateJobResponse{}, err
	}

	defer response.Body.Close()

	buf, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return UpdateJobResponse{}, err
	}

	var result UpdateJobResponse

	err = json.Unmarshal(buf, &result)

	return result, err
}

// GetLatestDeployment returns the most recent deployment for a job.
func (c *DefaultClient) GetLatestDeployment(name string) (Deployment, error) {
	response, err := nomadClient.Get(fmt.Sprintf("%s/v1/job/%s/deployment", c.Address, name))

	if err != nil {
		return Deployment{}, err
	}

	buf, _ := ioutil.ReadAll(response.Body)

	var result Deployment

	err = json.Unmarshal(buf, &result)

	return result, err
}

// StopJob stops a given job.
func (c *DefaultClient) StopJob(ID string, purge bool) error {
	var purgeParam string

	if purge {
		purgeParam = "true"
	} else {
		purgeParam = "false"
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/job/%s?purge=%s", c.Address, ID, purgeParam), nil)

	if err != nil {
		return err
	}

	_, reqErr := nomadClient.Do(req)

	return reqErr
}

// ReadJob reads a job by id.
func (c *DefaultClient) ReadJob(ID string) (ReadJobResponse, error) {
	response, err := nomadClient.Get(fmt.Sprintf("%s/v1/job/%s", c.Address, ID))

	if err != nil {
		return ReadJobResponse{}, err
	}

	buf, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return ReadJobResponse{}, err
	}

	var result ReadJobResponse

	err = json.Unmarshal(buf, &result)

	return result, err
}
