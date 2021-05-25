package config

import (
	"fmt"
	"log"
	"sync"

	envConfig "github.com/JeremyLoy/config"
)

// Conf holds config
var Conf *Config
var once sync.Once
var configFile = "config.env"

// Config is the parent config struct
type Config struct {
	DefaultCron string   `config:"DEFAULT_CRON"`
	Notifier    Notifier `config:"NOTIFIER"`
	Scheduler   Scheduler
}

// Notifier holds config for notifiers
type Notifier struct {
	Formspree Formspree `config:"FORMSPREE"`
}

// Formspree holds formspree config
type Formspree struct {
	FormID string `config:"FORM_ID"`
}

// Scheduler holds configs for schedulers
type Scheduler struct {
	DistrictIDs []int `config:"districts"`
}

var defaultConfig = Config{
	DefaultCron: "@every 1m", // Run once every hour beginning of the hour
	Notifier:    Notifier{},
	Scheduler:   Scheduler{},
}

// InitializeConfig initializes config variable
func InitializeConfig() {
	once.Do(setupConfig)
}

func setupConfig() {
	var builder *envConfig.Builder
	c := new(Config)

	*c = defaultConfig

	if configFile != "" {
		fmt.Println("CONFIG FILE NOT EMPTY:", configFile)
		builder = envConfig.From(configFile).FromEnv()
	} else {
		builder = envConfig.FromEnv()
	}
	err := builder.To(c)

	if err != nil {
		log.Fatalf("Failed to read config file. Check if it exists at config.env")
	}

	Conf = c
	fmt.Printf("Config: %#v\n", *Conf)
	fmt.Printf("Configggg: %#v\n", *c)
}
