package config

import (
	"fmt"
	"os"
)

func NewProducerBroker() string {
	host := os.Getenv("PRODUCER_HOST")
	port := os.Getenv("PRODUCER_PORT")

	return fmt.Sprintf("%s:%s", host, port)
}

func NewConsumerBroker() string {
	host := os.Getenv("CONSUMER_HOST")
	port := os.Getenv("CONSUMER_PORT")

	return fmt.Sprintf("%s:%s", host, port)
}
