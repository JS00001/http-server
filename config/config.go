package config

import (
	"os"
	"reflect"
)

type Config struct {
	ENV                   string
	MONGODB_URL           string
	JWT_SECRET            string
	JWT_EXPIRATION        string
	AWS_ACCESS_KEY_ID     string
	AWS_SECRET_ACCESS_KEY string
	AWS_SES_SENDER        string
}

func GetConfig(key string) string {
	var config Config

	prodConfig := Config{
		ENV:         "prod",
		MONGODB_URL: os.Getenv("MONGODB_URL"),

		JWT_SECRET:     os.Getenv("JWT_SECRET"),
		JWT_EXPIRATION: os.Getenv("JWT_EXPIRATION"),

		AWS_SES_SENDER:        os.Getenv("AWS_SES_SENDER"),
		AWS_ACCESS_KEY_ID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWS_SECRET_ACCESS_KEY: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	devConfig := Config{
		ENV:         "dev",
		MONGODB_URL: "mongodb://localhost:27017",

		JWT_SECRET:     "secret",
		JWT_EXPIRATION: "30m",

		AWS_SES_SENDER:        os.Getenv("AWS_SES_SENDER"),
		AWS_ACCESS_KEY_ID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWS_SECRET_ACCESS_KEY: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	if os.Getenv("ENV") == "prod" {
		config = prodConfig
	} else {
		config = devConfig
	}

	value := getField(&config, key)

	return value
}

// Get the value of a struct field using reflection.
func getField(config *Config, key string) string {
	reflectValue := reflect.ValueOf(config)
	v := reflect.Indirect(reflectValue).FieldByName(key)
	return v.String()
}
