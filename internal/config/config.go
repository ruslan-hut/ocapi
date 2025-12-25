package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
)

type Config struct {
	Env    string `yaml:"env" env-default:"local" env-required:"true"`
	Listen struct {
		BindIP string `yaml:"bind_ip" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env-default:"9800"`
		ApiKey string `yaml:"key" env-default:""`
	} `yaml:"listen"`
	SQL struct {
		Enabled  bool   `yaml:"enabled" env-default:"false"`
		Driver   string `yaml:"driver" env-default:"mysql"`
		HostName string `yaml:"hostname" env-default:"localhost"`
		UserName string `yaml:"username" env-default:"root"`
		Password string `yaml:"password" env-default:""`
		Database string `yaml:"database" env-default:""`
		Port     string `yaml:"port" env-default:"8080"`
		Prefix   string `yaml:"prefix" env-default:""`
	} `yaml:"sql"`
	Images struct {
		Path string `yaml:"path" env-default:""`
		Url  string `yaml:"url" env-default:""`
	} `yaml:"images"`
	Product struct {
		CustomFields []string `yaml:"custom_fields"` // additional allowed custom field names
	} `yaml:"product"`
	Telegram struct {
		Enabled bool   `yaml:"enabled" env-default:"false"`
		ApiKey  string `yaml:"api_key" env-default:""`
	} `yaml:"telegram"`
}

var instance *Config
var once sync.Once

func MustLoad(path string) *Config {
	var err error
	once.Do(func() {
		instance = &Config{}
		if err = cleanenv.ReadConfig(path, instance); err != nil {
			desc, _ := cleanenv.GetDescription(instance, nil)
			err = fmt.Errorf("%s; %s", err, desc)
			instance = nil
			log.Fatal(err)
		}
	})
	return instance
}
