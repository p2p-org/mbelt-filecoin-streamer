package config

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	APIUrl     string
	APIWsUrl   string
	APIToken   string
	KafkaHosts string
	PgUrl      string
}

func NewConfig(envFilePath string) (*Config, error) {
	err := gotenv.Load(envFilePath)
	if err != nil {
		logrus.Warn("no .env file found, trying environment variables from OS")
	}
	viper.SetEnvPrefix("MBELT_FILECOIN_STREAMER")
	viper.AutomaticEnv()

	conf := &Config{
		APIUrl:     viper.GetString("API_URL"),
		APIWsUrl:   viper.GetString("API_WS_URL"),
		APIToken:   viper.GetString("API_TOKEN"),
		KafkaHosts: viper.GetString("KAFKA"), // "localhost:9092",
		PgUrl:      viper.GetString("PG_URL"),
	}

	switch {
	case conf.APIUrl == "":
		return nil, errors.New("configuration filed APIUrl is empty")
	case conf.APIWsUrl == "":
		return nil, errors.New("configuration filed APIWsUrl is empty")
	case conf.APIToken == "":
		return nil, errors.New("configuration filed APIToken is empty")
	case conf.KafkaHosts == "":
		return nil, errors.New("configuration filed KafkaHosts is empty")
	case conf.PgUrl == "":
		return nil, errors.New("configuration filed PgUrl is empty")
	}

	return conf, nil
}
