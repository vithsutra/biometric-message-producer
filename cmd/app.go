package main

import (
	"log"
	"time"

	"github.com/vithsutra/biometric-project-message-processor/handlers"
	"github.com/vithsutra/biometric-project-message-processor/repository"
)

func Start(db *database, mqttConn *mqttConn, kafkaProducer *kafkaProducer) {

	dbRepo := repository.NewPostgresRepository(db.conn)
	kafkaRepo := repository.NewKafkaRepository(kafkaProducer.producer, kafkaProducer.topic)
	messageHandler := handlers.NewMessageHandler(dbRepo, kafkaRepo)

	for {
		if status := mqttConn.client.IsConnected(); !status {
			if token := mqttConn.client.Connect(); token.Wait() && token.Error() != nil {
				log.Println("failed to connnect to mqtt broker, Error:", token.Error().Error())
			}

			if mqttConn.client.IsConnected() {
				mqttConn.client.Subscribe("+/connection", 1, messageHandler.DeviceConnectionRequestHandler)
				mqttConn.client.Subscribe("+/disconnection", 1, messageHandler.DeviceDisconnectionRequestHandler)
				mqttConn.client.Subscribe("+/deletesync", 1, messageHandler.DeleteSyncRequestHandler)
				mqttConn.client.Subscribe("+/deletesyncack", 1, messageHandler.DeleteSyncAckRequestHandler)
				mqttConn.client.Subscribe("+/insertsync", 1, messageHandler.InsertSyncRequestHandler)
				mqttConn.client.Subscribe("+/insertsyncack", 1, messageHandler.InsertSyncAckRequestHandler)
				mqttConn.client.Subscribe("+/attendance", 1, messageHandler.UpdateAttendanceRequestHandler)
			}
		}
		time.Sleep(time.Second * 1)
	}
}
