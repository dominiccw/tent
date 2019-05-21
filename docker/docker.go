package docker

// Docker interface to run docker related commands.
type Docker interface {
	BuildImage(name string, context string, tags []string, buildArgs map[string]string, target string, cacheFrom string, file string, output bool) error
	PushImage(name string, image string, output bool) error
}

// DefaultDocker contains the default setup for docker commands.
type DefaultDocker struct {
}
