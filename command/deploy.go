package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	config "github.com/pm-connect/tent/config"
	nomad "github.com/pm-connect/tent/nomad"
	"github.com/valyala/fasttemplate"
)

var evaluationNotCompleteSleep = time.Millisecond * 500
var healthyMatchesDesiredSleep = time.Millisecond * 500
var healthyGreaterThanZeroSleep = time.Second * 1
var healthyIsZeroSleep = time.Second * 5

// DeployCommand runs the build to prepare the project for deployment.
type DeployCommand struct {
	Meta
}

// Help displays help output for the command.
func (c *DeployCommand) Help() string {
	helpText := `
Usage: tent deploy [-env=]

	Deploy is used to build the project ready for deployment.
	
	-env=
        Specify the environment configuration to use.

General Options:

    ` + generalOptionsUsage() + `
    `

	return strings.TrimSpace(helpText)
}

// Synopsis displays the command synopsis.
func (c *DeployCommand) Synopsis() string { return "Deploy the project according to the config." }

// Name returns the name of the command.
func (c *DeployCommand) Name() string { return "deploy" }

// Run starts the build procedure.
func (c *DeployCommand) Run(args []string) int {
	var verbose bool
	var environment string

	flags := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	flags.BoolVar(&verbose, "verbose", false, "Turn on verbose output.")
	flags.StringVar(&environment, "env", "production", "Specify the environment to use.")
	flags.Parse(args)

	envConfig := c.Config.Environments[environment]

	if envConfig.NomadURL == "" {
		c.UI.Error(fmt.Sprintf("Unable to find any environment config for environment: %s", environment))
		return 1
	}

	args = flags.Args()

	if environment == "production" {
		c.UI.Warn("You are running using the Production environment!")
	}

	nomadURL := generateNomadURL(envConfig.NomadURL)

	nomadClient := nomad.DefaultClient{
		Address: nomadURL,
	}

	var concurrency int

	if c.Config.Concurrent {
		concurrency = 5
	} else {
		concurrency = 1
	}

	sem := make(chan bool, concurrency)

	errorCount := 0

	for name, deployment := range c.Config.Deployments {
		sem <- true
		go func(name string, deployment config.Deployment, verbose bool, errorCount *int, nomadClient nomad.Client, envConfig config.Environment) {
			defer func() { <-sem }()
			c.deploy(name, deployment, verbose, errorCount, nomadClient, envConfig)
		}(name, deployment, verbose, &errorCount, &nomadClient, envConfig)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	if errorCount > 0 {
		c.UI.Error("Exiting with errors.")
		return 1
	}

	return 0
}

func (c *DeployCommand) deploy(name string, deployment config.Deployment, verbose bool, errorCount *int, nomadClient nomad.Client, envConfig config.Environment) {
	c.UI.Output(fmt.Sprintf("===> [%s] Starting deployment.", name))

	if verbose {
		c.UI.Output(fmt.Sprintf("===> [%s] Loading nomad file: %s", name, deployment.NomadFile))
	}

	jobName := generateJobName(deployment.ServiceName, c.Config.Name, name)

	nomadFile := generateNomadFileName(deployment.NomadFile, jobName)

	nomadFileContents, err := loadNomadFile(nomadFile)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] %s", name, err))
		*errorCount++
		return
	}

	if verbose {
		c.UI.Output(fmt.Sprintf("===> [%s] Parsing nomad file and doing variable replacement: %s", name, deployment.NomadFile))
	}

	parsedFile, err := parseNomadFile(nomadFileContents, c.Config.Name, name, deployment, map[string]int{}, envConfig)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] %s", name, err))
		*errorCount++
		return
	}

	jsonOutput, id, err := nomadClient.ParseJob(parsedFile)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] Error building job spec:\n  %s", name, err))
		*errorCount++
		return
	}

	if len(id) == 0 {
		c.UI.Error(fmt.Sprintf("===> [%s] Invalid JobID returned from nomad.", name))
		*errorCount++
		return
	}

	existingJob, err := nomadClient.ReadJob(id)

	groupSizes := map[string]int{}

	if err == nil {
		for _, group := range existingJob.TaskGroups {
			groupSizes[group.Name] = group.Count
		}
	}

	parsedFile, err = parseNomadFile(nomadFileContents, c.Config.Name, name, deployment, groupSizes, envConfig)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] %s", name, err))
		*errorCount++
		return
	}

	if verbose {
		c.UI.Output(fmt.Sprintf("===> [%s] Nomad File: \n %s", name, parsedFile))
	}

	if verbose {
		c.UI.Output(fmt.Sprintf("===> [%s] Converting job file to json for job: %s", name, c.Config.Name))
	}

	jsonOutput, id, err = nomadClient.ParseJob(parsedFile)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] Error building job spec:\n  %s", name, err))
		*errorCount++
		return
	}

	if len(id) == 0 {
		c.UI.Error(fmt.Sprintf("===> [%s] Invalid JobID returned from nomad.", name))
		*errorCount++
		return
	}

	if verbose {
		c.UI.Output(fmt.Sprintf("===> [%s] Nomad File JSON: \n %s", name, jsonOutput))
	}

	c.UI.Output(fmt.Sprintf("===> [%s] Submitting job to nomad.", name))

	result, e := nomadClient.UpdateJob(id, jsonOutput)

	if e != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] Error updating job \"%s\":\n %s", name, c.Config.Name, e))
		*errorCount++
		return
	}

	c.UI.Info(fmt.Sprintf("===> [%s] Job successfully sent to nomad.", name))

	newJob, err := nomadClient.ReadJob(id)

	if err != nil {
		c.UI.Error(fmt.Sprintf("===> [%s] Error fetching created job \"%s\":\n %s", name, c.Config.Name, e))
		*errorCount++
		return
	}

	if result.EvalID == "" && newJob.Type == "batch" {
		return
	} else if result.EvalID == "" {
		out, _ := json.Marshal(newJob)
		c.UI.Error(fmt.Sprintf("===> [%s] Error during job update of type \"%s\". Missing eval ID! \nJob: %s", name, newJob.Type, string(out)))
		*errorCount++
		return
	}

	c.UI.Output(fmt.Sprintf("===> [%s] Monitoring deployment for success.", name))

	eval, _ := nomadClient.ReadEvaluation(result.EvalID)

	for eval.Status != "complete" {
		evalStatus, _ := nomadClient.ReadEvaluation(result.EvalID)
		eval = evalStatus

		if verbose {
			c.UI.Warn(fmt.Sprintf("===> [%s] Evaluation Status: %s", name, eval.Status))
		}

		time.Sleep(evaluationNotCompleteSleep)
	}

	nomadDeployment, err := nomadClient.GetLatestDeployment(id)

	if nomadDeployment.Status == "successful" {
		c.UI.Info(fmt.Sprintf("===> [%s] Deployment successful.", name))
	} else if nomadDeployment.Status == "running" {
		for nomadDeployment.Status == "running" {
			deploymentInfo, err := nomadClient.ReadDeployment(nomadDeployment.ID)

			if err != nil {
				c.UI.Error(fmt.Sprintf("===> [%s] Error monitoring deployment: %s", name, err))
				*errorCount++
				return
			}

			nomadDeployment = deploymentInfo

			var healthy, unhealthy, desired int

			for _, group := range nomadDeployment.TaskGroups {
				healthy += group.HealthyAllocs
				unhealthy += group.UnhealthyAllocs
				desired += group.DesiredTotal
			}

			if verbose {
				if unhealthy > 0 {
					c.UI.Warn(fmt.Sprintf("===> [%s] Deployment is: %s (Healthy: %d, Unhealthy %d, Desired: %d)", name, nomadDeployment.StatusDescription, healthy, unhealthy, desired))
				} else {
					c.UI.Output(fmt.Sprintf("===> [%s] Deployment is: %s (Healthy: %d, Unhealthy %d, Desired: %d)", name, nomadDeployment.StatusDescription, healthy, unhealthy, desired))
				}
			}

			if healthy == desired {
				time.Sleep(healthyMatchesDesiredSleep)
			} else if healthy > 0 {
				time.Sleep(healthyGreaterThanZeroSleep)
			} else {
				time.Sleep(healthyIsZeroSleep)
			}
		}

		if nomadDeployment.Status == "successful" {
			c.UI.Info(fmt.Sprintf("===> [%s] Deployment successful.", name))
		} else {
			*errorCount++
			c.UI.Error(fmt.Sprintf("===> [%s] Deployment unsuccessful. Status: %s", name, nomadDeployment.StatusDescription))
		}
	} else {
		c.UI.Error(fmt.Sprintf("===> [%s] Deployment unsuccessful.", name))
	}
}

func loadNomadFile(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("Unable to find nomad file: %s", path)
	}

	file, err := ioutil.ReadFile(path)

	if err != nil {
		return "", fmt.Errorf("Unable to load nomad file: %s err: %s", path, err)
	}

	return string(file), nil
}

func parseNomadFile(file string, serviceName string, deploymentName string, deployment config.Deployment, groupSizes map[string]int, environment config.Environment) (string, error) {
	template := file

	t := fasttemplate.New(template, "[!", "!]")

	context := map[string]string{
		"name":            serviceName,
		"deployment_name": deploymentName,
		"job_name":        generateJobName(deployment.ServiceName, serviceName, deploymentName),
	}

	for key, build := range deployment.Builds {
		context["image_"+key] = BuildTag(build.RegistryURL, build.Name, build.DeployTag)
	}

	for variable, value := range deployment.Variables {
		context["var_"+variable] = value
	}

	for variable, value := range environment.Variables {
		context["env_"+variable] = value
	}

	for group, size := range groupSizes {
		context["group_"+group+"_size"] = strconv.Itoa(size)
	}

	out := t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		if context[tag] != "" {
			return w.Write([]byte(context[tag]))
		}

		if strings.HasPrefix(tag, "group_") && strings.HasSuffix(tag, "_size") {
			group := strings.TrimSuffix(strings.TrimPrefix(tag, "group_"), "_size")

			if group == "" && groupSizes[deploymentName] != 0 {
				return w.Write([]byte(strconv.Itoa(groupSizes[deploymentName])))
			}

			if groupSizes[group] != 0 {
				return w.Write([]byte(strconv.Itoa(groupSizes[group])))
			}

			if deployment.StartInstances > 0 {
				return w.Write([]byte(strconv.Itoa(deployment.StartInstances)))
			}

			return w.Write([]byte("2"))
		}

		return w.Write([]byte(""))
	})

	return out, nil
}

func generateJobName(serviceName string, tentName string, deploymentName string) string {
	var jobName string

	if len(serviceName) > 0 {
		jobName = serviceName
	} else {
		jobName = fmt.Sprintf("%s-%s", tentName, deploymentName)
	}

	return jobName
}

func generateNomadURL(nomadURL string) string {
	if strings.HasSuffix(nomadURL, "/") {
		nomadURL = nomadURL[:len(nomadURL)-len("/")]
	}

	return nomadURL
}

func generateNomadFileName(nomadFile string, jobName string) string {
	if len(nomadFile) == 0 {
		nomadFile = fmt.Sprintf("%s.nomad", jobName)
	}

	return nomadFile
}
