package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vithsutra/biometric-project-message-processor/config"
)

func main() {

	config := config.InitConfig()

	db := NewDatabase(config.DatabaseUrl)

	db.CheckDatabaseConnection()

	defer db.CloseConnection()

	mqttConn := NewMqttConnection(
		config.MqttBrokerHost,
		config.MqttBrokerPort,
		config.MqttBrokerUserName,
		config.MqttBrokerPassword,
	)

	go Start(db, mqttConn)

	//graceful shutdown

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("shutting the service down...")
}
