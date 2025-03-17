package main

import "github.com/vithsutra/biometric-project-message-processor/config"

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

	producer := NewKafkaProducer([]string{config.KafkaBroker1Address}, config.KafkaTopic)

	defer producer.CloseConnection()

	Start(db, mqttConn, producer)

}
