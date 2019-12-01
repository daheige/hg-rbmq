package main

import (
	"encoding/json"
	"hg-rbmq/helper"
	"log"

	"github.com/streadway/amqp"
)

type MsgData struct {
	Id   int    `json:"id"`
	Data string `json:"data"`
}

func main() {

	// 建立rbmq连接
	conn, err := amqp.Dial("amqp://admin:admin@127.0.0.1:5672")
	helper.FailOnError(err, "failed to connect to rbmq")

	defer conn.Close()

	// 创建通道
	ch, err := conn.Channel()
	helper.FailOnError(err, "open a channel error")
	defer ch.Close()

	// 申请队列
	// 在发送消息之前我们需要用信道申请一个队列，存放我们的消息，具体的参数解释，请看代码注释
	// (name string, durable, autoDelete, exclusive, noWait bool, args Table)
	// durable 是否持久化 当我们重启我们的服务器时，我们的队列仍然存在
	// 因为服务会把持久化的queue存放在硬盘上，但是消息的持久化，还需要在发送消息的时候进行设置

	// autoDelete 对列没有用到的时候是否删除
	// exclusive 是否设置排他，true为是。如果设置为排他，则队列仅对首次声明他的连接可见，并在连接断开时自动删除
	// noWait 是否非阻塞，true表示是。阻塞：表示创建交换器的请求发送后，阻塞等待RMQ Server返回信息
	//arguments 其他参数 type Table map[string]interface{} 是一个hash类型
	q, err := ch.QueueDeclare(
		"hello",
		true,
		false,
		false,
		false,
		nil,
	)

	helper.FailOnError(err, "declare queue failed")

	//发送消息
	// 在发送消息的时候需要指定交换器和队列的名字，这样消费者才能准确的收到消息从该队列中
	// (exchange, key string, mandatory, immediate bool, msg Publishing)
	// exchange 交换机的名称
	// key 路由的名称，队列的名字 q.Name
	// mandatory mandatory
	// msg Publishing rbmq发送消息格式,是一个结构体

	body := "hello,heige"
	for i := 0; i < 800000; i++ {
		msg := MsgData{
			Id:   i,
			Data: body,
		}

		b, err := json.Marshal(msg)
		if err != nil {
			log.Println("json encode error: ", err)

			continue
		}

		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent, //消息持久化 Transient (0 or 1) or Persistent (2)
				ContentType:  "text/plain",    //消息格式 text/plain纯文本 application/octet-stream 数据流
				Body:         b,               // The application specific payload of the message
			},
		)

		if err != nil {
			log.Println(err, "send msg failed")
		} else {
			log.Println("send msg success")
		}
	}

	log.Println("send all msg success")

}

/**
启动生产者 go run app.go
*/
