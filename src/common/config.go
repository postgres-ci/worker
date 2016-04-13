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
	Logger     Logger        `yaml:"logger"`
	NumWorkers string        `yaml:"num_workers"`
	WorkingDir string        `yaml:"working_dir"`
	Debug      struct {
		Host string `host`
		Port uint16 `port`
	} `yaml:"debug"`
}

func (c *Config) DebugAddr() string {

	return fmt.Sprintf("%s:%d", c.Debug.Host, c.Debug.Port)
}

func (c *Config) GetNumWorkers() int {

	switch c.NumWorkers {
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

type Logger struct {
	Level   string `yaml:"level"`
	Logfile string `yaml:"logfile"`
}

func (l *Logger) LogLevel() log.Level {

	switch l.Level {
	case "info":
		return log.InfoLevel
	case "warning":
		return log.WarnLevel
	}

	return log.ErrorLevel
}
