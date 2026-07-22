package pkg

import (
	"context"
	"errors"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// RabbitClient RabbitMQ 客户端封装
type RabbitClient struct {
	conn     *amqp091.Connection
	exchange string // 默认交换机
}

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, msg []byte) error

// NewRabbitClient 初始化 RabbitMQ 连接与默认交换机
func NewRabbitClient(dsn, exchange string) (*RabbitClient, error) {
	conn, err := amqp091.Dial(dsn)
	if err != nil {
		return nil, err
	}

	client := &RabbitClient{
		conn:     conn,
		exchange: exchange,
	}

	// 启动时声明默认交换机（Topic 类型，灵活度最高，适合业务扩展）
	err = client.declareExchange(exchange)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	// 监听连接关闭事件，生产环境可在此触发重连
	go func() {
		closeChan := make(chan *amqp091.Error)
		conn.NotifyClose(closeChan)
		err := <-closeChan
		logx.Errorf("RabbitMQ连接断开: %v", err)
		// TODO: 生产环境补充自动重连逻辑
	}()

	logx.Info("RabbitMQ 初始化成功")
	return client, nil
}

// declareExchange 声明交换机（持久化，服务重启不丢失）
func (c *RabbitClient) declareExchange(name string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.ExchangeDeclare(
		name,
		"topic", // 交换机类型：主题模式，支持通配符路由
		true,    // 持久化
		false,   // 不自动删除
		false,   // 非内部交换机
		false,   // 不等待
		nil,
	)
}

// Publish 发布消息
func (c *RabbitClient) Publish(ctx context.Context, routingKey string, body []byte) error {
	if c.conn == nil || c.conn.IsClosed() {
		return errors.New("rabbitmq连接已断开")
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 发布持久化消息
	return ch.PublishWithContext(ctx,
		c.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent, // 消息持久化
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		})
}

// Consume 注册消费者（手动ACK + 公平分发）
func (c *RabbitClient) Consume(queueName, routingKey string, handler MessageHandler) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	// 声明持久化队列
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return err
	}

	// 队列绑定到交换机
	err = ch.QueueBind(queueName, routingKey, c.exchange, false, nil)
	if err != nil {
		_ = ch.Close()
		return err
	}

	// 公平分发：每个消费者同时只处理 1 条消息，处理完才接收下一条
	err = ch.Qos(1, 0, false)
	if err != nil {
		_ = ch.Close()
		return err
	}

	// 注册消费者（关闭自动ACK）
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return err
	}

	// 后台协程监听消息
	go func() {
		defer ch.Close()
		for d := range msgs {
			ctx := context.Background()
			// 执行业务处理
			err := handler(ctx, d.Body)
			if err != nil {
				logx.Errorf("消息处理失败: %v, 消息内容: %s", err, d.Body)
				// 处理失败：false 表示只拒绝当前消息，true 表示重新入队
				// 生产环境建议：重试3次失败后转入死信队列，避免无限循环
				_ = d.Nack(false, true)
				continue
			}
			// 处理成功：手动确认
			_ = d.Ack(false)
		}
		logx.Infof("队列 %s 消费者退出", queueName)
	}()

	logx.Infof("消费者启动成功，队列: %s, 路由键: %s", queueName, routingKey)
	return nil
}

// Close 关闭连接
func (c *RabbitClient) Close() {
	if c.conn != nil && !c.conn.IsClosed() {
		_ = c.conn.Close()
	}
}
