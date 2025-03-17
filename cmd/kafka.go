package main

import (
	"log"

	"github.com/IBM/sarama"
)

type kafkaProducer struct {
	topic    string
	producer sarama.SyncProducer
}

func NewKafkaProducer(kafkaBrokers []string, topic string) *kafkaProducer {

	admin, err := sarama.NewClusterAdmin(kafkaBrokers, sarama.NewConfig())

	if err != nil {
		log.Fatalln("Error occurred while creating the kafka admin, Error: ", err.Error())
	}

	defer admin.Close()

	topics, err := admin.ListTopics()

	if err != nil {
		log.Fatalln("error occurred while listing the topics, Error: ", err.Error())
	}

	if _, exists := topics[topic]; !exists {
		log.Println("kafka topic ", topic, " not exists")
		log.Println("creating the kafka topic ", topic)

		topicDetail := &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		}

		if err := admin.CreateTopic(topic, topicDetail, false); err != nil {
			log.Fatalln("failed to create the kafka topic ", topic, " Error: ", err.Error())
		}

		log.Println("kafka topic ", topic, " created successfully")
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(kafkaBrokers, config)

	if err != nil {
		log.Fatalln("error occurred while creating the kafka producer, Error: ", err.Error())
	}

	log.Println("kafka producer created succesfully")

	return &kafkaProducer{
		topic,
		producer,
	}
}

func (p *kafkaProducer) CloseConnection() {
	p.producer.Close()
}
