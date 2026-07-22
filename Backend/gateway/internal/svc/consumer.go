package svc

import (
	"IMM/common/pkg"
	"IMM/common/types"
	"encoding/json"
	"log"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// RabbitMQ消息队列消费者，处理服务器推送给用户

// 业务通知
func StartMqEventConsumer(svcCtx *ServiceContext) {
	// 从 channel 中消费消息，手动 ack
	events, err := svcCtx.MqEventChannel.Consume(
		"im.gateway.push.event", // 队列名
		"",                      // consumer tag（自动生成）
		false,                   // auto-ack: false，手动确认
		false,                   // exclusive
		false,                   // no-local
		false,                   // no-wait
		nil,                     // args
	)
	if err != nil {
		log.Fatalf("failed to consume: %v", err)
	}

	// 无限循环处理消息
	for msg := range events {
		// 解析消息体（假设为 JSON）
		var event types.MqMsg
		// fmt.Printf("\n——————EnvetConsumer——————\n%v\n", msg)
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			logx.Errorf("unmarshal error: %v", err)
			msg.Nack(false, false) // 无法解析，直接拒绝不重试
			continue
		}

		// 构造推送内容（可以包含 请求ID、type 和 data）
		pushMsg, _ := json.Marshal(map[string]interface{}{
			"req_id": event.ReqId,
			"type":   event.Type,
			"data":   event.Data,
		})

		// 调用 Hub 推送给目标用户
		if svcCtx.ClientHub.SendToUser(event.Uid, pushMsg) {
			msg.Ack(false) // 确认消费成功
		} else {
			// 投递失败，Ack后不重回消息队列，暂存至Redis中，等连接恢复重新拉取，若Redis中对应消息过期再去数据库中取
			msg.Ack(false)
			/* Reids操作
			svcCtx.OffLineStorage.Save(event.Uid, &pkg.Message{
				Type:      event.Type,
				Data:      event.Data,
				Timestamp: time.Now().Unix(),
			})
			*/
		}
	}
}

// 聊天信息通知
func StartMqChatConsumer(svcCtx *ServiceContext) {
	// 从 channel 中消费消息，手动 ack
	chatMsg, err := svcCtx.MqMsgChannel.Consume(
		"im.gateway.push.chat", // 队列名
		"",                     // consumer tag（自动生成）
		false,                  // auto-ack: false，手动确认
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		log.Fatalf("failed to consume: %v", err)
	}

	// 无限循环处理消息
	for msg := range chatMsg {
		// 解析消息体（假设为 JSON）
		var chat types.MqMsg
		if err := json.Unmarshal(msg.Body, &chat); err != nil {
			logx.Errorf("unmarshal error: %v", err)
			msg.Nack(false, false) // 无法解析，直接拒绝不重试
			continue
		}

		// 构造推送内容（可以包含 type 和 data）
		pushMsg, _ := json.Marshal(map[string]interface{}{
			"req_id": chat.ReqId,
			"type":   chat.Type,
			"data":   chat.Data,
		})

		// 调用 Hub 推送给目标用户
		if svcCtx.ClientHub.SendToUser(chat.Uid, pushMsg) {
			msg.Ack(false) // 确认消费成功
		} else {
			// 投递失败，Ack后不重回消息队列，暂存至Redis中，等连接恢复重新拉取，若Redis中对应消息过期再去数据库中取
			msg.Ack(false)
			// Reids操作
			svcCtx.OffLineStorage.Save(chat.Uid, &pkg.Message{
				Type:      chat.Type,
				Data:      chat.Data,
				Timestamp: time.Now().Unix(),
			})
		}
	}
}
