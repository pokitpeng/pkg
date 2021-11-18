package rabbitmq

import (
	"github.com/pokitpeng/pkg/logger"
	"github.com/streadway/amqp"
)

var (
	MQ  Pool // mq连接池
	log = logger.NewDevelopLog()
)

func GetChannel(v interface{}) (*amqp.Channel, error) {
	return v.(*amqp.Connection).Channel()
}

// RabbitMQ ...
type RabbitMQ struct {
	QueueName    string
	ExchangeName string
	ExchangeKind string // exchange kind
	Key          string // routingKey
	args         amqp.Table
}

// NewRabbitMQ ...
func NewRabbitMQ(queueName, exchangeName, exchangeKind, key string, args amqp.Table) *RabbitMQ {
	return &RabbitMQ{
		QueueName:    queueName,
		ExchangeName: exchangeName,
		ExchangeKind: exchangeKind,
		Key:          key,
		args:         args,
	}
}

// PublishRouting publish routing msg
func (r *RabbitMQ) PublishRouting(rch *amqp.Channel, message string) error {
	q, err := rch.QueueDeclare(
		r.QueueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		r.args,      // arguments
	)
	if err != nil {
		log.Debugf("PublishRouting err:%s", err.Error())
		return err
	}
	// log.Debugf("QueueDeclare: %v", r.QueueName)

	// 1.create exchange
	err = rch.ExchangeDeclare(
		r.ExchangeName,
		r.ExchangeKind,
		true,
		false,
		false,
		false,
		r.args,
	)
	if err != nil {
		log.Debugf("ExchangeDeclare err: %s", err.Error())
		return err
	}
	// log.Debugf("ExchangeDeclare: %v", r.ExchangeName)

	// 2. bind queue
	err = rch.QueueBind(
		q.Name,         // queue name
		r.Key,          // routing key
		r.ExchangeName, // exchange
		false,
		r.args)
	if err != nil {
		log.Errorf("QueueBind err:%s", err.Error())
		return err
	}
	// log.Debugf("QueueBind queue bind success, exchange:%v,queue:%v", r.ExchangeName, r.QueueName)

	// 3.send
	err = rch.Publish(
		r.ExchangeName,
		r.Key, // set routingKey
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(message),
		})
	if err != nil {
		log.Errorf("Publish err:%s", err.Error())
		return err
	}
	err = rch.Confirm(false)
	if err != nil {
		log.Errorf("mq confirm msg err:%s", err.Error())
		return err
	}
	log.Debugf("Publish: %v", message)
	return nil
}

// BroadcastRouting publish msg to many queue
func (r *RabbitMQ) BroadcastRouting(rch *amqp.Channel, message string, queues ...string) error {
	// 1.create exchange
	err := rch.ExchangeDeclare(
		r.ExchangeName,
		r.ExchangeKind,
		true,
		false,
		false,
		false,
		r.args,
	)
	if err != nil {
		log.Errorf("ExchangeDeclare err:%s", err.Error())
		return err
	}
	log.Debugf("ExchangeDeclare: %v", r.ExchangeName)

	for _, queue := range queues {
		q, err := rch.QueueDeclare(
			queue,  // name
			true,   // durable
			false,  // delete when unused
			false,  // exclusive
			false,  // no-wait
			r.args, // arguments
		)
		if err != nil {
			log.Errorf("QueueDeclare err:%s", err.Error())
			return err
		}
		log.Debugf("QueueDeclare: %v", r.QueueName)

		// 2. bind queue
		err = rch.QueueBind(
			q.Name,         // queue name
			r.Key,          // routing key
			r.ExchangeName, // exchange
			false,
			r.args)
		if err != nil {
			log.Errorf("QueueBind err:%s", err.Error())
			return err
		}
		log.Debugf("QueueBind: exchange:%v,queue:%v", r.ExchangeName, r.QueueName)
	}

	// 3.send
	err = rch.Publish(
		r.ExchangeName,
		r.Key, // set routingKey
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(message),
		})
	if err != nil {
		log.Errorf("Publish err:%s", err.Error())
		return err
	}
	log.Debugf("Publish: %v", message)
	return nil
}

// ReceiveRouting receive routing msg
func (r *RabbitMQ) ReceiveRouting(rch *amqp.Channel) (msgs <-chan amqp.Delivery, err error) {
	err = rch.ExchangeDeclare(
		r.ExchangeName, // name
		r.ExchangeKind, // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		r.args,         // arguments
	)
	if err != nil {
		log.Errorf("ExchangeDeclare err:%s", err.Error())
		return
	}

	q, err := rch.QueueDeclare(
		r.QueueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		r.args,      // arguments
	)
	if err != nil {
		log.Errorf("QueueDeclare err:%s", err.Error())
		return
	}

	err = rch.QueueBind(
		q.Name,         // queue name
		r.Key,          // routing key
		r.ExchangeName, // exchange
		false,
		r.args)
	if err != nil {
		log.Errorf("QueueBind err:%s", err.Error())
		return
	}
	err = rch.Qos(1, 0, true) // handler msg one by one
	if err != nil {
		return
	}

	msgs, err = rch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		r.args, // args
	)
	if err != nil {
		log.Errorf("Consume err:%s", err.Error())
		return
	}

	return
}

// GetMsgCount get msg count
func (r *RabbitMQ) GetMsgCount(rch *amqp.Channel) (int, error) {
	q, err := rch.QueueDeclare(
		r.QueueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		r.args,      // arguments
	)
	if err != nil {
		return 0, err
	}
	return q.Messages, nil
}

// CreateVhost create vhost
// func CreateVhost(httpPort int, vhost string) error {
// 	url := fmt.Sprintf("http://%s:%d/api/vhosts/%s", define.MQHost, httpPort, vhost)
// 	req, err := http.NewRequest("PUT", url, nil)
// 	if err != nil {
// 		// handle err
// 		return err
// 	}
// 	req.SetBasicAuth(define.MQUsername, define.MQPassword)
//
// 	resp, err := http.DefaultClient.Do(req)
//
// 	if err != nil {
// 		// handle err
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	return nil
//
// }
