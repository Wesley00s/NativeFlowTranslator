package queue

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   string
}

type RabbitMQProducer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQConsumer(amqpURL, queueName string) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("error to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		err := conn.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error to open channel: %w", err)
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error to declare queue: %w", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return nil, fmt.Errorf("error QoS: %w", err)
	}

	return &RabbitMQConsumer{Conn: conn, Channel: ch, Queue: queueName}, nil
}

func (r *RabbitMQConsumer) StartConsuming() (<-chan amqp.Delivery, error) {
	return r.Channel.Consume(r.Queue, "", false, false, false, false, nil)
}

func NewRabbitMQProducer(url string) (*RabbitMQProducer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQProducer{conn: conn, ch: ch}, nil
}

func (p *RabbitMQProducer) Publish(queueName string, body []byte) error {

	_, err := p.ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = p.ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (r *RabbitMQConsumer) Close() {
	if r.Channel != nil {
		err := r.Channel.Close()
		if err != nil {
			return
		}
	}
	if r.Conn != nil {
		err := r.Conn.Close()
		if err != nil {
			return
		}
	}
	log.Println("🔌 RabbitMQ Closed.")
}

func (p *RabbitMQProducer) Close() {
	if p.ch != nil {
		err := p.ch.Close()
		if err != nil {
			return
		}
	}
	if p.conn != nil {
		err := p.conn.Close()
		if err != nil {
			return
		}
	}
}
