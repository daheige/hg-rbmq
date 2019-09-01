# 消息持久化
    go run pro/pro2.go
    
    go run client_ack2.go
    
    durable参数在生产者和消费者程序中都要指定为True
    
    关于消息持久化
    
    将消息设置为Persistent并不能百分百地完全保证消息不会丢失。虽然RabbitMQ知道要将消息写到磁盘，但在RabbitMQ接收到消息和写入磁盘前还是有个时间空档。
    
    因为RabbitMQ并不会对每一个消息都执行fsync(2),因此消息可能只是写入缓存而不是磁盘。
    
    所以Persistent选项并不是完全强一致性的，但在应付我们的简单场景已经足够。如需对消息完全持久化，可参考publisher confirms.
    
    访问http://localhost:15672/#/queues/%2F/hello_task
    查看任务详情
    
    当消费者消费完毕后，发送了确认ack信息后，消息就会从队列中删除
 
# 公平分发
    有时候队列的轮询调度并不能满足我们的需求，假设有这么一个场景，存在两个消费者程序，所有的单数序列消息都是长耗时任务而双数序列消息则都是简单任务，那么结果将是一个消费者一直处于繁忙状态而另外一个则几乎没有任务被挂起。当RabbitMQ对此情况却是视而不见，仍然根据轮询来分发消息。
    
    导致这种情况发生的根本原因是RabbitMQ是根据消息的入队顺序进行派发，而并不关心在线消费者还有多少未确认的消息，它只是简单的将第N条消息分发到第N个消费者：
    
    p---->queue--->c
                -->c
    为了避免这种情况，我们可以给队列设置预取数(prefect count)为1。它告诉RabbitMQ不要一次性分发超过1个的消息给某一个消费者，换句话说，就是当分发给该消费者的前一个消息还没有收到ack确认时，RabbitMQ将不会再给它派发消息，而是寻找下一个空闲的消费者目标进行分发。
    
    err = ch.Qos(
        1,      // prefetch count
        0,      // prefetch size
        false,  // global
    )
    log.Println(err)
    