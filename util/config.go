package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config for app
type Config struct {
	ChatGPT ChatGptConfig `json:"chatgpt" yaml:"chatgpt" mapstructure:"chatgpt"`
}

// ChatGptConfig config structure
type ChatGptConfig struct {
	// DBDriver Database engine, postgres
	DBDriver string `json:"dbDriver" yaml:"dbDriver" mapstructure:"dbDriver"`
	// DBSource Database dsn, postgresql://myuser:mypass@localhost:5432/chatgpt?sslmode=disable
	DBSource string `json:"DBSource" yaml:"DBSource" mapstructure:"DBSource"`

	// ServerAddress the web server listen address
	ServerAddress string `json:"serverAddress,omitempty" yaml:"serverAddress,omitempty" mapstructure:"serverAddress,omitempty"`

	// TokenSymmetricKey Symmetric key for token,  key size must be exactly 32 characters
	TokenSymmetricKey string `json:"tokenSymmetricKey" yaml:"tokenSymmetricKey" mapstructure:"tokenSymmetricKey"`
	// AccessTokenDuration the token validity period, e.g. 15m
	AccessTokenDuration time.Duration `json:"accessTokenDuration" yaml:"accessTokenDuration" mapstructure:"accessTokenDuration"`
	// RefreshTokenDuration
	RefreshTokenDuration time.Duration `json:"refreshTokenDuration" yaml:"refreshTokenDuration" mapstructure:"refreshTokenDuration"`

	// ChatGPTAPIKey your ChatGPT API key, sk-I7BZxx
	ChatGPTAPIKey string `json:"chatGPTAPIKey" yaml:"chatGPTAPIKey" mapstructure:"chatGPTAPIKey"`
	// Model which model are we going to use
	Model     string `json:"model" yaml:"model" mapstructure:"model"`
	MaxTokens int    `json:"maxTokens" yaml:"maxTokens" mapstructure:"maxTokens"`

	// Proxy http proxy, http://127.0.01:8080
	Proxy string `json:"proxy,omitempty" yaml:"proxy,omitempty" mapstructure:"proxy,omitempty"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
