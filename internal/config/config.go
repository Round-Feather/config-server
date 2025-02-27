package config

type AppConfig struct {
	Server struct {
		Port string `koanf:"port"`
	} `koanf:"server"`
	Repo RepoConfig `koanf:"repo"`
	App  Cfg        `koanf:"app"`
}

type RepoConfig struct {
	Url      string `koanf:"url"`
	Branch   string `koanf:"branch"`
	Account  string `koanf:"account"`
	Password string `koanf:"password"`
}

type Cfg struct {
	Profile string `koanf:"profile"`
}
