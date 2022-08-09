package config

import (
	"strings"
	"sync"

	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	cfg "github.com/mm-dict/pearl-exporter/config"
)

// Config struct holds all of the runtime confgiguration for the application
type Config struct {
	*cfg.BaseConfig
	channels []string
	username string
	password string
}

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

// Init populates the Config struct based on environmental runtime configuration
func Init() Config {

	listenPort := cfg.GetEnv("LISTEN_PORT", "9115")
	os.Setenv("LISTEN_PORT", listenPort)
	ac := cfg.Init()

	appConfig := Config{
		&ac,
		nil,
		"",
		"",
	}

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowInfo())
	logger = log.With(logger, "caller", log.DefaultCaller)

	channels := os.Getenv("CHANNELS")
	if channels != "" {
		appConfig.SetChannels(strings.Split(channels, ", "), logger)
	}
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	if username != "" {
		appConfig.SetUsername(username)
	} else if password != "" {
		appConfig.SetPassword(password)
	}

	return appConfig
}

// Overrides the entire list of repositories
func (c *Config) SetChannels(channels []string, logger log.Logger) {
	c.channels = channels
	c.setScrapeQueries(logger)
}

// SetAPIToken accepts a string oauth2 token for usage in http.request
func (c *Config) SetUsername(username string) {
	c.username = username
}

// SetAPITokenFromFile accepts a file containing an oauth2 token for usage in http.request
func (c *Config) SetPassword(password string) {
	c.password = password
}

// Init populates the Config struct based on environmental runtime configuration
// All URL's are added to the TargetURL's string array
func (c *Config) setScrapeQueries(logger log.Logger) error {

	if len(c.channels) == 0 {
		level.Info(logger).Log("No targets specified. Only rate limit endpoint will be scraped")
	}

	return nil
}
