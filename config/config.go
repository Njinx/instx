package config

import (
	"embed"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var DEFAULT_CONFIG embed.FS

type Config struct {
	Default_instance string
	Proxy            struct {
		Port int
	}
	Updater struct {
		update_interval int
	}
	Advanced struct {
		Initial_resp_weight          int
		Search_resp_weight           int
		Google_SEARCH_resp_weight    int
		Wikipedia_SEARCH_resp_weight int
		Outlier_multiplier           int
	}
}

var configCache Config

func readFd(fd *os.File) string {
	raw, err := io.ReadAll(fd)
	if err != nil {
		log.Fatalf("Could not read config file: %s\n", err.Error())
	}
	return string(raw[:])
}

func getConfigData() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	DEFAULT_PATH := filepath.Join(user.HomeDir, ".config/searx_space_autoselector.conf")

	envPath := os.Getenv("SEARX_SPACE_AUTOSELECTOR_CONFIG")
	if envPath != "" {
		fd, err := os.OpenFile(envPath, os.O_CREATE, 0644)
		if err != nil {
			log.Printf(
				"Could not read config file \"%s\". Falling back to default path \"%s\"",
				envPath,
				DEFAULT_PATH)
		} else {
			defer fd.Close()
			return readFd(fd)
		}
	}

	fd, err := os.OpenFile(DEFAULT_PATH, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf(
			"Could not read config file \"%s\". Exiting.",
			DEFAULT_PATH)
	} else {
		defer fd.Close()
		return readFd(fd)
	}

	// Should never occur
	return ""
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
