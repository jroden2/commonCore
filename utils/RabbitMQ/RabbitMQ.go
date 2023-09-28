package rabbitConfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jroden2/backend_assignments/assignment_01/commonCore/utils"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type rabbitMQ struct {
	logger      *zerolog.Logger
	connections map[string]connection
}

func NewRabbitMQ(logger *zerolog.Logger) RabbitMQ {
	if logger == nil {
		tLogger := utils.InitialiseLogger()
		logger = &tLogger
	}

	rq := &rabbitMQ{logger: logger, connections: map[string]connection{}}
	queueDetailsMap = make(map[string]queueDetails)
	rq.BindOrDeclareQueue("RABBITMQ_URI", AuditExchange, "topic", AuditRoutingKey, AuditQueue)

	go rq.ManageConnections()

	return rq
}

type RabbitMQ interface {
	Listen(queueName string, sub *Subscriber)
	Send(queueName string, requestBody interface{}) error
}

func (r *rabbitMQ) BindOrDeclareQueue(environment, exchange, exchangeType, routingKey, queueName string) {
	rConn, err := r.getConnection(queueName)
	if err != nil && rConn == nil {
		utils.LogError("BindOrDeclareQueue", err, *r.logger)
	}
	if rConn != nil && !rConn.ch.IsClosed() {
		utils.LogInfo("BindOrDeclareQueue", "connection is open", *r.logger)
		return
	}

	// ---------

	var conn *amqp.Connection
	uri, ok := os.LookupEnv(environment)
	if !ok {
		utils.LogFatal("BindOrDeclareQueue", fmt.Errorf("failed to find ENV var for %s", environment), *r.logger)
	}

	conn, err = amqp.Dial(uri)
	if err != nil {
		if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "dial tcp") {
			utils.LogError("BindOrDeclareQueue", errors.New("connection to RabbitMQ lost, reconnecting"), *r.logger)
			return
		}
		utils.LogError("BindOrDeclareQueue", err, *r.logger)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		utils.LogError("BindOrDeclareQueue", err, *r.logger)
		return
	}

	err = ch.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		utils.LogError("BindOrDeclareQueue", err, *r.logger)
		return
	}

	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		utils.LogError("BindOrDeclareQueue", err, *r.logger)
		return
	}

	err = ch.QueueBind(
		queueName,  // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		utils.LogError("BindOrDeclareQueue", err, *r.logger)
		return
	}

	cnnx := connection{
		ch:         ch,
		queueName:  queueName,
		routingKey: routingKey,
		exchange:   exchange,
	}
	r.connections[queueName] = cnnx

	queueDetailsMap[queueName] = queueDetails{
		Environment:  environment,
		Exchange:     exchange,
		ExchangeType: exchangeType,
		RoutingKey:   routingKey,
		QueueName:    queueName,
	}
	utils.LogInfo("bindOrDeclare", fmt.Sprintf("connected to %s queue", queueName), *r.logger)
}

func (r *rabbitMQ) ManageConnections() {
	utils.LogInfo("ManageConnections", "managing connections", *r.logger)
	for {
		for qN := range r.connections {
			tQ, ok := queueDetailsMap[qN]
			if !ok {
				utils.LogError("ManageConnections", fmt.Errorf("queue details for %s queue not found", qN), *r.logger)
			}

			conn, err := r.getConnection(qN)
			if err != nil || conn.ch.IsClosed() {
				utils.LogError("ManageConnections", fmt.Errorf("queue %s connection closed, reconnecting", qN), *r.logger)
				r.BindOrDeclareQueue(tQ.Environment, tQ.Exchange, tQ.ExchangeType, tQ.RoutingKey, tQ.QueueName)
			}
		}
		time.Sleep(90 * time.Second)
	}
}

func (r *rabbitMQ) ReconnectQueue(queueName string) error {
	utils.LogInfo("ReconnectQueue", fmt.Sprintf("reconnecting to '%s' queue", queueName), *r.logger)
	tQ, ok := queueDetailsMap[queueName]
	if !ok {
		err := fmt.Errorf("queue details for %s queue not found", queueName)
		utils.LogError("ManageConnections", err, *r.logger)
		return err
	}

	r.BindOrDeclareQueue(tQ.Environment, tQ.Exchange, tQ.ExchangeType, tQ.RoutingKey, tQ.QueueName)
	return nil
}

func (r *rabbitMQ) Listen(queueName string, sub *Subscriber) {
	var err error
	var ch *amqp.Channel
	var conn *connection
	conn, err = r.getConnection(queueName)
	if err != nil {
		utils.LogError("Listen", err, *r.logger)
		err = r.ReconnectQueue(queueName)
		if err != nil {
			utils.LogFatal("Listen", err, *r.logger)
		}
	}
	ch = conn.ch
	utils.LogInfo("Listen", fmt.Sprintf("listening to %s queue", queueName), *r.logger)

	msgs, err := ch.Consume(
		conn.queueName, // queue
		"OMI-Producer", // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		utils.LogError("BindOrDeclareQueue", err, *r.logger)
		utils.LogFatal("BindOrDeclareQueue", errors.New("failed to register a consumer"), *r.logger)
	}

	var alive chan struct{}
	go func() {
		for m := range msgs {
			utils.LogInfo("Listen", "Ack OK", *r.logger)
			sub.ServeDelivery(ch)(&m)
		}
	}()
	<-alive
}

func (r *rabbitMQ) Send(queueName string, requestBody interface{}) error {
	var err error
	var ch *amqp.Channel
	var conn *connection
	conn, err = r.getConnection(queueName)
	if err != nil {
		utils.LogError("Send", err, *r.logger)
		err = r.ReconnectQueue(queueName)
		if err != nil {
			utils.LogFatal("Send", err, *r.logger)
		}
	}
	ch = conn.ch

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(requestBody)
	if err != nil {
		utils.LogError("Send", err, *r.logger)
		return err
	}

	err = ch.PublishWithContext(ctx,
		conn.exchange,   // exchange
		conn.routingKey, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		utils.LogError("Send", err, *r.logger)
		return err
	}
	return nil
}

func (r *rabbitMQ) getConnection(queueName string) (*connection, error) {
	conn, exists := r.connections[queueName]
	if !exists {
		return nil, fmt.Errorf("connection for queue '%s' not found", queueName)
	}
	if conn.ch.IsClosed() {
		return nil, fmt.Errorf("connection for queue '%s' closed", queueName)
	}

	return &conn, nil
}
