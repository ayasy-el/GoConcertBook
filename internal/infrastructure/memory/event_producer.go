package memory

import "context"

type EventProducer struct{}

func NewEventProducer() *EventProducer {
	return &EventProducer{}
}

func (p *EventProducer) Publish(_ context.Context, _ string, _ string, _ []byte) error {
	return nil
}
