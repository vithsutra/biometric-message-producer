package main

import (
	"log"
	"sync"
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
		1000,
	)

	messageProcessor.Start()

	var subMu sync.Mutex
	subscribed := false

	subscribeToTopics := func(client mqtt.Client) {
		subMu.Lock()
		defer subMu.Unlock()
		if subscribed && client.IsConnected() {
			return
		}
		token := client.Subscribe("+/process/+/message", 1, func(c mqtt.Client, m mqtt.Message) {
			messageProcessor.Push(m)
		})
		if token.Wait() && token.Error() != nil {
			log.Println("failed to subscribe to topic, Error:", token.Error().Error())
		} else {
			log.Println("successfully subscribed to +/process/+/message")
			subscribed = true
		}
	}

	// Initial connection
	if token := mqttConn.client.Connect(); token.Wait() && token.Error() != nil {
		log.Println("failed to connect to mqtt broker on startup, will auto-reconnect. Error:", token.Error().Error())
	} else {
		log.Println("connected to mqtt broker")
		subscribeToTopics(mqttConn.client)
	}

	for {
		if !mqttConn.client.IsConnected() {
			subMu.Lock()
			subscribed = false
			subMu.Unlock()

			if token := mqttConn.client.Connect(); token.Wait() && token.Error() != nil {
				log.Println("reconnecting to mqtt broker, Error:", token.Error().Error())
			} else if mqttConn.client.IsConnected() {
				log.Println("reconnected to mqtt broker")
				subscribeToTopics(mqttConn.client)
			}
		}

		time.Sleep(2 * time.Second)
	}
}
