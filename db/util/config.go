package util

import "github.com/spf13/viper"

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	DBSource      string `mapstructure:"DB_SOURCE"`
	DBDriver      string `mapstructure:"DB_DRIVER"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
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
