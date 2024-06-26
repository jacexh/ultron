package mqtt

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/wosai/ultron/v2"
)

type (
	MQTTPublishers struct {
		name          string
		topicSelector TopicSelector
		qos           byte
		retained      bool
		generator     PayloadGenerator
		pool          MQTTClientPool
	}

	MQTTSubscribers struct {
		name          string
		topicSelector TopicSelector
		qos           byte
		pool          MQTTClientPool
		handler       mqtt.MessageHandler
	}

	MQTTAttacker interface {
		Role() string
		ultron.Attacker
	}

	MQTTClientPool interface {
		Get() (mqtt.Client, error)
		Put(mqtt.Client)
	}

	TopicSelector func(mqtt.Client) string

	PayloadGenerator func(context.Context) interface{}

	singleClient struct {
		client mqtt.Client
	}

	fixedNumberClientPool struct {
		index   uint32
		clients []mqtt.Client
	}

	dynamicClientPool struct {
		opt  mqtt.ClientOptions
		pool sync.Pool
	}
)

type (
	MQTTAttackerOption func(MQTTAttacker)
)

var (
	_ MQTTAttacker = (*MQTTPublishers)(nil)
	_ MQTTAttacker = (*MQTTSubscribers)(nil)
)

func NewMQTTPublishers(name string, opts ...MQTTAttackerOption) *MQTTPublishers {
	pub := &MQTTPublishers{name: name}
	pub.Apply(opts...)
	return pub
}

func (pub *MQTTPublishers) Name() string {
	return fmt.Sprintf("%s: %s", "publisher", pub.name)
}

func (pub *MQTTPublishers) Fire(ctx context.Context) error {
	if pub.generator == nil || pub.topicSelector == nil || pub.pool == nil {
		panic("no PayloadGenerator/topicSelector/ClientPool provided")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ctx = ultron.AllocateStorageInContext(ctx)
	defer ultron.ClearStorageInContext(ctx)

	client, err := pub.pool.Get()
	if err != nil {
		return err
	}
	defer pub.pool.Put(client)

	token := client.Publish(pub.topicSelector(client), pub.qos, pub.retained, pub.generator(ctx))
	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (pub *MQTTPublishers) Role() string {
	return "publisher"
}

func (pub *MQTTPublishers) Apply(opts ...MQTTAttackerOption) {
	for _, opt := range opts {
		opt(pub)
	}
}

func NewMQTTSubscribers(name string, opts ...MQTTAttackerOption) *MQTTSubscribers {
	sub := &MQTTSubscribers{name: name}
	sub.Apply(opts...)
	return sub
}

func (sub *MQTTSubscribers) Name() string {
	return fmt.Sprintf("%s: %s", "subscriber", sub.name)
}

func (sub *MQTTSubscribers) Fire(ctx context.Context) error {
	if sub.topicSelector == nil || sub.pool == nil {
		panic("no TopicSelector/ClientPool provided")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ctx = ultron.AllocateStorageInContext(ctx)
	defer ultron.ClearStorageInContext(ctx)

	client, err := sub.pool.Get()
	if err != nil {
		return err
	}
	defer sub.pool.Put(client)

	token := client.Subscribe(sub.topicSelector(client), sub.qos, sub.handler)
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

func (sub *MQTTSubscribers) Role() string {
	return "subscriber"
}

func (sub *MQTTSubscribers) Apply(opts ...MQTTAttackerOption) {
	for _, opt := range opts {
		opt(sub)
	}
}

func WithQOS(qos byte) MQTTAttackerOption {
	return func(i MQTTAttacker) {
		switch attacker := i.(type) {
		case *MQTTPublishers:
			attacker.qos = qos
		case *MQTTSubscribers:
			attacker.qos = qos
		}
	}
}

func WithPayloadGenerator(fn func(context.Context) interface{}) MQTTAttackerOption {
	return func(i MQTTAttacker) {
		if attacker, ok := i.(*MQTTPublishers); ok {
			attacker.generator = fn
		}
	}
}

func WithMQTTClientPool(p MQTTClientPool) MQTTAttackerOption {
	return func(i MQTTAttacker) {
		if p == nil {
			return
		}
		switch attacker := i.(type) {
		case *MQTTPublishers:
			attacker.pool = p
		case *MQTTSubscribers:
			attacker.pool = p
		}
	}
}

func WithTopicSelector(fn TopicSelector) MQTTAttackerOption {
	return func(i MQTTAttacker) {
		switch attacker := i.(type) {
		case *MQTTPublishers:
			attacker.topicSelector = fn
		case *MQTTSubscribers:
			attacker.topicSelector = fn
		}
	}
}

func WithMessageHandler(handler mqtt.MessageHandler) MQTTAttackerOption {
	return func(m MQTTAttacker) {
		if attacker, ok := m.(*MQTTSubscribers); ok {
			attacker.handler = handler
		}
	}
}

func WithRetained(retained bool) MQTTAttackerOption {
	return func(m MQTTAttacker) {
		if attacker, ok := m.(*MQTTPublishers); ok {
			attacker.retained = retained
		}
	}
}

func SingleTopic(name string) TopicSelector {
	return func(c mqtt.Client) string {
		return name
	}
}

func RoundTripTopics(topics ...string) TopicSelector {
	var index uint32 = 0
	length := len(topics)
	topics = topics[:]

	return func(c mqtt.Client) string {
		v := atomic.AddUint32(&index, 1)
		return topics[(v-1)%uint32(length)]
	}
}

func NewSingleClientPool(opt *mqtt.ClientOptions) MQTTClientPool {
	opt.SetClientID(uuid.NewString())
	client := mqtt.NewClient(opt)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return &singleClient{client: client}
}

func (sc *singleClient) Get() (mqtt.Client, error) {
	return sc.client, nil
}

func (sc *singleClient) Put(client mqtt.Client) {
}

func NewFixedNumberClientPool(num int, opts *mqtt.ClientOptions) MQTTClientPool {
	p := &fixedNumberClientPool{
		clients: make([]mqtt.Client, num),
	}

	for i := 0; i < num; i++ {
		opts.SetClientID(uuid.NewString())
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		p.clients[i] = client
	}
	return p
}

func (fc *fixedNumberClientPool) Get() (mqtt.Client, error) {
	v := atomic.AddUint32(&fc.index, 1)
	client := fc.clients[int(v-1)%len(fc.clients)]
	return client, nil
}

func (fc *fixedNumberClientPool) Put(client mqtt.Client) {

}

func NewDynamicClientPool(opts *mqtt.ClientOptions) MQTTClientPool {
	copy := *opts
	return &dynamicClientPool{
		opt: copy,
		pool: sync.Pool{New: func() any {
			opts := new(mqtt.ClientOptions)
			*opts = copy
			opts.SetClientID(uuid.NewString())

			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				return token.Error()
			}
			return client
		}},
	}

}

func (dc *dynamicClientPool) Get() (mqtt.Client, error) {
	v := dc.pool.Get()

	switch p := v.(type) {
	case error:
		return nil, p
	case mqtt.Client:
		return p, nil
	}
	return nil, errors.New("unreachable code")
}

func (dc *dynamicClientPool) Put(client mqtt.Client) {
	dc.pool.Put(client)
}
