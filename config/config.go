package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Variables struct {
	DatabaseUrl        string
	MqttBrokerHost     string
	MqttBrokerPort     string
	MqttBrokerUserName string
	MqttBrokerPassword string
	RedisUrl           string
}

func InitConfig() *Variables {
	serverMode := os.Getenv("SERVER_MODE")

	if serverMode != "dev" && serverMode != "prod" {
		log.Fatalln("please set SERVER_MODE to dev or prod")
	}

	if serverMode == "dev" {
		if err := godotenv.Load(); err != nil {
			log.Fatalln("failed to load the .env file, Error: ", err.Error())
		}
	}

	variable := new(Variables)

	dbUrl := os.Getenv("DATABASE_URL")

	if dbUrl == "" {
		log.Fatalln("missing or empty DATABASE_URL env variable")
	}

	mqttBrokerHost := os.Getenv("MQTT_BROKER_HOST")

	if mqttBrokerHost == "" {
		log.Fatalln("missing or empty MQTT_BROKER_HOST env variable")
	}

	mqttBrokerPort := os.Getenv("MQTT_BROKER_PORT")

	if mqttBrokerPort == "" {
		log.Fatalln("missing or empty MQTT_BROKER_PORT env variable")
	}

	mqttBrokerUserName := os.Getenv("MQTT_BROKER_USERNAME")

	if mqttBrokerUserName == "" {
		log.Fatalln("missing or empty MQTT_BROKER_USERNAME env variable")
	}

	mqttBrokerPassword := os.Getenv("MQTT_BROKER_PASSWORD")

	if mqttBrokerPassword == "" {
		log.Fatalln("missing or empty MQTT_BROKER_PASSWORD")
	}

	redisUrl := os.Getenv("REDIS_URL")

	if redisUrl == "" {
		log.Fatalln("missing or empty REDIS_URL")
	}

	variable.DatabaseUrl = dbUrl
	variable.MqttBrokerHost = mqttBrokerHost
	variable.MqttBrokerPort = mqttBrokerPort
	variable.MqttBrokerUserName = mqttBrokerUserName
	variable.MqttBrokerPassword = mqttBrokerPassword
	variable.RedisUrl = redisUrl

	return variable
}
