package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Output *Output `mapstructure:"output"`
	Rules  []*Rule `mapstructure:"rules"`
}

type Output struct {
	Method   string `mapstructure:"method"`
	Filename string `mapstructure:"filename"`
}

type Rule struct {
	Name     string     `mapstructure:"name"`
	Query    string     `mapstructure:"query"`
	Regex    string     `mapstructure:"regex"`
	Decoders []*Decoder `mapstructure:"decoders"`
}

type Decoder struct {
	Id     string `mapstructure:"id"`
	Repeat int    `mapstructure:"repeat"`
}

func New() (*Config, error) {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return nil, fmt.Errorf("fatal error config file: %s", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed unmarshalling config: %v", err)
	}

	return &cfg, nil
}
