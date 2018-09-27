package nomad

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Deployment status.
type Deployment struct {
	ID                string
	JobID             string
	Status            string
	StatusDescription string
	CreateIndex       int
	ModifyIndex       int
	JobCreateIndex    int
	JobModifyIndex    int
	JobVersion        int
	TaskGroups        map[string]DeploymentTaskGroup
}

// DeploymentTaskGroup is the task group for the deployment
type DeploymentTaskGroup struct {
	Promoted        bool
	DesiredCanaries int
	DesiredTotal    int
	PlacedAllocs    int
	HealthyAllocs   int
	UnhealthyAllocs int
}

// ReadDeployment returns the data for a given deployment id.
func (c *DefaultClient) ReadDeployment(ID string) (Deployment, error) {
	response, err := nomadClient.Get(fmt.Sprintf("%s/v1/deployment/%s", c.Address, ID))

	if err != nil {
		return Deployment{}, err
	}

	buf, _ := ioutil.ReadAll(response.Body)

	var result Deployment

	json.Unmarshal(buf, &result)

	return result, nil
}
