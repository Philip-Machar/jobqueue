package queue

type Queue interface {
	Publish(body []byte) error
	Close() error
}
