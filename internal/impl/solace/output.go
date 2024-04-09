package solace

import (
	"context"
	"fmt"
	"sync"

	"github.com/benthosdev/benthos/v4/public/service"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/resource"
)

// ---go:embed input_description.md
var outputDescription string = "Demo"

func solaceOuputSpec() *service.ConfigSpec {

	cfg := []*service.ConfigField{
		service.NewObjectField(endpointObject,
			service.NewStringField(queueNameField).Example("my/queue").Optional(),
			service.NewInterpolatedStringField(topicNameField).Example("my/dynamic/topic"),
		),
		service.NewOutputMaxInFlightField(),
	}
	cfg = append(cfg, getDefaultConfigFields()...)

	return service.NewConfigSpec().
		Categories("Services").
		Summary("Publishes messages to a Solace PubSub+ Message Broker").
		Description(outputDescription).
		Fields(cfg...)
}

func init() {
	err := service.RegisterOutput("solace", solaceOuputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (out service.Output, maxInFlight int, err error) {
			output, err := solaceOutputFromConfig(conf, mgr)
			if err != nil {
				return nil, 0, err
			}
			return output, output.max_in_flight, err
		})
	if err != nil {
		panic(err)
	}
}

type solaceOutput struct {
	client              *Client
	persistentPublisher solace.PersistentMessagePublisher
	endpointConfig      endpointConfig
	max_in_flight       int
	log                 *service.Logger
	m                   sync.RWMutex
}

func (o *solaceOutput) Connect(ctx context.Context) error {
	o.m.Lock()
	defer o.m.Unlock()

	err := o.client.Connect(ctx)
	if err != nil {
		return err
	}

	o.log.Debug("Creating persistent message producer")
	o.persistentPublisher, err = o.client.getMessagingService().CreatePersistentMessagePublisherBuilder().
		OnBackPressureWait(uint(o.max_in_flight)).Build()
	if err != nil {
		return err
	}

	o.log.Debug("Starting persistent message producer")
	if err := o.persistentPublisher.Start(); err != nil {
		return err
	}

	return nil
}

func (a *solaceOutput) Close(ctx context.Context) error {
	if a.client == nil {
		return fmt.Errorf("client is nil")
	}
	return a.client.Disconnect(ctx)
}

func (o *solaceOutput) Write(ctx context.Context, msg *service.Message) error {

	builder := o.client.getMessagingService().MessageBuilder()

	data, err := msg.AsBytes()
	if err != nil {
		return err
	}
	message, err := builder.BuildWithByteArrayPayload(data)
	if err != nil {
		return err
	}

	// interpolate topic name
	topic := resource.TopicOf(o.endpointConfig.topicName.String(msg))
	// this is probably slow, but the easiest method for now
	// return o.persistentPublisher.PublishAwaitAcknowledgement(message, topic, 5*time.Second, nil)
	return o.persistentPublisher.Publish(message, topic, nil, nil)
}

func solaceOutputFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (output *solaceOutput, err error) {
	client, err := NewSharedSolaceClient(mgr.Logger(), conf)
	if err != nil {
		return nil, err
	}
	output = &solaceOutput{
		log:            mgr.Logger(),
		client:         client,
		endpointConfig: endpointConfig{},
	}
	output.max_in_flight, err = conf.FieldMaxInFlight()
	if err != nil {
		return nil, err
	}

	if output.endpointConfig.topicName, err = conf.FieldInterpolatedString(endpointObject, topicNameField); err != nil {
		return nil, err
	}

	return output, err
}
