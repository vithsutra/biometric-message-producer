package main

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

type mqttConn struct {
	client mqtt.Client
}

func onMqttConnection(client mqtt.Client) {
	log.Println("connected to broker")
}

func onMqttDisconnection(client mqtt.Client, err error) {
	log.Println("disconnected from the mqtt broker, Error: ", err.Error())
}

func NewMqttConnection(brokerHost, brokerPort, userName, password string) *mqttConn {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(fmt.Sprintf("tcp://%v:%v", brokerHost, brokerPort))
	opts.SetClientID(uuid.NewString())
	opts.SetUsername(userName)
	opts.SetPassword(password)
	opts.OnConnect = onMqttConnection
	opts.OnConnectionLost = onMqttDisconnection

	client := mqtt.NewClient(opts)

	return &mqttConn{
		client: client,
	}

}
