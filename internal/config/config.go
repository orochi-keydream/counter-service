package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Service  ServiceConfig  `yaml:"service"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Database DatabaseConfig `yaml:"database"`
}

type ServiceConfig struct {
	GrpcPort int `yaml:"grpc_port"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	DatabaseName string `yaml:"dbname"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
}

type KafkaConfig struct {
	Brokers   []string        `yaml:"brokers"`
	Producers ProducerConfigs `yaml:"producers"`
	Consumers ConsumerConfigs `yaml:"consumers"`
}

type ProducerConfigs struct {
	DialogueCommands ProducerConfig `yaml:"dialogue_commands"`
}

type ProducerConfig struct {
	Topic string `yaml:"topic"`
}

type ConsumerConfigs struct {
	CounterCommands ConsumerConfig `yaml:"counter_commands"`
}

type ConsumerConfig struct {
	Topic string `yaml:"topic"`
}

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "", "Specifies the path to the config file.")
	flag.Parse()
}

func LoadConfig() Config {
	if configPath == "" {
		panic("path to a config file not specified")
	}

	var config Config
	err := cleanenv.ReadConfig(configPath, &config)

	if err != nil {
		panic(err)
	}

	return config
}
