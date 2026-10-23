package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Url string `yaml:"url" mapstructure:"url"`
	} `yaml:"database" mapstructure:"database"`
	Server struct {
		Name        string `yaml:"name" mapstructure:"name"`
		Port        string `yaml:"port" mapstructure:"port"`
		Production  bool   `yaml:"production" mapstructure:"production"`
		Trace       bool   `yaml:"trace" mapstructure:"trace"`
		FrontendURL string `yaml:"frontend_url" mapstructure:"frontend_url"`
	} `yaml:"server" mapstructure:"server"`
	Service struct {
	} `yaml:"service" mapstructure:"service"`
	JWT struct {
		AccessTokenSecret  string `yaml:"access_token_secret" mapstructure:"access_token_secret"`
		AccessTokenExpire  int    `yaml:"access_token_expire" mapstructure:"access_token_expire"`
		RefreshTokenSecret string `yaml:"refresh_token_secret" mapstructure:"refresh_token_secret"`
		RefreshTokenExpire int    `yaml:"refresh_token_expire" mapstructure:"refresh_token_expire"`
	} `yaml:"jwt" mapstructure:"jwt"`
	Redis struct {
		Addr string `yaml:"addr" mapstructure:"addr"`
	} `yaml:"redis" mapstructure:"redis"`
	OAuth struct {
		Google struct {
			ClientID     string `yaml:"client_id" mapstructure:"client_id"`
			ClientSecret string `yaml:"client_secret" mapstructure:"client_secret"`
			RedirectURL  string `yaml:"redirect_url" mapstructure:"redirect_url"`
		} `yaml:"google" mapstructure:"google"`
	} `yaml:"oauth" mapstructure:"oauth"`
	AI struct {
		Provider    string  `yaml:"provider" mapstructure:"provider"`
		ApiKey      string  `yaml:"api_key" mapstructure:"api_key"`
		Model       string  `yaml:"model" mapstructure:"model"`
		BaseURL     string  `yaml:"base_url" mapstructure:"base_url"`
		Timeout     int     `yaml:"timeout" mapstructure:"timeout"`
		MaxTokens   int     `yaml:"max_tokens" mapstructure:"max_tokens"`
		Temperature float64 `yaml:"temperature" mapstructure:"temperature"`
	} `yaml:"ai" mapstructure:"ai"`
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

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}
	if err := cfg.ValidateAuth(); err != nil {
		return err
	}
	AppConfig = cfg
	return nil
}

// ValidateAuth fails closed before an authentication service can issue tokens.
// Errors identify the invalid setting without including credential values.
func (c *Config) ValidateAuth() error {
	if c == nil {
		return fmt.Errorf("authentication configuration is required")
	}
	for _, key := range []struct{ name, value string }{
		{"jwt.access_token_secret", c.JWT.AccessTokenSecret},
		{"jwt.refresh_token_secret", c.JWT.RefreshTokenSecret},
	} {
		if len(key.value) < 32 || strings.TrimSpace(key.value) != key.value || strings.Contains(key.value, "${") || strings.Contains(key.value, "fallback-for-dev") {
			return fmt.Errorf("%s must be an explicitly configured signing key of at least 32 bytes", key.name)
		}
	}
	if c.JWT.AccessTokenSecret == c.JWT.RefreshTokenSecret {
		return fmt.Errorf("access and refresh signing keys must differ")
	}
	return nil
}
