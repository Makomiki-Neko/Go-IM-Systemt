package main

import (
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// ========== 配置常量 ==========
// 本地 RabbitMQ 默认连接信息：
// - 默认端口 5672（管理界面是 15672，不要搞混）
// - 默认账号密码 guest/guest（仅本地访问可用）
// - 默认虚拟主机 /
const (
	rabbitURL = "amqp://guest:guest@127.0.0.1:5672/"
	queueName = "demo_hello_queue"
)

// failOnError 是一个辅助函数：统一处理错误并打印日志
// 教学目的：RabbitMQ 操作几乎每一步都可能出错，必须检查错误
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err) // Fatalf 会打印日志并退出程序
	}
}

func main1() {
	log.Println("=== RabbitMQ 可用性验证程序启动 ===")

	// ========== 第 1 步：建立 TCP 连接 ==========
	// 教学点：
	// - Connection 是 TCP 长连接，重量级对象，一个进程通常只建一个
	// - 不要每条消息都建连接，会严重损耗性能
	conn, err := amqp091.Dial(rabbitURL)
	failOnError(err, "无法连接到 RabbitMQ，请检查服务是否启动、端口和账号是否正确")
	defer conn.Close() // 程序退出前关闭连接，释放资源
	log.Println("✅ 成功建立 RabbitMQ 连接")

	// ========== 第 2 步：创建通道 Channel ==========
	// 教学点：
	// - Channel 是轻量级的，建立在 Connection 之上
	// - 绝大多数 API 操作都在 Channel 上进行（声明队列、发消息、消费消息）
	// - 一个 Connection 可以开多个 Channel，不同协程建议用不同 Channel
	ch, err := conn.Channel()
	failOnError(err, "无法创建 Channel")
	defer ch.Close()
	log.Println("✅ 成功创建 Channel")

	// ========== 第 3 步：声明队列 ==========
	// 教学点：
	// - QueueDeclare 是幂等的：队列不存在则创建，已存在则直接使用
	// - 参数说明：
	//   1. name: 队列名
	//   2. durable: 是否持久化（true=重启RabbitMQ后队列还在，但消息不一定，消息也要标记持久化）
	//   3. autoDelete: 是否自动删除（最后一个消费者断开后删除队列）
	//   4. exclusive: 是否排他（仅当前连接可见，连接关闭自动删除）
	//   5. noWait: 是否不等待服务器确认
	//   6. args: 额外参数（如死信、TTL等）
	_, err = ch.QueueDeclare(
		queueName, // 队列名称
		false,     // durable: 演示用不需要持久化
		false,     // autoDelete
		false,     // exclusive
		false,     // noWait
		nil,       // args
	)
	failOnError(err, "声明队列失败")
	log.Printf("✅ 队列 [%s] 已就绪", queueName)

	// ========== 第 4 步：启动消费者协程 ==========
	// 教学点：消费者是持续监听的，所以放在独立 goroutine 中运行
	go startConsumer(ch)

	// 稍微等一下消费者启动，避免第一条消息发了消费者还没注册上
	time.Sleep(500 * time.Millisecond)

	// ========== 第 5 步：生产者发送消息 ==========
	startProducer(ch)

	// 给消费者一点时间处理完所有消息
	log.Println("⏳ 等待 2 秒让消费者处理完毕...")
	time.Sleep(2 * time.Second)

	log.Println("=== 验证结束，程序退出 ===")
}

// startProducer 生产者：发送 5 条测试消息
func startProducer(ch *amqp091.Channel) {
	log.Println("📤 生产者开始发送消息...")

	for i := 1; i <= 5; i++ {
		// 消息体必须是 []byte 类型
		body := fmt.Sprintf("Hello RabbitMQ! 这是第 %d 条消息", i)

		// Publish 发布消息
		// 参数说明：
		// 1. exchange: 交换机名，空字符串表示使用默认交换机（direct类型）
		// 2. key: 路由键，使用默认交换机时，路由键 = 队列名
		// 3. mandatory: 是否强制路由（消息无法路由到队列时返回给生产者）
		// 4. immediate: 是否立即（没有消费者在监听则返回）
		// 5. msg: 消息内容，Publishing 结构体
		err := ch.Publish(
			"",        // exchange: 默认交换机
			queueName, // routing key: 队列名
			false,     // mandatory
			false,     // immediate
			amqp091.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
				// DeliveryMode: amqp.Persistent, // 消息持久化，需要队列也durable才生效
			},
		)
		failOnError(err, "发送消息失败")

		log.Printf("📤 已发送: %s", body)
		time.Sleep(200 * time.Millisecond) // 模拟发送间隔
	}

	log.Println("📤 生产者所有消息发送完毕")
}

// startConsumer 消费者：持续监听队列并打印收到的消息
func startConsumer(ch *amqp091.Channel) {
	log.Println("📥 消费者启动，开始监听队列...")

	// Consume 注册消费者，返回一个 <-chan Delivery 通道
	// 参数说明：
	// 1. queue: 队列名
	// 2. consumer: 消费者标签，用于区分不同消费者
	// 3. autoAck: 是否自动确认
	//    - true: 消息发出后立即算消费成功，不管业务有没有处理完
	//    - false: 手动确认，业务处理完调用 d.Ack(false) 才会从队列删除
	// 4. exclusive: 是否排他消费
	// 5. noLocal: 不接收本连接发送的消息（RabbitMQ不支持，忽略）
	// 6. noWait: 是否不等待
	// 7. args: 额外参数
	msgs, err := ch.Consume(
		queueName, // 队列名
		"",        // consumer tag，空字符串由服务器自动生成
		true,      // autoAck: 演示用自动确认，简单方便
		false,     // exclusive
		false,     // noLocal
		false,     // noWait
		nil,       // args
	)
	failOnError(err, "注册消费者失败")

	// 用 range 持续读取消息通道，通道关闭则循环结束
	for d := range msgs {
		log.Printf("📥 收到消息: %s", d.Body)
		// 如果 autoAck = false，这里需要手动确认：
		// d.Ack(false) // false 表示只确认当前这一条消息
	}

	log.Println("📥 消费者通道关闭，退出")
}
