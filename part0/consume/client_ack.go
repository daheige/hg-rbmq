package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

var wait time.Duration

func init() {
	flag.DurationVar(&wait, "-wait-time", 2*time.Second, "graceful wait time")
	flag.Parse()
}

func custMsg(i int) {
	log.Println("current consume index: ", i)
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
	q, err := ch.QueueDeclare(
		"hello", //name 队列
		false,   //durable 是否持久化
		false,   //autoDelete 是否自动删除
		false,
		false,
		nil)

	if err != nil {
		log.Println("define queue error: ", err)
		return
	}

	//从队列中取出消息进行消费
	//手动确认消息被消息了
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)

	/**
	消息ack
	当消费者的autoack为true时，一旦收到消息就会直接把该消息设置为删除状态，如果消息的处理时间之内，消费者挂掉了那么这条消息就会丢失掉。
	rabbitmq支持消息ack机制，将autoack设为false，当处理完毕再手动触发ack操作。如果处理消息的过程中挂掉了，那么这条消息就会分发给其他都消费者。
	*/
	for msg := range msgs {
		log.Printf("Received a message: %s", msg.Body)

		//模拟消息耗时
		time.Sleep(time.Duration(rand.Int63n(20)) * time.Millisecond)
		log.Printf("Done")

		//发送确认消息
		msg.Ack(false)
	}

}
func main() {
	//开启多个消费者
	var nums = 100
	for i := 0; i < nums; i++ {
		go custMsg(i)
	}

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
