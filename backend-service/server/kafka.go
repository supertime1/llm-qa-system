package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Define all Kafka topics as constants
const (
	TopicLLMResponses = "llm-responses"
	// TopicUserMessages  = "user-messages"
	// TopicDoctorReviews = "doctor-reviews"
	// Add other topics as needed
)

// KafkaConfig holds configuration for Kafka
type KafkaConfig struct {
	Brokers []string
	Topics  []string
}

// GetDefaultTopics returns all Kafka topics used by the system
func GetDefaultTopics() []string {
	return []string{
		TopicLLMResponses,
		// TopicUserMessages,
		// TopicDoctorReviews,
	}
}

// InitializeKafka creates required topics if they don't exist
func InitializeKafka(config KafkaConfig) error {
	conn, err := kafka.DialContext(context.Background(), "tcp", config.Brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %v", err)
	}
	defer conn.Close()

	var topicConfigs []kafka.TopicConfig
	for _, topic := range config.Topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Printf("Note: %v (this is usually fine if topics already exist)", err)
		return nil
	}

	log.Printf("Successfully created Kafka topics: %v", config.Topics)
	return nil
}

// NewKafkaWriter creates a new Kafka writer instance
func NewKafkaWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		BatchSize:    1, // For immediate writes
	}
}
