# 关闭连接

    func (ch *Channel) Close() error
    func (c *Connection) Close() error

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

    持久化的作用是防止在重启、关闭、宕机的情况下数据丢失，持久化是相对靠谱的，如果数据在内存中，没来得及存盘就发生了重启，那么数据还是会丢失。
    持久化可分为三类：
    1、Exchange的持久化（创建时设置参数）
    2、Queue的持久化（创建时设置参数）
    3、Msg的持久化（发送时设置Args）

# 公平分发

    有时候队列的轮询调度并不能满足我们的需求，假设有这么一个场景，存在两个消费者程序，所有的单数序列消息都是长耗时任务而双数序列消息则都是简单任务，那么结果将是一个消费者一直处于繁忙状态而另外一个则几乎没有任务被挂起。当RabbitMQ对此情况却是视而不见，仍然根据轮询来分发消息。

    导致这种情况发生的根本原因是RabbitMQ是根据消息的入队顺序进行派发，而并不关心在线消费者还有多少未确认的消息，它只是简单的将第N条消息分发到第N个消费者：

    p---->queue--->c
                -->c
    队列设置预取数(prefect count)为1
    为了避免这种情况，我们可以给队列设置预取数(prefect count)为1。它告诉RabbitMQ不要一次性分发超过1个的消息给某一个消费者，换句话说，就是当分发给该消费者的前一个消息还没有收到ack确认时，RabbitMQ将不会再给它派发消息，而是寻找下一个空闲的消费者目标进行分发。

    err = ch.Qos(
        1,      // prefetch count
        0,      // prefetch size
        false,  // global
    )
    log.Println(err)

    go rbmq源码
    Qos
    func (ch *Channel) Qos(prefetchCount, prefetchSize int, global bool) error
    注意：这个在推送模式下非常重要，通过设置Qos用来防止消息堆积。
    prefetchCount：消费者未确认消息的个数。
    prefetchSize ：消费者未确认消息的大小。
    global ：是否全局生效，true表示是。全局生效指的是针对当前connect里的所有channel都生效。

# 关于队列长度

    NOTE：如果所有的消费者都繁忙，队列可能会被消息填满。你需要注意这种情况
    要么通过增加消费者来处理，要么改用其他的策略

# 发送消息

    发送消息
    func (ch *Channel) Publish(exchange, key string, mandatory, immediate bool, msg Publishing) error
    exchange：要发送到的交换机名称，对应图中exchangeName。
    key：路由键，对应图中RoutingKey。
    mandatory：直接false，不建议使用，后面有专门章节讲解。
    immediate ：直接false，不建议使用，后面有专门章节讲解。
    msg：要发送的消息，msg对应一个Publishing结构，Publishing结构里面有很多参数，这里只强调几个参数，其他参数暂时列出，但不解释。

    # cat $(find ./amqp) | grep -rin type.*Publishing
    type Publishing struct {
            Headers Table
            // Properties
            ContentType     string //消息的类型，通常为“text/plain”
            ContentEncoding string //消息的编码，一般默认不用写
            DeliveryMode    uint8  //消息是否持久化，2表示持久化，0或1表示非持久化。
            Body []byte //消息主体
            Priority        uint8 //消息的优先级 0 to 9
            CorrelationId   string    // correlation identifier
            ReplyTo         string    // address to to reply to (ex: RPC)
            Expiration      string    // message expiration spec
            MessageId       string    // message identifier
            Timestamp       time.Time // message timestamp
            Type            string    // message type name
            UserId          string    // creating user id - ex: "guest"
            AppId           string    // creating application id
    }

# 接受消息 – 推模式

    RMQ Server主动把消息推给消费者

    func (ch *Channel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args Table) (<-chan Delivery, error)
    queue:队列名称。
    consumer:消费者标签，用于区分不同的消费者。
    autoAck:是否自动回复ACK，true为是，回复ACK表示高速服务器我收到消息了。建议为false，手动回复，这样可控性强。
    exclusive:设置是否排他，排他表示当前队列只能给一个消费者使用。
    noLocal:如果为true，表示生产者和消费者不能是同一个connect。
    nowait：是否非阻塞，true表示是。阻塞：表示创建交换器的请求发送后，阻塞等待RMQ Server返回信息。非阻塞：不会阻塞等待RMQ Server的返回信息，而RMQ Server也不会返回信息。（不推荐使用）
    args：直接写nil，没研究过，不解释。
    注意下返回值：返回一个<- chan Delivery类型，遍历返回值，有消息则往下走， 没有则阻塞

# 绑定

    到现在，我们已经创建了一个fanout类型的exchange和一个队列，然而exchange并不知道它要哪个队列是应该被分发消息的。因此，我们需要明确的指定exchange和队列队列之间的关系，这个操作称之为绑定。
        err = ch.QueueBind(
        q.Name,    //queue name
        "",        //routing key
        "logs",    //exchange
        false,
        nil
    )
    现在，logs exchange中的消息就会被分发到我们的队列。

    查看绑定列表，仍然可以通过命令来查看：

    rabbitmqctl list_bindings

    emit_log.go
    https://github.com/rabbitmq/rabbitmq-tutorials/blob/master/go/emit_log.go
