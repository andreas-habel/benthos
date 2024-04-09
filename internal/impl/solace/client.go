package solace

import (
	"context"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/service"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
)

// Client is the shared implementation for the input and output component
type Client struct {
	messagingService solace.MessagingService
	properties       config.ServicePropertyMap
	log              *service.Logger
}

func NewSharedSolaceClient(logger *service.Logger, conf *service.ParsedConfig) (*Client, error) {
	client := &Client{
		log:        logger,
		properties: make(config.ServicePropertyMap),
	}
	if err := client.initSharedConfig(conf); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) getMessagingService() solace.MessagingService {
	return c.messagingService
}

func (c *Client) initSharedConfig(conf *service.ParsedConfig) (err error) {
	if c.properties[config.TransportLayerPropertyHost], err = conf.FieldString("transport", stripPropertyName(config.TransportLayerPropertyHost)); err != nil {
		return err
	}
	if c.properties[config.TransportLayerPropertyKeepAliveInterval], err = conf.FieldInt("transport", stripPropertyName(config.TransportLayerPropertyKeepAliveInterval)); err != nil {
		return err
	}
	if c.properties[config.TransportLayerPropertyReconnectionAttempts], err = conf.FieldInt("transport", stripPropertyName(config.TransportLayerPropertyReconnectionAttempts)); err != nil {
		return err
	}
	if c.properties[config.TransportLayerPropertyReconnectionAttemptsWaitInterval], err = conf.FieldInt("transport", stripPropertyName(config.TransportLayerPropertyReconnectionAttemptsWaitInterval)); err != nil {
		return err
	}
	if c.properties[config.TransportLayerPropertyConnectionRetries], err = conf.FieldInt("transport", stripPropertyName(config.TransportLayerPropertyConnectionRetries)); err != nil {
		return err
	}
	if c.properties[config.TransportLayerPropertyConnectionRetriesPerHost], err = conf.FieldInt("transport", stripPropertyName(config.TransportLayerPropertyConnectionRetriesPerHost)); err != nil {
		return err
	}

	if c.properties[config.ServicePropertyVPNName], err = conf.FieldString("service", stripPropertyName(config.ServicePropertyVPNName)); err != nil {
		return err
	}
	if c.properties[config.ServicePropertyGenerateSenderID], err = conf.FieldBool("service", stripPropertyName(config.ServicePropertyGenerateSenderID)); err != nil {
		return err
	}
	if c.properties[config.ServicePropertyGenerateSendTimestamps], err = conf.FieldBool("service", stripPropertyName(config.ServicePropertyGenerateSendTimestamps)); err != nil {
		return err
	}
	if c.properties[config.ServicePropertyGenerateReceiveTimestamps], err = conf.FieldBool("service", stripPropertyName(config.ServicePropertyGenerateReceiveTimestamps)); err != nil {
		return err
	}

	if c.properties[config.AuthenticationPropertyScheme], err = conf.FieldString("authentication", stripPropertyName(config.AuthenticationPropertyScheme)); err != nil {
		return err
	}
	if c.properties[config.AuthenticationPropertySchemeBasicUserName], err = conf.FieldString("authentication", "basic", stripPropertyName(config.AuthenticationPropertySchemeBasicUserName)); err != nil {
		return err
	}
	if c.properties[config.AuthenticationPropertySchemeBasicPassword], err = conf.FieldString("authentication", "basic", stripPropertyName(config.AuthenticationPropertySchemeBasicPassword)); err != nil {
		return err
	}

	// print all config properties, when debug logging is active, but obfuscate passwords!
	for k, v := range c.properties {
		for _, obf := range obfuscatedProperties {
			if k == obf {
				v = obfuscated
			}
		}
		c.log.With("solace", "config").Debugf("solace property %s=%v", k, v)
	}

	return nil
}

func (c *Client) Connect(ctx context.Context) error {
	var err error
	c.log.Debug("Creating solace messaging service")
	c.messagingService, err = messaging.NewMessagingServiceBuilder().
		FromConfigurationProvider(c.properties).
		Build()
	if err != nil {
		return err
	}

	c.log.Debug("Trying to connect to solace broker")
	if err := c.messagingService.Connect(); err != nil {
		return err
	}

	return nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	c.log.Info("disconnecting from solace broker")
	if c.messagingService != nil {
		return c.messagingService.Disconnect()
	}
	return fmt.Errorf("messaging service is nil")
}
