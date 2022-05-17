package config

import (
	"embed"
	"errors"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed searx_space_autoselector.yaml
var DEFAULT_CONFIG embed.FS

type Config struct {
	DefaultInstance string `yaml:"default_instance"`
	Proxy           struct {
		Port int `yaml:"port"`
	} `yaml:"proxy"`
	Updater struct {
		updateInterval int `yaml:"update_interval"`
	} `yaml:"updater"`
	Advanced struct {
		InitialRespWeight         int `yaml:"initial_resp_weight"`
		SearchRespWeight          int `yaml:"search_resp_weight"`
		GoogleSearchRespWeight    int `yaml:"google_search_resp_weight"`
		WikipediaSearchRespWeight int `yaml:"wikipedia_search_resp_weight"`
		OutlierMultiplier         int `yaml:"outlier_multipler"`
	} `yaml:"advanced"`
}

var configCache Config

func createDefaultConfig(path string) (*os.File, error) {
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	configData, err := DEFAULT_CONFIG.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	_, err = fd.Write(configData)
	if err != nil {
		return nil, err
	}

	return fd, nil
}

func getConfigDataFromPath(path string) ([]byte, error) {
	var fd *os.File
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		fd, err = createDefaultConfig(path)
		if err != nil {
			return nil, err
		}
	} else {
		fd, err = os.OpenFile(path, os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
	}

	data, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getConfigData() []byte {

	// Try environment variable
	envPath := os.Getenv("SEARX_SPACE_AUTOSELECTOR_CONFIG")
	if envPath != "" {
		data, err := getConfigDataFromPath(envPath)
		if err != nil {
			log.Fatalf(
				"Could not read config file at \"%s\": %s",
				envPath, err.Error())
		}
		return data
	}

	// Try hardcoded path
	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	DEFAULT_PATH := filepath.Join(user.HomeDir, ".config/searx_space_autoselector.yaml")
	data, err := getConfigDataFromPath(envPath)
	if err != nil {
		log.Fatalf(
			"Could not read config file at \"%s\": %s",
			DEFAULT_PATH, err.Error())
	}
	return data
}

func ParseConfig() *Config {
	if configCache != (Config{}) {
		return &configCache
	}

	conf := Config{}
	err := yaml.Unmarshal([]byte(getConfigData()), &conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	configCache = conf
	return &conf
}
