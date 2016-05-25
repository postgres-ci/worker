package docker

import (
	"github.com/fsouza/go-dockerclient"

	"crypto/md5"
	"crypto/rand"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

type Client interface {
	CreateConteiner(image string, options CreateContainerOptions) (*Container, error)
	ListContainers() ([]Container, error)
}

func Bind(c Config) (*client, error) {

	rnd := make([]byte, 25)
	rand.Read(rnd)

	var (
		client = client{
			auth: docker.AuthConfiguration{
				Username:      c.Auth.Username,
				Password:      c.Auth.Password,
				Email:         c.Auth.Email,
				ServerAddress: c.Auth.ServerAddress,
			},
			binds: c.Binds,
			cache: make(map[string]time.Time),
			hash:  fmt.Sprintf("%x", md5.Sum(rnd)),
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
	auth     docker.AuthConfiguration
	binds    []string
	mutex    sync.RWMutex
	cache    map[string]time.Time
	hash     string
	sequence int
}

func (c *client) CreateConteiner(image string, options CreateContainerOptions) (*Container, error) {

	if err := c.pullImage(image); err != nil {

		return nil, err
	}

	c.mutex.Lock()
	c.sequence++
	c.mutex.Unlock()

	createdContainer, err := c.Client.CreateContainer(docker.CreateContainerOptions{
		Name: fmt.Sprintf("pci-seq-%d-%s", c.sequence, c.hash),
		Config: &docker.Config{
			Image:      image,
			WorkingDir: options.WorkingDir,
			Env:        options.Env,
			Entrypoint: options.Entrypoint,
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

		c.RemoveContainer(createdContainer.ID)

		return nil, err
	}

	container, err := c.InspectContainer(createdContainer.ID)

	if err != nil {

		c.RemoveContainer(createdContainer.ID)

		return nil, err
	}

	return &Container{
		Name:        container.Name,
		IPAddress:   container.NetworkSettings.IPAddress,
		CreatedAt:   container.Created,
		client:      c,
		containerID: createdContainer.ID,
	}, nil
}

func (c *client) pullImage(image string) error {

	c.mutex.RLock()

	if ts, ok := c.cache[image]; ok && ts.Add(30*time.Minute).After(time.Now()) {

		c.mutex.RUnlock()

		return nil
	}

	c.mutex.RUnlock()

	err := c.PullImage(
		docker.PullImageOptions{
			Repository: image,
		}, c.auth,
	)

	if err != nil {

		return err
	}

	c.mutex.Lock()

	c.cache[image] = time.Now()

	c.mutex.Unlock()

	return nil
}

func (c *client) RemoveContainer(containerID string) error {

	c.StopContainer(containerID, 0)

	return c.Client.RemoveContainer(docker.RemoveContainerOptions{
		ID:            containerID,
		RemoveVolumes: true,
		Force:         true,
	})
}

func (c *client) ListContainers() ([]Container, error) {

	apiContainers, err := c.Client.ListContainers(docker.ListContainersOptions{All: true})

	if err != nil {

		return nil, err
	}

	containers := make([]Container, 0, len(apiContainers))

	for _, apiContainer := range apiContainers {

		if container, err := c.InspectContainer(apiContainer.ID); err == nil {

			containers = append(containers, Container{
				Name:        container.Name,
				IPAddress:   container.NetworkSettings.IPAddress,
				CreatedAt:   container.Created,
				client:      c,
				containerID: apiContainer.ID,
			})
		}
	}

	return containers, nil
}
