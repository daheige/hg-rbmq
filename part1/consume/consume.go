package main

import (
	"context"
	"flag"
	"hg-rbmq/helper"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

var wait time.Duration

func init() {
	flag.DurationVar(&wait, "wait-time", 3*time.Second, "consume gracefully exit, eg: 15s or 1m")
	flag.Parse()
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
	// arguments 其他参数 type Table map[string]interface{} 是一个hash类型
	q, err := ch.QueueDeclare(
		"hello",
		true,
		false,
		false,
		false,
		nil,
	)

	helper.FailOnError(err, "declare queue failed")

	// (queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args Table)
	// queue 队列名称
	// consumer 消费者名称
	// autoAck 是否自动确认，消费者和rbmq服务端是否自动应答
	// exclusive 是否排他
	// noLocal 如果设置为true，则生产者与消费者不能是同一连接
	// noWait 是否非阻塞 true阻塞,false非阻塞
	// arguments 其他参数 type Table map[string]interface{} 是一个hash类型

	msgs, err := ch.Consume(q.Name, "hello-consume", false, false, false, false, nil)
	helper.FailOnError(err, "create consumer failed")

	// 开启独立携程进行消费
	// 是否设置自动应答，这里推荐false我们使用了data.Acknowledger.Ack(d.DeliveryTag, true)进行手动应答
	// 防止消息丢失，只有确保我们收到消息，才应该告诉服务器，将这条消息进行删除，另外需要注意的是
	// 我们创建了一个协程来持续消费消息，只要生产者生产消息，发送到相应的队列，我们就可以对消息进行打印
	log.Println("wait msg come...")

	go func() {
		log.Println("consume msg...")

		for data := range msgs {
			log.Println("recv msg: ", string(data.Body))

			// 消费者手动应答
			// data.Acknowledger.Ack(data.DeliveryTag, true)

			// 下面这一行和上面效果一样
			data.Ack(true)
		}
	}()

	chSig := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// recivie signal to exit main goroutine
	//window signal
	// signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGHUP)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2, os.Interrupt, syscall.SIGHUP)

	// Block until we receive our signal.
	sig := <-chSig

	log.Println("exit signal: ", sig.String())
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	<-ctx.Done()

	log.Println("shutting down")
}

/**
启动消费者 go run consume.go
// 当我们启动发送者，这里就会接收到消息
2019/12/01 16:55:29 recv msg:  {"id":810,"data":"hello,heige"}
2019/12/01 16:55:29 recv msg:  {"id":811,"data":"hello,heige"}
2019/12/01 16:55:29 recv msg:  {"id":5,"data":"hello,heige"}
2019/12/01 16:55:29 recv msg:  {"id":42,"data":"hello,heige"}
2019/12/01 16:55:29 recv msg:  {"id":41,"data":"hello,heige"}
2019/12/01 16:55:29 recv msg:  {"id":43,"data":"hello,heige"}
2019/12/01 16:55:29 recv msg:  {"id":3,"data":"hello,heige"}
2019/12/01 16:55:29 recv msg:  {"id":2,"data":"hello,heige"}

*/
