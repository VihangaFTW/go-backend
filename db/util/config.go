package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	DBSource string `mapstructure:"DB_SOURCE"`
	DBDriver string `mapstructure:"DB_DRIVER"`

	HTTPServerAddress string `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress string `mapstructure:"GRPC_SERVER_ADDRESS"`

	PasetoHexKey         string        `mapstructure:"PASETO_SYMMETRIC_KEY"` //* 32 bytes hex string
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

// LoadConfig is responsible for loading the configuration from a file or env variable
func LoadConfig(path string) (config Config, err error) {

	// tells Viper where to look for config files
	viper.AddConfigPath(path)
	// set the config file name to "app"
	viper.SetConfigName("app")
	// tells Viper the file format is .env format
	viper.SetConfigType("env")

	// enables automatic reading of environmental variables.
	// By using viper.AutomaticEnv(), Viper will look for environment variables that match the keys
	// defined in the configuration struct (like DB_SOURCE, DB_DRIVER, and SERVER_ADDRESS) and use
	// those values if they are set.
	viper.AutomaticEnv()

	// actually reads the config file (app.env)
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	// maps the loaded values into Config struct
	err = viper.Unmarshal(&config)
	return
}
