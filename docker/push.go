package docker

import (
	"fmt"
	"os/exec"
	"strings"
)

// PushImage pushes a given docker tag.
func (b *DefaultDocker) PushImage(name string, image string, output bool) error {
	args := []string{"push", image}

	if output {
		fmt.Println(fmt.Sprintf("===> [%s]    Docker Args: %s", name, args))
	}

	cmd := exec.Command("docker", args...)

	out, err := cmd.CombinedOutput()

	if output {
		lines := strings.Split(string(out), "\n")

		for _, line := range lines {
			fmt.Println(fmt.Sprintf("===> [%s]    ", name) + line)
		}
	}

	return err
}
