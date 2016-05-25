package docker

import (
	"github.com/fsouza/go-dockerclient"

	"bytes"
	"fmt"
	"strings"
	"time"
)

type CreateContainerOptions struct {
	Env        []string
	Binds      []string
	WorkingDir string
	Entrypoint []string
}

type Container struct {
	Name        string
	IPAddress   string
	CreatedAt   time.Time
	client      *client
	containerID string
	output      bytes.Buffer
}

func (c *Container) ID() string {

	return c.containerID
}

func (c *Container) Output() string {

	return c.output.String()
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
		OutputStream: &c.output,
		ErrorStream:  &c.output,
	}); err != nil {

		return err
	}

	inspect, err := c.client.InspectExec(exec.ID)

	if err != nil {

		return err
	}

	if inspect.ExitCode != 0 {

		return fmt.Errorf("Execute cmd '%s' failed. Output: %s", cmd, c.Output())
	}

	return nil
}

func (c *Container) Destroy() error {

	return c.client.RemoveContainer(c.containerID)
}
