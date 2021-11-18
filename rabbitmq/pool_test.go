package rabbitmq

import (
	"fmt"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestPool(t *testing.T) {
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
		IdleTimeout: 15 * time.Second,
	}
	p, err := NewChannelPool(poolConfig)
	if err != nil {
		fmt.Println("err=", err)
	}

	// do something
	// conn:=v.(*amqp.Connection).Channel()

	// 释放连接池中的所有连接
	defer p.Release()

	// 查看当前连接中的数量

	// 从连接池中取连接
	num := 10
	var cc = []interface{}{}
	for i := 0; i < num; i++ {
		v, err := p.Get()
		if err != nil {
			log.Infof("get connect err:%s", err.Error())
		}
		cc = append(cc, v)
	}
	time.Sleep(time.Millisecond * 100)
	log.Infof("after get, connect num:%d", p.Len())
	// 将连接放回连接池中
	for _, v := range cc {
		p.Put(v)
	}
	log.Infof("after put, connect num:%d", p.Len())
	time.Sleep(2 * time.Second)
}
