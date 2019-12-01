package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

func main() {
	//发送消息
	var nums = 200
	var wg sync.WaitGroup
	t := time.Now()
	wg.Add(nums)
	for i := 0; i < nums; i++ {
		go func() {
			defer wg.Done()

			sendMsg()
		}()
	}

	wg.Wait()

	log.Println("send message ok")
	log.Println("cost time: ", time.Now().Sub(t))
}

func sendMsg() {
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

	for i := 0; i < 10000; i++ {
		//随机停顿1-5ms
		time.Sleep(time.Duration(rand.Int63n(5)) * time.Millisecond)

		now := time.Now()
		rand.Seed(now.UnixNano())
		body := map[string]interface{}{
			"msg":        "Hello World33!",
			"rnd":        rand.Int63n(999999),
			"index":      i,
			"created_at": now.Format("2006-01-02 15:04:05"),
		}

		b, _ := json.Marshal(body)

		err = ch.Publish("", q.Name, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        b,
		})

		if err != nil {
			log.Println("push msg error: ", err)
			return
		}

		log.Println("push success")
	}

}
