package docker

import (
	"github.com/fsouza/go-dockerclient"

	"path/filepath"
	"time"
)

type Client interface {
	CreateConteiner(image string, options CreateContainerOptions) (*Container, error)
}

func Bind(c Config) (*client, error) {
	var (
		client = client{
			auth: docker.AuthConfiguration{
				Username:      c.Auth.Username,
				Password:      c.Auth.Password,
				Email:         c.Auth.Email,
				ServerAddress: c.Auth.ServerAddress,
			},
			binds:     c.Binds,
			pullCache: make(map[string]time.Time),
		}
		err error
	)

	if c.TlsCertPath != "" {

		client.Client, err = docker.NewTLSClient(c.Endpoint,
			filepath.Join(c.TlsCertPath, "cert.pem"),
			filepath.Join(c.TlsCertPath, "key.pem"),
			filepath.Join(c.TlsCertPath, "ca.pem"),
		)

	} else {

		client.Client, err = docker.NewClient(c.Endpoint)
	}

	if err != nil {

		return nil, err
	}

	return &client, nil
}

type client struct {
	*docker.Client
	binds     []string
	auth      docker.AuthConfiguration
	pullCache map[string]time.Time
}

func (c *client) CreateConteiner(image string, options CreateContainerOptions) (*Container, error) {

	if err := c.pullImage(image); err != nil {

		return nil, err
	}

	createdContainer, err := c.Client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      image,
			WorkingDir: options.WorkingDir,
			Env:        options.Env,
		},
		HostConfig: &docker.HostConfig{
			Privileged:    false,
			Binds:         append(c.binds, options.Binds...),
			RestartPolicy: docker.NeverRestart(),
		},
	})

	if err != nil {

		return nil, err
	}

	if err := c.StartContainer(createdContainer.ID, nil); err != nil {

		c.removeContainer(createdContainer.ID)

		return nil, err
	}

	container, err := c.InspectContainer(createdContainer.ID)

	if err != nil {

		c.removeContainer(createdContainer.ID)

		return nil, err
	}

	return &Container{
		client:      c,
		IPAddress:   container.NetworkSettings.IPAddress,
		containerID: createdContainer.ID,
	}, nil
}

func (c *client) pullImage(image string) error {

	if cache, found := c.pullCache[image]; found && cache.Add(time.Minute*5).After(time.Now()) {

		return nil
	}

	c.pullCache[image] = time.Now()

	return c.PullImage(
		docker.PullImageOptions{
			Repository: image,
		}, c.auth,
	)
}

func (c *client) removeContainer(containerID string) error {

	c.StopContainer(containerID, 0)

	return c.RemoveContainer(docker.RemoveContainerOptions{
		ID:            containerID,
		RemoveVolumes: true,
		Force:         true,
	})
}
