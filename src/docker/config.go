package docker

type Config struct {
	Endpoint    string   `yaml:"endpoint"`
	Binds       []string `yaml:"binds"`
	TlsCertPath string   `yaml:"tls_cert_path"`
	Auth        struct {
		Username      string `yaml:"username"`
		Password      string `yaml:"password"`
		Email         string `yaml:"email"`
		ServerAddress string `yaml:"server_address"`
	} `yaml:"auth"`
}
