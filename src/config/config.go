package config

import (
	"fmt"
	logger "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func Open(path string) (Config, error) {

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
	BuildDir string  `yaml:"build_dir"`
	Scripts  string  `yaml:"scripts"`
	Logger   Logger  `yaml:"logger"`
	Connect  Connect `yaml:"connect"`
	Docker   Docker  `yaml:"docker"`
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
	Level string `yaml:"level"`
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

type Docker struct {
	Endpoint    string             `yaml:"endpoint"`
	TlsCertPath string             `yaml:"tls_cert_path"`
	Auth        dockerRegistryAuth `yaml:"auth"`
}

type dockerRegistryAuth struct {
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	Email         string `yaml:"email"`
	ServerAddress string `yaml:"server_address"`
}
