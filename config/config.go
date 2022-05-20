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

//go:embed instx.yaml
var DEFAULT_CONFIG_FS embed.FS

const DEFAULT_CONFIG_FILE = "instx.yaml"

type Config struct {
	DefaultInstance string `yaml:"default_instance"`
	Proxy           struct {
		Port           int    `yaml:"port"`
		PreferencesUrl string `yaml:"preferences_url"`
	} `yaml:"proxy"`
	Updater struct {
		UpdateInterval int64 `yaml:"update_interval"`
		Advanced       struct {
			InitialRespWeight         float64 `yaml:"initial_resp_weight"`
			SearchRespWeight          float64 `yaml:"search_resp_weight"`
			GoogleSearchRespWeight    float64 `yaml:"google_search_resp_weight"`
			WikipediaSearchRespWeight float64 `yaml:"wikipedia_search_resp_weight"`
			OutlierMultiplier         float64 `yaml:"outlier_multiplier"`
		} `yaml:"advanced"`
		Criteria struct {
			MinimumCspGrade   string   `yaml:"minimum_csp_grade"`
			MinimumTlsGrade   string   `yaml:"minimum_tls_grade"`
			AllowedHttpGrades []string `yaml:"allowed_http_grades,flow"`
			AllowAnalytics    bool     `yaml:"allow_analytics"`
			IsOnion           bool     `yaml:"is_onion"`
			RequireDnssec     bool     `yaml:"require_dnssec"`
			SearxngPreference string   `yaml:"searxng_preference"`
		} `yaml:"criteria"`
	} `yaml:"updater"`
}

func createDefaultConfig(path string) (*os.File, error) {
	baseDir := filepath.Dir(path)
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		os.MkdirAll(baseDir, 0755)
	}

	fd, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	configData, err := DEFAULT_CONFIG_FS.ReadFile(DEFAULT_CONFIG_FILE)
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
	fd.Close()
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

	DEFAULT_PATH := filepath.Join(user.HomeDir, ".config/", DEFAULT_CONFIG_FILE)
	data, err := getConfigDataFromPath(DEFAULT_PATH)
	if err != nil {
		log.Fatalf(
			"Could not read config file at \"%s\": %s",
			DEFAULT_PATH, err.Error())
	}
	return data
}

var notFirstRun bool
var configCache Config

func ParseConfig() *Config {
	if notFirstRun {
		return &configCache
	} else {
		notFirstRun = true
	}

	conf := Config{}
	err := yaml.Unmarshal([]byte(getConfigData()), &conf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	errs := conf.validateConfig()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(err.Error())
		}
		os.Exit(1)
	}

	configCache = conf
	return &conf
}
