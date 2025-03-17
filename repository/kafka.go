package repository

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/vithsutra/biometric-project-message-processor/models"
)

type kafkaRepository struct {
	topic    string
	producer sarama.SyncProducer
}

func NewKafkaRepository(producer sarama.SyncProducer, topic string) *kafkaRepository {
	return &kafkaRepository{
		topic,
		producer,
	}
}

func (repo *kafkaRepository) PublishAttendance(attendance *models.Attendance) error {
	jsonBytes, _ := json.Marshal(attendance)
	msg := sarama.ProducerMessage{
		Topic: repo.topic,
		Value: sarama.ByteEncoder(jsonBytes),
	}
	_, _, err := repo.producer.SendMessage(&msg)
	return err
}
