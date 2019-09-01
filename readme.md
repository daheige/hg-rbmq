# rbmq 安装

    Ubuntu18.04安装RabbitMQ
    一.安装erlang

    由于rabbitMq需要erlang语言的支持，在安装rabbitMq之前需要安装erlang

    sudo apt-get install erlang-nox
    二.安装Rabbitmq

    更新源

    sudo apt-get update
    安装

    sudo apt-get install rabbitmq-server
    启动、停止、重启、状态rabbitMq命令

    sudo rabbitmq-server start
    sudo rabbitmq-server stop
    sudo rabbitmq-server restart
    sudo rabbitmqctl status
    三.添加admin，并赋予administrator权限

    添加admin用户，密码设置为admin。

    sudo rabbitmqctl add_user  admin  admin

    赋予权限

    sudo rabbitmqctl set_user_tags admin administrator

    赋予virtual host中所有资源的配置、写、读权限以便管理其中的资源

    sudo rabbitmqctl  set_permissions -p / admin '.*' '.*' '.*'
    四.启动

    安装了Rabbitmq后，默认也安装了该管理工具，执行命令即可启动

    sudo  rabbitmq-plugins enable rabbitmq_management（先定位到rabbitmq安装目录）
    浏览器访问http://localhost:15672/

# rbmq 参考

    https://blog.csdn.net/vrg000/article/details/81165030

# go rbmq 参考

    https://blog.csdn.net/lastsweetop/article/details/91038836
    https://www.cnblogs.com/hsnblog/p/9956566.html

# go rbmq 运行

    生产者：
        heige@daheige:/web/go/hg-rbmq$ go run pro/production.go
        2019/08/12 22:21:12 push success
        heige@daheige:/web/go/hg-rbmq$ go run pro/production.go
        2019/08/12 22:21:14 push success
        heige@daheige:/web/go/hg-rbmq$ go run pro/production.go
        2019/08/12 22:21:15 push success

    消费者：
        heige@daheige:/web/go/hg-rbmq/consume$ go run client.go
        2019/08/12 22:17:25 Received a message: Hello World33!
        2019/08/12 22:21:04 Received a message: Hello World33!
        2019/08/12 22:21:12 Received a message: Hello World33!
        2019/08/12 22:21:14 Received a message: Hello World33!
        2019/08/12 22:21:15 Received a message: Hello World33!

    当访问http://localhost:15672/#/queues/%2F/hello
    getMsg发现消息已经被消费掉了

# 压力测试
    200个生产者 200w消息,重用连接ch
    go run pro/production.go
    2019/09/01 10:43:23 send message ok
    2019/09/01 10:43:23 cost time:  5m56.307315797s
    
    
    100个消费者 60个独立协程进行消费
    go run consume/client_ack.go
    
    访问http://127.0.0.1:15672/#/queues 查看
    生产者qps 2w/s
    消费者qps 1.2w/s
    ack qps  5000/s
    
    直观分布图： http://127.0.0.1:15672/#/queues/%2F/hello

# 参考文档
    http://raylei.cn/index.php/archives/50/
    
    
    