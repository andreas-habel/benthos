package solace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benthosdev/benthos/v4/public/service"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"
)

// ---go:embed input_description.md
var inputDescription string = "Demo"

func solaceInputSpec() *service.ConfigSpec {

	cfg := []*service.ConfigField{
		service.NewObjectField(endpointObject,
			service.NewStringField(queueNameField).Example("my/queue"),
		),
	}
	cfg = append(cfg, getDefaultConfigFields()...)

	return service.NewConfigSpec().
		Categories("Services").
		Summary("Reads messages from Solace PubSub+ Message Broker").
		Description(inputDescription).
		Fields(cfg...)
}

func init() {
	err := service.RegisterBatchInput("solace", solaceInputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			return solaceInputFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type receivedMessage struct {
	message message.InboundMessage
	err     error
}

type solaceInput struct {
	client             *Client
	persistentReceiver solace.PersistentMessageReceiver
	endpointConfig     endpointConfig
	log                *service.Logger
	m                  sync.RWMutex
}

func solaceInputFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (input *solaceInput, err error) {
	client, err := NewSharedSolaceClient(mgr.Logger(), conf)
	if err != nil {
		return nil, err
	}

	input = &solaceInput{
		log:            mgr.Logger(),
		client:         client,
		endpointConfig: endpointConfig{},
	}

	if input.endpointConfig.queueName, err = conf.FieldString(endpointObject, queueNameField); err != nil {
		return nil, err
	}
	return input, err
}

func (i *solaceInput) Connect(ctx context.Context) error {
	i.m.Lock()
	defer i.m.Unlock()

	err := i.client.Connect(ctx)
	if err != nil {
		return err
	}

	i.log.Debug("Creating persistent message receiver")
	queue := resource.QueueDurableExclusive(i.endpointConfig.queueName)
	persistentReceiver, err := i.client.getMessagingService().CreatePersistentMessageReceiverBuilder().
		WithMessageClientAcknowledgement().Build(queue)
	if err != nil {
		return err
	}
	i.persistentReceiver = persistentReceiver

	i.log.Debug("Starting persistent message receiver")
	if err := persistentReceiver.Start(); err != nil {
		return err
	}

	return nil
}

func (i *solaceInput) Close(ctx context.Context) error {
	if i.persistentReceiver != nil && i.persistentReceiver.IsRunning() {
		i.log.Info("Shutting down persistent message receiver")
		i.persistentReceiver.Terminate(30 * time.Second) // TODO: make timeout configurable
	}

	return i.client.Disconnect(ctx)
}

// getMessage waits for a new incoming message and returns it in a channel
func getMessage(msgChan chan<- receivedMessage, reader solace.PersistentMessageReceiver) {
	// endless retrieval of messages, due to missing ctx support of CGO.
	// will be canceled from outside
	msg, err := reader.ReceiveMessage(-1)
	msgChan <- receivedMessage{message: msg, err: err}
	close(msgChan)
}

func (a *solaceInput) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	var part *service.Message
	var rmsg receivedMessage
	var msg message.InboundMessage

	if a.persistentReceiver == nil {
		return nil, nil, fmt.Errorf("persistent receiver is nil")
	}
	if !a.persistentReceiver.IsRunning() {
		return nil, nil, fmt.Errorf("persistent receiver is not running")
	}

	msgChan := make(chan receivedMessage)
	go getMessage(msgChan, a.persistentReceiver)

	select {
	case rmsg = <-msgChan:
		if rmsg.err != nil {
			return nil, nil, rmsg.err
		}
		msg = rmsg.message
	case <-ctx.Done():
		a.log.Debugf("received cancellation of ReadBatch due to error \"%v\"", ctx.Err())
		return nil, nil, ctx.Err()
	}

	if data, ok := msg.GetPayloadAsBytes(); ok {
		part = service.NewMessage(data)
	} else if data, ok := msg.GetPayloadAsString(); ok {
		part = service.NewMessage([]byte(data))
	}
	// TODO: add msg.GetPayloadAsMap()
	if part == nil {
		return nil, nil, fmt.Errorf("could not extract content from message")
	}

	// Set message headers and structured data properties as metadata
	part.MetaSetMut(metaPrefix+"destination_name", msg.GetDestinationName())
	setStringHeader(part, "message_id", msg.GetApplicationMessageID)
	setStringHeader(part, "message_type", msg.GetApplicationMessageType)
	setStringHeader(part, "correlation_id", msg.GetCorrelationID)
	setStringHeader(part, "content_encoding", msg.GetHTTPContentEncoding)
	setStringHeader(part, "content_type", msg.GetHTTPContentType)
	setStringHeader(part, "sender_id", msg.GetSenderID)

	if rgmi, ok := msg.GetReplicationGroupMessageID(); ok {
		part.MetaSetMut(metaPrefix+"rgmi", rgmi.String())
	}

	if prio, ok := msg.GetPriority(); ok {
		part.MetaSetMut(metaPrefix+"priority", prio)
	}

	for k, v := range msg.GetProperties() {
		part.MetaSetMut(metaPrefix+"prop_"+k, v)
	}

	return service.MessageBatch{part}, func(ctx context.Context, err error) error {
		return a.persistentReceiver.Ack(msg)
	}, nil
}

func setStringHeader(part *service.Message, headerName string, headerFunc func() (string, bool)) {
	if value, ok := headerFunc(); ok {
		part.MetaSetMut(metaPrefix+headerName, value)
	}
}
