package common

import (
	"fmt"
	logger "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/docker"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func ReadConfig(path string) (Config, error) {

	var config Config

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
	WorkingDir string        `yaml:"working_dir"`
	Assets     string        `yaml:"assets"`
	Logger     Logger        `yaml:"logger"`
	Connect    Connect       `yaml:"connect"`
	Docker     docker.Config `yaml:"docker"`
	Debug      struct {
		Host string `host`
		Port uint16 `port`
	} `yaml:"debug"`
}

func (c *Config) DebugAddr() string {

	return fmt.Sprintf("%s:%d", c.Debug.Host, c.Debug.Port)
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

func (l *Logger) LogLevel() logger.Level {

	switch l.Level {
	case "info":
		return logger.InfoLevel
	case "warning":
		return logger.WarnLevel
	}

	return logger.ErrorLevel
}
