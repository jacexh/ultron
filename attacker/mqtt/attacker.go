package mqtt

import (
	"context"
	"errors"
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

	MQTTAttacker interface {
		Topic() string
		ultron.Attacker
	}
)

type (
	MQTTAttackerOption func(MQTTAttacker)
)

var (
	_ MQTTAttacker = (*MQTTPublisher)(nil)
	_ MQTTAttacker = (*MQTTSubscriber)(nil)
)

func NewMQTTPublisher(clientOpts *mqtt.ClientOptions, opts ...MQTTAttackerOption) *MQTTPublisher {
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

	ctx = ultron.AllocateStorageInContext(ctx)
	defer ultron.ClearStorageInContext(ctx)

	token := pub.client.Publish(pub.topic, pub.qos, pub.retained, pub.prepareFunc(ctx))
	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (pub *MQTTPublisher) Topic() string {
	return pub.topic
}

func (pub *MQTTPublisher) Apply(opts ...MQTTAttackerOption) {
	for _, opt := range opts {
		opt(pub)
	}
}

func NewMQTTSubscriber(clientOpts *mqtt.ClientOptions, opts ...MQTTAttackerOption) *MQTTSubscriber {
	client := mqtt.NewClient(clientOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic("cannot connect to broker")
	}
	sub := &MQTTSubscriber{client: client}
	sub.Apply(opts...)
	return sub
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

	ctx = ultron.AllocateStorageInContext(ctx)
	defer ultron.ClearStorageInContext(ctx)

	token := sub.client.Subscribe(sub.topic, sub.qos, sub.handler)
	select {
	case <-token.Done():
		if err := token.Error(); err != nil {
			return err
		}

		// issue: https://github.com/eclipse/paho.mqtt.golang/issues/380
		// return codes in SUBACK: https://www.emqx.com/en/blog/mqtt5-new-features-reason-code-and-ack
		if t, ok := token.(*mqtt.SubscribeToken); ok {
			for _, v := range t.Result() {
				if v == 0x80 {
					return errors.New("failed to subscribe topic")
				}
			}
		}
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

func (sub *MQTTSubscriber) Topic() string {
	return sub.topic
}

func (sub *MQTTSubscriber) Apply(opts ...MQTTAttackerOption) {
	for _, opt := range opts {
		opt(sub)
	}
}

func WithQOS(qos byte) MQTTAttackerOption {
	return func(i MQTTAttacker) {
		switch attacker := i.(type) {
		case *MQTTPublisher:
			attacker.qos = qos
		case *MQTTSubscriber:
			attacker.qos = qos
		}
	}
}

func WithPrepareFunc(fn func(context.Context) interface{}) MQTTAttackerOption {
	return func(i MQTTAttacker) {
		if attacker, ok := i.(*MQTTPublisher); ok {
			attacker.prepareFunc = fn
		}
	}
}

func WithMQTTClient(client mqtt.Client) MQTTAttackerOption {
	return func(i MQTTAttacker) {
		switch attacker := i.(type) {
		case *MQTTPublisher:
			attacker.client = client
		case *MQTTSubscriber:
			attacker.client = client
		}
	}
}

func WithTopic(topic string) MQTTAttackerOption {
	return func(i MQTTAttacker) {
		switch attacker := i.(type) {
		case *MQTTPublisher:
			attacker.topic = topic
		case *MQTTSubscriber:
			attacker.topic = topic
		}
	}
}

func WithMessageHandler(handler mqtt.MessageHandler) MQTTAttackerOption {
	return func(m MQTTAttacker) {
		if attacker, ok := m.(*MQTTSubscriber); ok {
			attacker.handler = handler
		}
	}
}

func WithRetained(retained bool) MQTTAttackerOption {
	return func(m MQTTAttacker) {
		if attacker, ok := m.(*MQTTPublisher); ok {
			attacker.retained = retained
		}
	}
}
