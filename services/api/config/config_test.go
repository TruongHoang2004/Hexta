package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestValidateAuth(t *testing.T) {
	valid := func() *Config {
		cfg := &Config{}
		cfg.JWT.AccessTokenSecret = "fixture-access-signing-key-32-bytes-minimum"
		cfg.JWT.RefreshTokenSecret = "fixture-refresh-signing-key-32-bytes-minimum"
		return cfg
	}
	for _, production := range []bool{false, true} {
		t.Run(map[bool]string{false: "development", true: "production"}[production], func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				mutate func(*Config)
			}{
				{"missing access", func(c *Config) { c.JWT.AccessTokenSecret = "" }},
				{"missing refresh", func(c *Config) { c.JWT.RefreshTokenSecret = "" }},
				{"short access", func(c *Config) { c.JWT.AccessTokenSecret = "short-private-fixture" }},
				{"short refresh", func(c *Config) { c.JWT.RefreshTokenSecret = "short-private-fixture" }},
				{"unexpanded variable", func(c *Config) { c.JWT.AccessTokenSecret = "${UNEXPANDED_PRIVATE_SIGNING_KEY_VARIABLE}" }},
				{"legacy fallback", func(c *Config) { c.JWT.AccessTokenSecret = "access-token-secret-fallback-for-dev" }},
				{"shared keys", func(c *Config) { c.JWT.RefreshTokenSecret = c.JWT.AccessTokenSecret }},
				{"whitespace", func(c *Config) { c.JWT.RefreshTokenSecret = " " + c.JWT.RefreshTokenSecret }},
			} {
				t.Run(tc.name, func(t *testing.T) {
					cfg := valid()
					cfg.Server.Production = production
					tc.mutate(cfg)
					err := cfg.ValidateAuth()
					if err == nil {
						t.Fatal("expected invalid signing configuration to be rejected")
					}
					for _, secret := range []string{cfg.JWT.AccessTokenSecret, cfg.JWT.RefreshTokenSecret} {
						if secret != "" && strings.Contains(err.Error(), secret) {
							t.Fatal("validation disclosed a signing credential")
						}
					}
				})
			}
			cfg := valid()
			cfg.Server.Production = production
			if err := cfg.ValidateAuth(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAIConfigParsing(t *testing.T) {
	yamlConfig := `
server:
  name: test_api
  port: 8080
ai:
  provider: gemini
  api_key: test-api-key
  model: gemini-2.0-flash
  base_url: https://generativelanguage.googleapis.com
  timeout: 30
  max_tokens: 4096
  temperature: 0.7
`

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(strings.NewReader(yamlConfig))
	assert.NoError(t, err)

	var cfg Config
	err = v.Unmarshal(&cfg)
	assert.NoError(t, err)

	assert.Equal(t, "gemini", cfg.AI.Provider)
	assert.Equal(t, "test-api-key", cfg.AI.ApiKey)
	assert.Equal(t, "gemini-2.0-flash", cfg.AI.Model)
	assert.Equal(t, "https://generativelanguage.googleapis.com", cfg.AI.BaseURL)
	assert.Equal(t, 30, cfg.AI.Timeout)
	assert.Equal(t, 4096, cfg.AI.MaxTokens)
	assert.Equal(t, 0.7, cfg.AI.Temperature)
}
