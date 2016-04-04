package docker

import (
	"github.com/fsouza/go-dockerclient"

	"bytes"
	"fmt"
	"strings"
)

type CreateContainerOptions struct {
	Env        []string
	Binds      []string
	WorkingDir string
}

type Container struct {
	client      *client
	containerID string
	IPAddress   string
	Output      bytes.Buffer
}

func (c *Container) RunCmd(cmd string) error {

	exec, err := c.client.CreateExec(docker.CreateExecOptions{
		Container:    c.containerID,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          strings.Fields(cmd),
	})

	if err != nil {

		return err
	}

	if err := c.client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: &c.Output,
		ErrorStream:  &c.Output,
	}); err != nil {

		return err
	}

	inspect, err := c.client.InspectExec(exec.ID)

	if err != nil {

		return err
	}

	if inspect.ExitCode != 0 {

		return fmt.Errorf("Execute cmd '%s' failed", cmd)
	}

	return nil
}

func (c *Container) Destroy() error {

	return c.client.removeContainer(c.containerID)
}
