package svc

import (
	"IMM/common/types"
	"IMM/gateway/internal/hub"
	"IMM/rpc/chat/chat"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// 处理源自客户端长连接中发送的内容，解析客户端推送给RPC服务

// handleMessage 解析并路由消息，生产者，重新发送至RPC服务
func HandleMessage(client *hub.Client, data []byte, svcCtx *ServiceContext) {
	var req struct {
		Type    string          `json:"type"` // 暂且定义为 "业务类型(heartBeat/chat/updataFile/ack).操作类型(privateMsg/groupMsg/fileType)"
		ReqId   uint64          `json:"reqId"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &req); err != nil {
		// 无效消息，返回错误响应
		sendErrorResponse(client, req.ReqId, "invalid json", svcCtx)
		return
	}

	fmt.Printf("\n ————————Receive:%s———————— \n", req.Type)

	bus := strings.Split(req.Type, ".")

	switch bus[0] {
	case "heartBeat":
		// 心跳响应，直接返回空响应
		sendResponse(client, "heartBeat", req.ReqId, 0, fmt.Sprintf("Received HeartBeat, Your Send is %v: ", req))
	case "chat":
		// 处理聊天消息（可能调用RPC或直接转发）
		handleChat(client, req.ReqId, req.Payload, bus[1], svcCtx)
	case "updateFile":
		handleFile(client, req.ReqId, req.Payload, bus[1], svcCtx)
	case "ack":
		// 客户端确认消息，可能更新消息状态
		handleAck(client, req.ReqId, req.Payload, bus[1], svcCtx)
		//logx.Infof("ack received: %d", req.ReqId)
	default:
		sendErrorResponse(client, req.ReqId, "unknown type", svcCtx)
	}
}

func handleChat(client *hub.Client, reqId uint64, payload json.RawMessage, msgType string, svcCtx *ServiceContext) {
	// 解析业务操作类型，进行业务处理（调用RPC处理消息，然后转发给目标用户）
	ctx := context.Background()
	switch msgType {
		case "SendPrivateMsg":
		//fmt.Printf("\n ————————Handle:%s———————— \n", msgType)
		var in struct {
			FromUserId  uint64 `json:"from_user_id,string"`
			ToUserId    uint64 `json:"to_user_id,string"`
			MsgType     int32  `json:"msg_type"`
			Content     string `json:"content"`
			ClientMsgId uint64 `json:"client_msg_id"`
		}
		err := json.Unmarshal(payload, &in)
		if err != nil {
			//fmt.Printf("\n ————————Handle:Marshall Failed %v———————— \n", err)
			sendErrorResponse(client, reqId, errors.New("Failed to Unmarshal Payload, err: "+err.Error()).Error(), svcCtx)
			return
		}

		//fmt.Printf("\n ————————Handle:Call RPC———————— \n")

		// 需要确保in中都存在值

		r, err := svcCtx.ChatRPC.SendPrivateMessage(ctx, &chat.SendPrivateMessageReq{
			FromUserId:  in.FromUserId,
			ToUserId:    in.ToUserId,
			MsgType:     in.MsgType,
			Content:     in.Content,
			ClientMsgId: in.ClientMsgId,
		})
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}

		// 发送消息ACK给客户端
		//fmt.Printf("\n ————————Handle:Send Ack———————— \n")
		d, _ := json.Marshal(map[string]string{
			"msg_id":        strconv.FormatUint(r.MsgId, 10),
			"client_msg_id": strconv.FormatUint(in.ClientMsgId, 10),
			"send_time":     strconv.FormatInt(r.SendTime, 10),
		})
		ackbody, _ := packagePushMsg(reqId, in.FromUserId, "ack.Msg", d)
		err = svcCtx.MqMsgChannel.Publish(
			"im.events",
			"im.gateway.push.event.ack.chat",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Transient,
				ContentType:  "application/json",
				Body:         ackbody,
				Timestamp:    time.Now(),
			},
		)

		//fmt.Printf("\n ————————Handle:Push———————— \n")
		// 封装推送消息
		body, err := packagePushMsg(reqId, in.ToUserId, "chat.privateMsg", r.CommonResponse.Data)
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}
		// 投递消息
		err = svcCtx.MqMsgChannel.Publish(
			"im.chat",
			"im.gateway.push.chat.private",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent, // 消息持久化
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
			},
		)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("Push Failed, Error: "+err.Error()).Error(), svcCtx)
		}

	case "SendGroupMsg":

	case "GetNewPrivateMsg":
		var in struct {
			UserId     uint64 `json:"user_id,string"`
			FromUserId uint64 `json:"from_user_id,string"`
			StartMsgId uint64 `json:"start_msg_id,string"`
			Limit      int32  `json:"limit"`
		}
		err := json.Unmarshal(payload, &in)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("Failed to Unmarshal Payload, err: "+err.Error()).Error(), svcCtx)
			return
		}
		r, err := svcCtx.ChatRPC.GetUnreadChatMessage(ctx, &chat.GetChatUnreadMessagesReq{
			UserId:     in.UserId,
			FromUserId: in.FromUserId,
			StartMsgId: in.StartMsgId,
			Limit:      in.Limit,
		})
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
		}
		// 封装推送消息
		body, err := packagePushMsg(reqId, in.UserId, "chat.privateUnreceiveMsgBlock", r.Data)
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}
		// 投递消息
		err = svcCtx.MqMsgChannel.Publish(
			"im.chat",
			"im.gateway.push.chat.private",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent, // 消息持久化
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
			},
		)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("Push Failed, Error: "+err.Error()).Error(), svcCtx)
		}

	case "GetNewGroupMsg":

	case "GetHistoryPrivateMsg":
		var in struct {
			UserId     uint64 `json:"user_id,string"`
			FromUserId uint64 `json:"from_user_id,string"`
			StartMsgId uint64 `json:"start_msg_id,string"`
			Limit      int32  `json:"limit"`
		}
		err := json.Unmarshal(payload, &in)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("Failed to Unmarshal Payload, err: "+err.Error()).Error(), svcCtx)
			return
		}
		r, err := svcCtx.ChatRPC.GetHistoryChatMessage(ctx, &chat.GetChatHistoryMessagesReq{
			UserId:     in.UserId,
			FromUserId: in.FromUserId,
			StartMsgId: in.StartMsgId,
			Limit:      in.Limit,
		})
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}
		// 封装推送消息
		body, err := packagePushMsg(reqId, in.UserId, "chat.privateHistoryMsgBlock", r.Data)
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}
		// 投递消息
		err = svcCtx.MqMsgChannel.Publish(
			"im.chat",
			"im.gateway.push.chat.private",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent, // 消息持久化
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
			},
		)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("Push Failed, Error: "+err.Error()).Error(), svcCtx)
		}

	case "GetHistoryGroupMsg":

	default:
		sendErrorResponse(client, reqId, "Unknown Type.", svcCtx)
	}
}

func handleFile(client *hub.Client, reqId uint64, payload json.RawMessage, msgType string, svcCtx *ServiceContext) {
	var fileInfo struct {
		Id   string `json:"file_id"` // 客户端生成，雪花唯一ID
		Name string `json:"file_name"`
		Size int64  `json:"file_size"` // KB
		T    string `json:"file_type"`
	}
	err := json.Unmarshal(payload, &fileInfo)
	if err != nil {
		sendErrorResponse(client, reqId, errors.New("Failed to Unmarshal Payload, err: "+err.Error()).Error(), svcCtx)
		return
	}
	// 最大500M
	if fileInfo.Size == 0 || fileInfo.Size > 500000 {
		sendErrorResponse(client, reqId, errors.New("File Size Exception.").Error(), svcCtx)
		return
	}
	s := strings.Split(fileInfo.Name, ".")
	fileName := fileInfo.Id + "." + s[len(s)-1] // 取文件类型后缀

	switch msgType {
	case "Picture":
		// 创建 PutObject 请求
		req, _ := svcCtx.S3.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String("my-bucket"),                    // 你的桶名
			Key:    aws.String("ChatFile/picture/" + fileName), // 文件的唯一标识，如 "user/123/avatar.jpg"
		})

		// 3. 生成预签名 URL，有效期设为 3 分钟
		urlStr, err := req.Presign(3 * time.Minute)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("File Update URL Get Failed, Error: "+err.Error()).Error(), svcCtx)
			return
		}

		fmt.Printf("————URL——————：\n%v\n\n", urlStr)

		// 推送上传URL	客户端需解码Json
		data, err := json.Marshal(map[string]string{
			"fileName": fileInfo.Name,
			"fileId":   "ChatFile/picture/" + fileName,
			"url":      urlStr,
			"expire":   strconv.FormatInt((time.Now().Add(5 * time.Minute).Unix()), 10),
		})
		body, err := packagePushMsg(reqId, client.UserId, "file.UpdateUrl", data)
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}
		// 投递消息
		err = svcCtx.MqEventChannel.Publish(
			"im.events",
			"im.gateway.push.event.file",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent, // 消息持久化
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
			},
		)

	case "Voice":
		// 创建 PutObject 请求
		req, _ := svcCtx.S3.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String("my-bucket"),                  // 你的桶名
			Key:    aws.String("ChatFile/voice/" + fileName), // 文件的唯一标识，如 "user/123/avatar.jpg"
		})

		// 3. 生成预签名 URL，有效期设为 3 分钟
		urlStr, err := req.Presign(3 * time.Minute)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("File Update URL Get Failed, Error: "+err.Error()).Error(), svcCtx)
			return
		}

		// 推送上传URL
		data, err := json.Marshal(map[string]string{
			"fileName": fileInfo.Name,
			"fileId":   "ChatFile/voice/" + fileName,
			"url":      urlStr,
			"expire":   strconv.FormatInt((time.Now().Add(3 * time.Minute).Unix()), 10),
		})
		body, err := packagePushMsg(reqId, client.UserId, "file.UpdateUrl", data)
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}
		// 投递消息
		err = svcCtx.MqEventChannel.Publish(
			"im.events",
			"im.gateway.push.event.file",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent, // 消息持久化
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
			},
		)

	case "Audio":
		// 创建 PutObject 请求
		req, _ := svcCtx.S3.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String("my-bucket"),                  // 你的桶名
			Key:    aws.String("ChatFile/audio/" + fileName), // 文件的唯一标识，如 "user/123/avatar.jpg"
		})

		// 3. 生成预签名 URL，有效期设为 3 分钟
		urlStr, err := req.Presign(3 * time.Minute)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("File Update URL Get Failed, Error: "+err.Error()).Error(), svcCtx)
			return
		}

		// 推送上传URL
		data, err := json.Marshal(map[string]string{
			"fileName": fileInfo.Name,
			"fileId":   "ChatFile/audio/" + fileName,
			"url":      urlStr,
			"expire":   strconv.FormatInt((time.Now().Add(3 * time.Minute).Unix()), 10),
		})
		body, err := packagePushMsg(reqId, client.UserId, "file.UpdateUrl", data)
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}
		// 投递消息
		err = svcCtx.MqEventChannel.Publish(
			"im.events",
			"im.gateway.push.event.file",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent, // 消息持久化
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
			},
		)

	case "Video":
		// 创建 PutObject 请求
		req, _ := svcCtx.S3.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String("my-bucket"),                  // 你的桶名
			Key:    aws.String("ChatFile/video/" + fileName), // 文件的唯一标识，如 "user/123/avatar.jpg"
		})

		// 3. 生成预签名 URL，有效期设为 10 分钟
		urlStr, err := req.Presign(10 * time.Minute)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("File Update URL Get Failed, Error: "+err.Error()).Error(), svcCtx)
			return
		}

		// 推送上传URL
		data, err := json.Marshal(map[string]string{
			"fileName": fileInfo.Name,
			"fileId":   "ChatFile/video/" + fileName,
			"url":      urlStr,
			"expire":   strconv.FormatInt((time.Now().Add(10 * time.Minute).Unix()), 10),
		})
		body, err := packagePushMsg(reqId, client.UserId, "file.UpdateUrl", data)
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}
		// 投递消息
		err = svcCtx.MqEventChannel.Publish(
			"im.events",
			"im.gateway.push.event.file",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent, // 消息持久化
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
			},
		)

	default:
		// 创建 PutObject 请求
		req, _ := svcCtx.S3.PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String("my-bucket"),                 // 你的桶名
			Key:    aws.String("ChatFile/file/" + fileName), // 文件的唯一标识，如 "user/123/avatar.jpg"
		})

		// 3. 生成预签名 URL，有效期设为 10 分钟
		urlStr, err := req.Presign(10 * time.Minute)
		if err != nil {
			sendErrorResponse(client, reqId, errors.New("File Update URL Get Failed, Error: "+err.Error()).Error(), svcCtx)
			return
		}

		// 推送上传URL
		data, err := json.Marshal(map[string]string{
			"fileName": fileInfo.Name,
			"fileId":   "ChatFile/file/" + fileName,
			"url":      urlStr,
			"expire":   strconv.FormatInt((time.Now().Add(10 * time.Minute).Unix()), 10),
		})
		body, err := packagePushMsg(reqId, client.UserId, "file.UpdateUrl", data)
		if err != nil {
			sendErrorResponse(client, reqId, err.Error(), svcCtx)
			return
		}
		// 投递消息
		err = svcCtx.MqEventChannel.Publish(
			"im.events",
			"im.gateway.push.event.file",
			false,
			false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent, // 消息持久化
				ContentType:  "application/json",
				Body:         body,
				Timestamp:    time.Now(),
			},
		)
	}
}

func handleAck(client *hub.Client, reqId uint64, payload json.RawMessage, msgType string, svcCtx *ServiceContext) {
	var ackInfo struct {
		TargetId uint64 `json:"target_id,string"`
		MsgID    uint64 `json:"msg_id,string"`
	}
	err := json.Unmarshal(payload, &ackInfo)
	if err != nil {
		sendErrorResponse(client, reqId, errors.New("Failed to Unmarshal Payload, err: "+err.Error()).Error(), svcCtx)
		return
	}
	ctx := context.Background()
	switch msgType {
	case "PrivateMsgRead":
		svcCtx.ChatRPC.UpdateChatMsgLastRead(ctx, &chat.UpdateLastReadMsgRep{
			UserId:   client.UserId,
			TargetId: ackInfo.TargetId,
			MsgId:    ackInfo.MsgID,
		})
	case "GroupMsgRead":
		svcCtx.ChatRPC.UpdataGroupMsgLastRead(ctx, &chat.UpdateLastReadMsgRep{
			UserId:   client.UserId,
			TargetId: ackInfo.TargetId,
			MsgId:    ackInfo.MsgID,
		})
	}
}

func packagePushMsg(reqId uint64, toUserId uint64, t string, data []byte) ([]byte, error) {
	pubMsg := types.MqMsg{
		Uid:   toUserId, // 发送给
		ReqId: reqId,    // 对应请求
		Type:  t,
		Data:  data,
	}
	body, err := json.Marshal(pubMsg) // json序列化
	if err != nil {
		return nil, errors.New("Push Failed, Error: " + err.Error())
	}
	return body, nil
}

// sendResponse 构造响应并发送
func sendResponse(client *hub.Client, typ string, reqId uint64, code int, data interface{}) {
	resp := map[string]interface{}{
		"type":  typ,
		"reqId": reqId,
		"code":  code,
		"data":  data,
	}
	msg, _ := json.Marshal(resp)
	// 通过发送通道写入
	select {
	case client.Send <- msg:
	default:
		// 若通道已满，丢弃或记录日志
		logx.Errorf("client send buffer full, drop response")
	}
}

func sendErrorResponse(client *hub.Client, reqId uint64, errMsg string, svcCtx *ServiceContext) {
	e, _ := json.Marshal(map[string]string{
		"ErrorMsg": errMsg,
	})
	resp := types.MqMsg{
		Uid:   client.UserId,
		ReqId: reqId,
		Type:  "Error",
		Data:  e,
	}
	msg, _ := json.Marshal(resp)
	// client.Send <- msg
	err := svcCtx.MqEventChannel.Publish(
		"im.events",
		"im.gateway.push.event.error",
		false,
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Transient, // 错误信息 非持久化
			ContentType:  "application/json",
			Body:         msg,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		logx.Errorf("Error MSG Send Failed, %v", err)
	}
}
