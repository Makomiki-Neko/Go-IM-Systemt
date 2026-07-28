// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"IMM/common/pkg"
	"IMM/gateway/internal/config"
	"IMM/gateway/internal/hub"
	"IMM/rpc/ai/aiservice"
	"IMM/rpc/chat/chatservice"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	ClientHub      *hub.Hub
	MqConn         *amqp091.Connection // 维护与RabbitMQ的连接
	MqEventChannel *amqp091.Channel    // 事件队列
	MqMsgChannel   *amqp091.Channel    // 消息队列
	OffLineStorage *pkg.RedisStorage
	ChatRPC        chatservice.ChatService
	AiRPC          aiservice.AiService
	S3             *s3.S3
}

func NewServiceContext(c config.Config) *ServiceContext {

	// 1. 连接 RabbitMQ
	conn, err := amqp091.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		c.RabbitMQ.Username, c.RabbitMQ.Password,
		c.RabbitMQ.Host, c.RabbitMQ.Port, c.RabbitMQ.VHost))
	if err != nil {
		log.Fatalf("failed to connect rabbitmq: %v", err)
	}
	ch, err := conn.Channel()    // 用于事件推送Channel
	chMsg, err := conn.Channel() // 消息推送Channel
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}

	// 2. 声明交换机（Exchange）和队列（Queue）
	// topic 交换机，支持通配符 * #，用于模糊路由；direct：直连交换机，路由键必须与绑定键完全相等（点对点）；fanout：扇出交换机，忽略路由键，广播到所有绑定队列
	err = ch.ExchangeDeclare("im.events", "topic", true, false, false, false, nil) // 业务事件
	err = ch.ExchangeDeclare("im.chat", "topic", true, false, false, false, nil)   // 聊天信息
	if err != nil {
		log.Fatalf("failed to declare exchange: %v", err)
	}

	// 声明 Gateway 队列
	queueEvent, err := ch.QueueDeclare("im.gateway.push.event", true, false, false, false, nil)
	queueChat, err := ch.QueueDeclare("im.gateway.push.chat", true, false, false, false, nil)
	queueAi, err := ch.QueueDeclare("im.gateway.push.ai", true, false, false, false, nil) // LLM 调用请求队列
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	// 绑定队列到交换机
	err = ch.QueueBind(queueEvent.Name, "im.gateway.push.event.#", "im.events", false, nil)
	err = ch.QueueBind(queueAi.Name, "im.gateway.push.ai.#", "im.events", false, nil)
	err = ch.QueueBind(queueChat.Name, "im.gateway.push.chat.#", "im.chat", false, nil)
	if err != nil {
		log.Fatalf("failed to bind queue: %v", err)
	}

	// 创建S3客户端
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(c.SeaweedFS_S3.Region),   // 区域是 SDK 必填项，SeaweedFS 不校验[reference:8]
		Endpoint:         aws.String(c.SeaweedFS_S3.Endpoint), // 指向你启动的 S3 Gateway 地址
		Credentials:      credentials.NewStaticCredentials(c.SeaweedFS_S3.AccessKey, c.SeaweedFS_S3.SecretKey, ""),
		DisableSSL:       aws.Bool(true), // SeaweedFS 默认 HTTP
		S3ForcePathStyle: aws.Bool(true), // 使用URL路径
	})
	if err != nil {
		log.Fatalf("Failed to Init S3 Session: %v", err)
	}
	s3 := s3.New(sess)

	ctx := &ServiceContext{
		Config:         c,
		ClientHub:      hub.NewHub(),
		MqConn:         conn,
		MqEventChannel: ch,
		MqMsgChannel:   chMsg,
		OffLineStorage: &pkg.RedisStorage{
			Rdb: redis.NewClient(&redis.Options{Addr: c.RedisConf.Host, Password: c.RedisConf.Pass, DB: c.RedisConf.DB}),
			Ctx: context.Background(),
		},
		ChatRPC: chatservice.NewChatService(zrpc.MustNewClient(c.ChatRPC)),
		AiRPC:   aiservice.NewAiService(zrpc.MustNewClient(c.AiRPC)),
		S3:      s3,
	}

	// 事件通知队列消费者
	go StartMqEventConsumer(ctx)
	// 聊天信息推送消费者
	go StartMqChatConsumer(ctx)

	return ctx
}
