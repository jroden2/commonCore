package rabbitConfig

import amqp "github.com/rabbitmq/amqp091-go"

var (
	AuditQueue      = "operation.events"
	AuditExchange   = "local.omi.audit"
	AuditRoutingKey = "userEvents"
)

type connection struct {
	ch         *amqp.Channel
	queueName  string
	routingKey string
	exchange   string
}

type queueDetails struct {
	Environment  string
	Exchange     string
	ExchangeType string
	RoutingKey   string
	QueueName    string
}

var queueDetailsMap map[string]queueDetails
