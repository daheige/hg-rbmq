package main

import (
	"context"
	"flag"
	"log"
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

	//从队列中取出消息进行消费
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)

	go func() {
		for msg := range msgs {
			log.Printf("Received a message: %s", msg.Body)
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
