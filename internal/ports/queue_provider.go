package ports

type MessagePublisher interface {
	Publish(queue string, body []byte) error
}
