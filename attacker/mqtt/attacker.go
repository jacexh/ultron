package mqtt

import (
	"context"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/wosai/ultron/v2"
)

type (
	MQTTPublisher struct {
		topic       string
		qos         byte
		retained    bool
		prepareFunc func(context.Context) interface{}
		client      mqtt.Client
	}

	MQTTSubscriber struct {
		topic   string
		qos     byte
		client  mqtt.Client
		handler mqtt.MessageHandler
	}
)

type (
	MQTTPublisherOption func(*MQTTPublisher)
)

var (
	_ ultron.Attacker = (*MQTTPublisher)(nil)
	_ ultron.Attacker = (*MQTTSubscriber)(nil)
)

func NewMQTTPublisher(clientOpts *mqtt.ClientOptions, opts ...MQTTPublisherOption) *MQTTPublisher {
	client := mqtt.NewClient(clientOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic("cannot connect to broker")
	}
	pub := &MQTTPublisher{client: client}
	pub.Apply(opts...)
	return pub
}

func (pub *MQTTPublisher) Name() string {
	return fmt.Sprintf("%s -> %s", "publisher", pub.topic)
}

func (pub *MQTTPublisher) Fire(ctx context.Context) error {
	if pub.prepareFunc == nil {
		panic("no prepare function provided")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ultron.AllocateStorageInContext(ctx)
	defer ultron.ClearStorageInContext(ctx)

	token := pub.client.Publish(pub.topic, pub.qos, pub.retained, pub.prepareFunc(ctx))
	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (pub *MQTTPublisher) Apply(opts ...MQTTPublisherOption) {
	for _, opt := range opts {
		opt(pub)
	}
}

func WithQOS(qos byte) MQTTPublisherOption {
	return func(pub *MQTTPublisher) {
		pub.qos = qos
	}
}

func WithPrepareFunc(fn func(context.Context) interface{}) MQTTPublisherOption {
	return func(pub *MQTTPublisher) {
		if fn != nil {
			pub.prepareFunc = fn
		}
	}
}

func WithMQTTClient(client mqtt.Client) MQTTPublisherOption {
	return func(m *MQTTPublisher) {
		m.client = client
	}
}

func NewMQTTSubscriber(clientOpts *mqtt.ClientOptions) *MQTTSubscriber {
	client := mqtt.NewClient(clientOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic("cannot connect to broker")
	}
	return &MQTTSubscriber{client: client}
}

func (sub *MQTTSubscriber) Name() string {
	return fmt.Sprintf("%s <- %s", "subscriber", sub.topic)
}

func (sub *MQTTSubscriber) Fire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ultron.AllocateStorageInContext(ctx)
	defer ultron.ClearStorageInContext(ctx)

	token := sub.client.Subscribe(sub.topic, sub.qos, sub.handler)
	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()

	}
}
