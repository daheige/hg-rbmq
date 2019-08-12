package main

import (
	"log"

	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://root:root@:5672/") //通过amqp协议连接到rbmq
	if err != nil {
		log.Println("rbmq connection error: ", err)
		return
	}

	defer conn.Close()
	//打开通道
	ch, err := conn.Channel()
	if err != nil {
		log.Println("open chan error: ", err)
		return
	}

	defer ch.Close()

	//声明队列
	q, err := ch.QueueDeclare("hello", false, false, false, false, nil)

	if err != nil {
		log.Println("define queue error: ", err)
		return
	}

	body := "Hello World33!"

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	})

	if err != nil {
		log.Println("push msg error: ", err)
		return
	}

	log.Println("push success")
}
