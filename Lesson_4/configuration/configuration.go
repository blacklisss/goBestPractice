package configuration

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/namsral/flag"
	"gopkg.in/yaml.v2"
	"net/url"
	"os"
	"path/filepath"
)

type Config struct {
	StartUrl string `json:"startUrl" yaml:"startUrl"`
}

func Load(configFile string) (config *Config, err error) {
	config = NewConfig()
	switch filepath.Ext(configFile) {
	case ".json":
		if err = LoadJsonConfig(&configFile, config); err != nil {
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

	if len(os.Args) > 1 && filepath.Ext(os.Args[0]) != ".test" { // тут не нашел лучшего решения исключить тесты из вызова. Буду благодарен, если подскажете варианты
		if err = LoadFromArgs(config); err != nil {
			return
		}
	}

	return
}

func LoadJsonConfig(configFile *string, config *Config) error {
	contents, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Error read config file: %s\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(contents, config)
	if err != nil {
		return fmt.Errorf("invalid json: %s\n", err)
	}

	if !validateUrl(&config.StartUrl) {
		return fmt.Errorf("необходимо ввести валидный Url")
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
		return fmt.Errorf("invalid yaml: %s\n", err)
	}

	if !validateUrl(&config.StartUrl) {
		return fmt.Errorf("необходимо ввести валидный Url")
	}

	return nil
}

func LoadEnvConfig(configFile *string, config *Config) error {
	if err := godotenv.Load(*configFile); err != nil {
		return fmt.Errorf("произошла ошибка парсинга файла окружения")
	}

	config.StartUrl = os.Getenv("EXTERNAL_URL")
	if !validateUrl(&config.StartUrl) {
		return fmt.Errorf("необходимо ввести валидный Url")
	}

	return nil
}

func LoadFromArgs(config *Config) error {
	var (
		externalUrl = flag.String("external-url", "", "Внешний URL в полном формате")
	)

	flag.Parse()

	config.StartUrl = *externalUrl

	if !validateUrl(&config.StartUrl) {
		return fmt.Errorf("необходимо ввести валидный Url")
	}

	return nil
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func validateUrl(url *string) bool {
	if *url != "" && !isUrl(*url) {
		return false
	}

	return true
}

func NewConfig() *Config {
	return &Config{}
}
