// go test --count=1 -run ^TestRabbitMQ$
package rabbitmq

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pokitpeng/pkg/logger"
	"github.com/streadway/amqp"
)

var (
	mqurl = "amqp://admin:datatom.com@192.168.50.97:30211/"

	exchangeKind = "fanout"
	maxPriority  = 255
	queueName    = "test"
	exchangeName = "test"
	routingKey   = "test"
)

func TestRabbitMQ(t *testing.T) {
	initMQ()

	// send()
	receive()

	// 释放连接池中的所有连接
	defer MQ.Release()
}

func initMQ() {
	var err error
	// factory 创建连接的方法
	factory := func() (interface{}, error) { return amqp.Dial(mqurl) }

	// close 关闭连接的方法
	close := func(v interface{}) error {
		connection := v.(*amqp.Connection)
		return connection.Close()
	}

	// ping 检测连接的方法
	ping := func(v interface{}) (err error) {
		closeChan := make(chan *amqp.Error, 1)
		ch, err := v.(*amqp.Connection).Channel()
		if err != nil {
			return err
		}
		notifyClose := ch.NotifyClose(closeChan)
		select {
		case <-notifyClose:
			return err
		default:
			return nil
		}
	}

	// 创建一个连接池： 初始化5，最大空闲连接是20，最大并发连接30
	poolConfig := &Config{
		InitialCap: 5,  // 资源池初始连接数
		MaxIdle:    20, // 最大空闲连接数
		MaxCap:     30, // 最大并发连接数
		Factory:    factory,
		Close:      close,
		Ping:       ping,
		// 连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: 30 * time.Second,
	}
	MQ, err = NewChannelPool(poolConfig)
	if err != nil {
		fmt.Println("err=", err)
	}

	// do something
	// conn:=v.(*amqp.Connection).Channel()

}

func send() {
	var wg sync.WaitGroup
	logger.NewDevEnv()

	rc, err := MQ.Get()
	if err != nil {
		logger.Errorf("get rabbitmq connection err:%v", err)
		return
	}
	defer MQ.Put(rc)

	rch, err := GetChannel(rc)
	defer rch.Close()

	args := make(amqp.Table)
	args["x-max-priority"] = int64(maxPriority)
	rabbitmq := NewRabbitMQ(queueName, exchangeName, exchangeKind, routingKey, args)

	closeChan := make(chan *amqp.Error, 1)
	// if channel err,throw amqp.Error and catch the err
	notifyClose := rch.NotifyClose(closeChan)

	for i := 0; i < 10; i++ {
		go func(i int) {
			wg.Add(1)
			defer wg.Done()

			time.Sleep(time.Duration(i) * time.Second)
			select {
			case err := <-notifyClose:
				logger.Panicf("notifyClose, err:%v", err)
			default:
				msg := "hello" + strconv.Itoa(i)
				err = rabbitmq.PublishRouting(rch, msg)
				if err != nil {
					logger.Errorf("PublishRouting err:%v", err)
				}
			}
		}(i)
	}

	wg.Wait()
	logger.Infof("goroutine finished")
	time.Sleep(1 * time.Second)
}

func receive() {
	rc, err := MQ.Get()
	if err != nil {
		logger.Errorf("get rabbitmq connection err:%v", err)
		return
	}
	defer MQ.Put(rc)

	rch, err := GetChannel(rc)
	defer rch.Close()

	args := make(amqp.Table)
	args["x-max-priority"] = int64(maxPriority)
	rabbitmq := NewRabbitMQ(queueName, exchangeName, exchangeKind, routingKey, args)

	msgs, err := rabbitmq.ReceiveRouting(rch)
	if err != nil {
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
