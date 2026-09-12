package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

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
