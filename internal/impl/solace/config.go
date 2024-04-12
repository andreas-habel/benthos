package solace

import (
	"github.com/benthosdev/benthos/v4/public/service"
	"solace.dev/go/messaging/pkg/solace/config"
)

const (
	obfuscated = "*********"
	metaPrefix = "solace_"

	endpointObject = "endpoint"
	topicNameField = "topic_name"
	queueNameField = "queue_name"
)

type endpointConfig struct {
	queueName string
	topicName *service.InterpolatedString
}

var obfuscatedProperties = []config.ServiceProperty{
	config.AuthenticationPropertySchemeBasicPassword,
	config.AuthenticationPropertySchemeClientCertPrivateKeyFilePassword,
	config.AuthenticationPropertySchemeOAuth2AccessToken,
	config.AuthenticationPropertySchemeOAuth2OIDCIDToken,
}

func getDefaultConfigFields() []*service.ConfigField {
	return []*service.ConfigField{

		// Transport Configuration
		service.NewObjectField("transport",
			service.NewStringField(stripPropertyName(config.TransportLayerPropertyHost)).
				Description("Host is an IPv4 or IPv6 address or host name of the broker to which to connect. Multiple entries are permitted when each address is separated by a comma. The entry for the Host property should provide a protocol, host and port.").
				Example("tcp://localhost:55555").
				Example("tcps://localhost:55443"),
			service.NewIntField(stripPropertyName(config.TransportLayerPropertyKeepAliveInterval)).
				Description("KeepAliveInterval is the amount of time (in milliseconds) to wait between sending out Keep-Alive messages.").
				Advanced().Optional().Default(32000), // TODO: clarify corrent default value
			service.NewIntField(stripPropertyName(config.TransportLayerPropertyConnectionRetries)).
				Description("connection_retries is how many times to try connecting to a broker during connection setup.").
				Advanced().Optional().Default(5),
			service.NewIntField(stripPropertyName(config.TransportLayerPropertyConnectionRetriesPerHost)).
				Description("connection_retries_per_host defines how many connection or reconnection attempts are made to a single host before moving to the next host in the list, when using a host list.").
				Advanced().Optional().Default(3),

			service.NewIntField(stripPropertyName(config.TransportLayerPropertyReconnectionAttempts)).
				Description(`Reconnection Attempts is the number reconnection attempts to the broker (or list of brokers) after a connected MessagingService goes down.
Zero means no automatic reconnection attempts, while a -1 means attempt to reconnect forever. The default valid range is greather than or equal to -1.
When using a host list, each time the API works through the host list without establishing a connection is considered a
reconnection retry. Each reconnection retry begins with the first host listed. After each unsuccessful attempt to reconnect
to a host, the API waits for the amount of time set for `+"`reconnection_attempts_wait_interval`"+` before attempting another
connection to a broker. The number of times attempted to connect to one broker before moving on to the
next listed host is determined by the value set for `+"`connection_retries_per_host`").
				Advanced().Optional().Default(-1),
			service.NewIntField(stripPropertyName(config.TransportLayerPropertyReconnectionAttemptsWaitInterval)).
				Description(`reconnection_attempts_wait_iinterval sets how much time (in milliseconds) to wait between each connection or reconnection attempt to the configured host.
If a connection or reconnection attempt to the configured host (which may be a list) is not successful, the API waits for
the amount of time set for reconnection_attempts_wait_interval, and then makes another connection or reconnection attempt.
The valid range is greater than or equal to zero.`).
				Advanced().Optional().Default(1000),
		).Description("Transport object allows to configure connection informations like hostname"),

		// Service Configuration
		service.NewObjectField("service",
			service.NewStringField(stripPropertyName(config.ServicePropertyVPNName)).
				Description("ServicePropertyVPNName name of the Message VPN to attempt to join when connecting to a broker.").
				Example("default").
				Default("default"),
			service.NewBoolField(stripPropertyName(config.ServicePropertyGenerateReceiveTimestamps)).
				Description("__(Input only)__ receive_timestamps specifies whether timestamps should be generated on inbound messages.").
				Advanced().Optional().Default(false),
			service.NewBoolField(stripPropertyName(config.ServicePropertyGenerateSenderID)).
				Description("__(Output only)__ generate_sender_id specifies whether the client name should be included in the SenderID message-header parameter.").
				Advanced().Optional().Default(false),
			service.NewBoolField(stripPropertyName(config.ServicePropertyGenerateSendTimestamps)).
				Description("__(Output only)__ generate_send_timestamps specifies whether timestamps should be generated for outbound messages.").
				Advanced().Optional().Default(false),
		).Description("service properties allow to configure the message vpn to connect to and specific client behaviors"),

		// Authentication Configuration
		service.NewObjectField("authentication",
			service.NewStringEnumField(stripPropertyName(config.AuthenticationPropertyScheme),
				config.AuthenticationSchemeBasic,
				config.AuthenticationSchemeClientCertificate,
				config.AuthenticationSchemeKerberos,
				config.AuthenticationSchemeOAuth2,
			).Description("AuthenticationPropertyScheme defines the keys for the authentication scheme type.").
				Example(config.AuthenticationSchemeBasic).
				Default(config.AuthenticationSchemeBasic),

			// BasicAuth Configuration
			service.NewObjectField("basic",
				service.NewStringField(stripPropertyName(config.AuthenticationPropertySchemeBasicUserName)).
					Description("Username specifies the username for basic authentication.").
					Optional(),
				service.NewStringField(stripPropertyName(config.AuthenticationPropertySchemeBasicPassword)).
					Description("Password specifies password for basic authentication.").
					Optional().Secret(),
			).Description("Configuration for basic auth"),
			service.NewObjectField("oauth2",
				service.NewStringField(stripPropertyName(config.AuthenticationPropertySchemeOAuth2AccessToken)).
					Description("AccessToken specifies an access token for OAuth 2.0 token-based authentication.").
					Optional().Secret(),
				service.NewStringField(stripPropertyName(config.AuthenticationPropertySchemeOAuth2IssuerIdentifier)).
					Description("IssuerIdentifier defines an optional issuer identifier for OAuth 2.0 token-based authentication.").
					Optional(),
				service.NewStringField(stripPropertyName(config.AuthenticationPropertySchemeOAuth2OIDCIDToken)).
					Description("OIDCIDToken specifies the ID token for Open ID Connect token-based authentication.").
					Optional(),
			).Description("Configuration for oauth2 auth").Advanced(),
			service.NewObjectField("client_certificate",
				service.NewStringField(stripPropertyName(config.AuthenticationPropertySchemeSSLClientCertFile)).
					Description("AuthenticationPropertySchemeSSLClientCertFile specifies the client certificate file used for Secure Socket Layer (SSL).").
					Optional(),
				service.NewStringField(stripPropertyName(config.AuthenticationPropertySchemeSSLClientPrivateKeyFile)).
					Description("AuthenticationPropertySchemeSSLClientPrivateKeyFile specifies the client private key file.").
					Optional(),
				service.NewStringField(stripPropertyName(config.AuthenticationPropertySchemeClientCertPrivateKeyFilePassword)).
					Description("PrivateKeyFilePassword specifies the private key password for client certificate authentication.").
					Optional().Secret(),
				service.NewStringField(stripPropertyName(config.AuthenticationPropertySchemeClientCertUserName)).
					Description("Username specifies the username to use when connecting with client certificate authentication.").
					Optional(),
			).Description("Configuration for client_certificate auth").Advanced(),
		),
	}
}
