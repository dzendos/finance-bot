package config

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const configFile = "data/config.yaml"

type Config struct {
	Token                       string `yaml:"token"`
	CbrServiceUrl               string `yaml:"cbr_service_url"`
	FrequencyCurrencyRateUpdate int    `yaml:"frequency_currency_rate_update"`

	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	TestDB   string `yaml:"test_db"`
	SslMode  string `yaml:"sslmode"`
}

type Service struct {
	Config Config
}

func New() (*Service, error) {
	s := &Service{}

	rawYAML, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot ReadFile")
	}

	err = yaml.Unmarshal(rawYAML, &s.Config)
	if err != nil {
		return nil, errors.Wrap(err, "cannot Unmarshal")
	}

	return s, nil
}

func (s *Service) Token() string {
	return s.Config.Token
}

func (s *Service) GetUrl() string {
	return s.Config.CbrServiceUrl
}

func (s *Service) GetUpdateRate() time.Duration {
	return time.Duration(s.Config.FrequencyCurrencyRateUpdate) * time.Second
}

func (s *Service) GetHost() string {
	return s.Config.Host
}

func (s *Service) GetPort() int {
	return s.Config.Port
}
