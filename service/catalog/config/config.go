package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Url string `yaml:"url" mapstructure:"url"`
	} `yaml:"database" mapstructure:"database"`
	Server struct {
		Name       string `yaml:"name" mapstructure:"name"`
		Port       string `yaml:"port" mapstructure:"port"`
		Production bool   `yaml:"production" mapstructure:"production"`
	} `yaml:"server" mapstructure:"server"`
	Redis struct {
		Addr string `yaml:"addr" mapstructure:"addr"`
	} `yaml:"redis" mapstructure:"redis"`
	Elasticsearch struct {
		Addresses []string `yaml:"addresses" mapstructure:"addresses"`
		Username  string   `yaml:"username" mapstructure:"username"`
		Password  string   `yaml:"password" mapstructure:"password"`
	} `yaml:"elasticsearch" mapstructure:"elasticsearch"`
}

var AppConfig *Config

func LoadConfig() error {
	godotenv.Load()

	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.SetConfigType("yml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config:", err)
		return err
	}

	// Expand environment variables in config values
	for _, key := range viper.AllKeys() {
		value := viper.GetString(key)
		viper.Set(key, os.ExpandEnv(value))
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return err
	}
	return nil
}
