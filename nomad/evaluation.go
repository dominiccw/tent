package nomad

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Evaluation details.
type Evaluation struct {
	ID                string
	Type              string
	TriggeredBy       string
	JobID             string
	Status            string
	StatusDescription string
}

// ReadEvaluation reads the requested evaluation.
func (c *DefaultClient) ReadEvaluation(ID string) (Evaluation, error) {
	response, err := nomadClient.Get(fmt.Sprintf("%s/v1/evaluation/%s", c.Address, ID))

	if err != nil {
		return Evaluation{}, err
	}

	buf, _ := ioutil.ReadAll(response.Body)

	var result Evaluation

	json.Unmarshal(buf, &result)

	return result, nil
}
