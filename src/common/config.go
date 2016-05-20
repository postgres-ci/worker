package common

import (
	log "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/docker"
	"gopkg.in/yaml.v2"

	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
)

func ReadConfig(path string) (Config, error) {

	var config Config

	if path == "" {

		config.setFromEnv()

		return config, nil
	}

	if _, err := os.Open(path); err != nil {

		if os.IsNotExist(err) {

			return config, fmt.Errorf("No such configuration file '%s'", path)
		}

		return config, fmt.Errorf("Could not open configuration file '%s': %v", path, err)
	}

	data, err := ioutil.ReadFile(path)

	if err != nil {

		return config, nil
	}

	if err := yaml.Unmarshal(data, &config); err != nil {

		return config, err
	}

	return config, nil
}

type Config struct {
	Assets     string        `yaml:"assets"`
	Connect    Connect       `yaml:"connect"`
	Docker     docker.Config `yaml:"docker"`
	Loglevel   string        `yaml:"loglevel"`
	NumWorkers string        `yaml:"num_workers"`
	WorkingDir string        `yaml:"working_dir"`
}

func (c *Config) setFromEnv() {

	var port uint32 = 5432

	if value, err := strconv.ParseUint(os.Getenv("DB_PORT"), 10, 32); err == nil {

		port = uint32(value)
	}

	c.Assets = os.Getenv("ASSETS")
	c.WorkingDir = os.Getenv("WORKING_DIR")
	c.NumWorkers = os.Getenv("NUM_WORKERS")
	c.Loglevel = os.Getenv("LOG_LEVEL")
	c.Connect = Connect{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_DATABASE"),
	}

	c.Docker = docker.Config{
		Endpoint:    os.Getenv("DOCKER_ENDPOINT"),
		TlsCertPath: os.Getenv("DOCKER_TLS_CERT_PATH"),
		Auth: docker.AuthConfig{
			Username:      os.Getenv("DOCKER_AUTH_USERNAME"),
			Password:      os.Getenv("DOCKER_AUTH_PASSWORD"),
			Email:         os.Getenv("DOCKER_AUTH_EMAIL"),
			ServerAddress: os.Getenv("DOCKER_AUTH_SERVER_ADDRESS"),
		},
	}
}

func (c *Config) LogLevel() log.Level {

	switch c.Loglevel {
	case "info":
		return log.InfoLevel
	case "warning":
		return log.WarnLevel
	}

	return log.ErrorLevel
}

func (c *Config) DebugAddr() string {

	return "127.0.0.1:8080"
}

func (c *Config) GetNumWorkers() int {

	switch c.NumWorkers {
	case "":
	case "auto":

		if cpus := runtime.NumCPU(); cpus > 2 {

			return int(float32(cpus) * 0.7)
		}

	default:

		if num, err := strconv.ParseInt(c.NumWorkers, 10, 0); err == nil {

			return int(num)
		}
	}

	return 1
}

type Connect struct {
	Host     string `yaml:"host"`
	Port     uint32 `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (c *Connect) DSN() string {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Username, c.Password, c.Database)

	if c.Port != 0 {

		dsn += fmt.Sprintf(" port=%d", c.Port)
	}

	return dsn
}
