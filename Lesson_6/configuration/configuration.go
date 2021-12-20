package configuration

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/namsral/flag"
	"gopkg.in/yaml.v2"
)

// File is not `goimports`-ed with -local github.com/golangci/golangci-lint (goimports)

type Config struct {
	StartURL string `json:"startUrl" yaml:"startUrl"`
}

func Load(configFile string) (config *Config, err error) {
	config = NewConfig()
	switch filepath.Ext(configFile) {
	case ".json":
		if err = LoadJSONConfig(&configFile, config); err != nil { // ST1003: func LoadJsonConfig should be LoadJSONConfig (stylecheck)
			return
		}
	case ".yaml":
		if err = LoadYamlConfig(&configFile, config); err != nil {
			return
		}
	case ".env":
		if err = LoadEnvConfig(&configFile, config); err != nil {
			return
		}
	default:
		return nil, fmt.Errorf("invalid format of configuration file")
	}

	if len(os.Args) > 1 && filepath.Ext(os.Args[0]) != ".test" {
		if err = LoadFromArgs(config); err != nil {
			return
		}
	}

	return
}

func LoadJSONConfig(configFile *string, config *Config) error {
	contents, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Error read config file: %s\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(contents, config)
	if err != nil {
		return fmt.Errorf("invalid json: %s", err) //  ST1005: error strings should not end with punctuation or a newline (stylecheck)
	}

	if !validateURL(&config.StartURL) {
		return fmt.Errorf("необходимо ввести валидный URL")
	}

	return nil
}

func LoadYamlConfig(configFile *string, config *Config) error {
	contents, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Error read config file: %s\n", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(contents, config)
	if err != nil {
		return fmt.Errorf("invalid yaml: %s", err) // ST1005: error strings should not end with punctuation or a newline (stylecheck)
	}

	if !validateURL(&config.StartURL) {
		return fmt.Errorf("необходимо ввести валидный URL")
	}

	return nil
}

func LoadEnvConfig(configFile *string, config *Config) error {
	if err := godotenv.Load(*configFile); err != nil {
		return fmt.Errorf("произошла ошибка парсинга файла окружения")
	}

	config.StartURL = os.Getenv("EXTERNAL_URL")
	if !validateURL(&config.StartURL) {
		return fmt.Errorf("необходимо ввести валидный URL")
	}

	return nil
}

func LoadFromArgs(config *Config) error {
	var (
		externalURL = flag.String("external-url", "", "Внешний URL в полном формате")
	)

	flag.Parse()

	config.StartURL = *externalURL

	if !validateURL(&config.StartURL) {
		return fmt.Errorf("необходимо ввести валидный URL")
	}

	return nil
}

func isURL(str string) bool { // ST1003: func isUrl should be isURL (stylecheck)
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func validateURL(currentURL *string) bool { // importShadow: shadow of imported package 'url' (gocritic)
	if *currentURL != "" && !isURL(*currentURL) {
		return false
	}

	return true
}

func NewConfig() *Config {
	return &Config{}
}
