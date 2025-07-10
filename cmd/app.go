package main

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/vithsutra/biometric-project-message-processor/processor"
	"github.com/vithsutra/biometric-project-message-processor/repository"
)

func Start(db *database, mqttConn *mqttConn) {

	dbRepo := repository.NewPostgresRepository(db.conn)

	messageProcessor := processor.NewMessageProcessor(
		mqttConn.client,
		dbRepo,
		20,
		500,
	)

	messageProcessor.Start()

	for {
		if status := mqttConn.client.IsConnected(); !status {
			if token := mqttConn.client.Connect(); token.Wait() && token.Error() != nil {
				log.Println("failed to connnect to mqtt broker, Error:", token.Error().Error())
			}

			if mqttConn.client.IsConnected() {
				//for device publish topic -> vs242s001/connection/message
				//device_id/process/message_type/message
				mqttConn.client.Subscribe("+/process/+/message", 1, func(c mqtt.Client, m mqtt.Message) {
					messageProcessor.Push(m)
				})
			}
		}

		time.Sleep(time.Second * 1)
	}
}
