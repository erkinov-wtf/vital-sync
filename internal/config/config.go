package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	LocalEnv   = "local"
	ReleaseEnv = "release"
)

const DefaultTimeout = 10 * time.Second

type Config struct {
	AppName  string `yaml:"app_name"`
	Env      string `yaml:"env"`
	Timezone string `yaml:"timezone"`
	Internal Internal
}

type Internal struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
	Jwt      Jwt      `yaml:"jwt"`
}

type Server struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Schema   string `yaml:"schema"`
	Password string `yaml:"password"`
	Timezone string // will be set in MustLoad
}

type Jwt struct {
	Audience        string `yaml:"audience"`
	Domain          string `yaml:"domain"`
	Realm           string `yaml:"realm"`
	Secret          string `yaml:"secret"`
	AccessTokenTTL  int    `yaml:"access_token_ttl"`
	RefreshTokenTTL int    `yaml:"refresh_token_ttl"`
}

func MustLoad() *Config {
	const configPath = "config/config.yml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config file: %s", err.Error())
	}

	if env := os.Getenv("ENV"); env != ReleaseEnv {
		log.Println("ENV is not defined as release, local config values will be used")
		cfg.Env = LocalEnv
	} else {
		log.Println("ENV is defined as release, production configs will be used")
		cfg.Env = ReleaseEnv
	}

	envConfigPath := fmt.Sprintf("config/%s.yml", cfg.Env)
	if _, err := os.Stat(envConfigPath); os.IsNotExist(err) {
		log.Fatalf("environment config file does not exist: %s", envConfigPath)
	}

	type envConfig struct {
		LocalConfigs      *Internal `yaml:"local_configs"`
		TestingConfigs    *Internal `yaml:"testing_configs"`
		ProductionConfigs *Internal `yaml:"production_configs"`
	}

	var envCfg envConfig
	if err := cleanenv.ReadConfig(envConfigPath, &envCfg); err != nil {
		log.Fatalf("cannot read environment config: %s", err.Error())
	}

	switch cfg.Env {
	case ReleaseEnv:
		log.Println("loading production configs...")
		if envCfg.ProductionConfigs != nil {
			cfg.Internal = *envCfg.ProductionConfigs
			updateDbCredentials(&cfg.Internal.Database)
			updateJwtSecret(&cfg.Internal.Jwt)
		} else {
			panic("production configs are not found")
		}

	default:
		log.Println("loading local configs...")
		if envCfg.LocalConfigs != nil {
			cfg.Internal = *envCfg.LocalConfigs
		}
	}

	log.Println("Configurations loaded")
	setTimezone(&cfg)

	return &cfg
}

func updateDbCredentials(db *Database) {
	if host := os.Getenv("DB_HOST"); host != "" {
		db.Host = host
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		db.Name = name
	}
	if user := os.Getenv("DB_USER"); user != "" {
		db.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		db.Password = password
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		db.Port = port
	}
}

func updateJwtSecret(currentSecret *Jwt) {
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		currentSecret.Secret = jwtSecret
	}
}
