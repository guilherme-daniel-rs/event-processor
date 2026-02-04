package config

import (
	"reflect"

	"github.com/spf13/viper"
)

var configuration *Config

type Config struct {
	AppName  string         `mapstructure:"APP_NAME"`
	Port     int            `mapstructure:"PORT" default:"8080"`
	AWS      awsConfig      `mapstructure:",squash"`
	SQS      sqsConfig      `mapstructure:",squash"`
	DynamoDB dynamoDBConfig `mapstructure:",squash"`
}

type awsConfig struct {
	Region          string `mapstructure:"AWS_REGION" default:"us-east-1"`
	AccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID" default:"local_access_key"`
	SecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY" default:"local_secret_key"`
	Endpoint        string `mapstructure:"AWS_ENDPOINT" default:"http://localhost:4566"`
}

type sqsConfig struct {
	QueueURL    string `mapstructure:"SQS_QUEUE_URL" default:"http://localhost:4566/000000000000/events-main"`
	MaxMessages int32  `mapstructure:"SQS_MAX_MESSAGES" default:"5"`
	WaitTimeSec int32  `mapstructure:"SQS_WAIT_TIME_SEC" default:"10"`
	MaxRetries  int32  `mapstructure:"SQS_MAX_RETRIES" default:"5"`
}

type dynamoDBConfig struct {
	TableName string `mapstructure:"EVENTS_TABLE" default:"events"`
}

func setDefaultValues(configStruct reflect.Type) {
	for i := 0; i < configStruct.NumField(); i++ {
		field := configStruct.Field(i)
		configName := field.Tag.Get("mapstructure")
		defaultValue := field.Tag.Get("default")

		if configName != "" && defaultValue != "" {
			viper.SetDefault(configName, defaultValue)
		}

		if field.Type.Kind() == reflect.Struct {
			setDefaultValues(field.Type)
		}
	}
}

func Load() error {
	viper.AutomaticEnv()

	configType := reflect.TypeOf(Config{})
	setDefaultValues(configType)

	configuration, _ = GetConfig()

	return nil
}

func Get() *Config {
	return configuration
}

func GetConfig() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
