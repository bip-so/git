package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	ENV_KEY        = "ENV"
	ENV_PRODUCTION = "production"
	ENV_DEV        = "development"

	SECRET_KEY  = "SECRET_KEY"
	DB_NAME     = "DB_NAME"
	DB_USER     = "DB_USER"
	DB_PASSWORD = "DB_PASSWORD"
	DB_HOST     = "DB_HOST"
	DB_PORT     = "DB_PORT"
)

func Init(configName, configPath string) {
	viper.SetConfigType("dotenv")
	viper.SetConfigFile(configName)
	viper.AddConfigPath(configPath)
	//Overrides if available.
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func GetConfigString(key string) string {
	value := viper.GetString(key)
	return value
}

func GetConfigInt(key string) int {
	value := viper.GetInt(key)
	return value
}

func GetConfigBool(key string) bool {
	value := viper.GetBool(key)
	return value
}

func IsDev() bool {
	return GetConfigString(ENV_KEY) == ENV_DEV
}

func GetSecretKey() string {
	return GetConfigString(SECRET_KEY)
}

func GetDBName() string {
	return GetConfigString(DB_NAME)
}

func GetDBUser() string {
	return GetConfigString(DB_USER)
}

func GetDBPassword() string {
	return GetConfigString(DB_PASSWORD)
}

func GetDBHost() string {
	return GetConfigString(DB_HOST)
}

func GetDBPort() int {
	return GetConfigInt(DB_PORT)
}
