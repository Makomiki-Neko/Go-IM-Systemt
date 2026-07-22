
# Go、GoZero、Gorm、SeaweedFS、JWT、

双Token鉴权，用户登录获取唯一RefreshToken，心跳刷新AccessToken，无心跳后AccessToken自动过期，支持断网恢复
SeaweedFS，用户头像流式转发上传至Filer中

RPC-Relation 包含MQ生产者，发送好友通知
GateWay 管理与客户端之间的WebSocket长连接，带心跳机制
GateWay 包含MQ消费者与生产者，两条推送队列，im.gateway.push.chat、im.gateway.push.event


# Chat
## 消息拉取
客户端维护各会话最后收到消息ID-LastMsgID，发送消息更新请求，服务器回应最新消息数量；
客户端分批发送拉取请求，拉取新消息，一次拉取从LastMsgID开始后N条最新消息（顺序拉取）；
拉取历史信息（未保存到客户端的），LastMsgID开始往上N条消息，直到MsgID小于客户端会话中的记录时停止发送拉取请求，包括对方发送的与自己发送的；

客户端在登录时应发送以下请求:
1、拉取个人信息，拉取好友列表，拉取好友信息，拉取好友请求通知，更新当此登录的RefreshToken，定期发送心跳请求，刷新AccessToken
2、使用AccessToken建立长连接，定期响应WS的心跳
3、拉取未读消息列表，根据未读信息列表，仅主动更新拉取有未读信息的对话的新消息，使用长连接发送拉取请求

对于收到消息的已读ACK回应，仅在对话窗口打开时自动发送包含对话里最新消息ID的ACK，仅每隔一定间隔作累计确认，只在打开对话视消息已读；
对话窗口内仅显示一定数量的消息，仅在滚对上拉时，进行拉取历史消息操作；

其余未存在新消息的对话，仅在点开对话时请求拉取一次新消息更新，在一次登录期间仅拉取主动拉取更新一次（长连接断开后重置该计数，即重连后再点开该对话仍需主动拉取新消息）；

客户端维护一个自增的ReqID全局变量，每次通过WS长连接向服务器发送请求时，以其作为标识；
客户端发送的每一条消息携带一个一次性的唯一雪花ID，作为服务器端的重复判定，真实消息ID由服务器返回的Msg_ID为准；
客户端发送文件类型消息时，先通过WS长连接发送文件上传请求，获取服务器返回的S3预签名URL后，通过PUT上传，之后再将文件地址以普通消息形式发送；


## 推送类型
"event.friendApply"——好友请求
"chat.privateMsg"——即时消息（单条），在线时接收的对方发送消息；
"chat.privateUnreceiveMsgBlock"——消息数组，当前对话中客户端未接收的消息，包含对方发送与己方发送，按顺序排列；
"chat.privateHistoryMsgBlock"——历史消息数组，当前对话中客户端未持有的历史信息，按倒序排列；
"file.UpdateUrl"——文件上传地址


## 客户端长连接发送类型
"chat.SendPrivateMsg"——发送即时消息
    {
        "type":"chat.SendPrivateMsg",
        "reqId":1,
        "payload":{
            "from_user_id":2069443741298462720,
            "to_user_id":2072248047311523840,
            "msg_type":1,
            "content":"Hello Niya!",
            "client_msg_id":0
        }
    }

"chat.GetNewPrivateMsg"——获取当前客户端未持有的新消息
    {
        "type":"chat.GetNewPrivateMsg",
        "reqId":1,
        "payload":{
            "from_user_id":2069443741298462720,
            "user_id":2072248047311523840,
            "start_msg_id":0,
            "limit":50
        }
    }

"chat.GetHistoryPrivateMsg"——获取当前客户端未持有的历史信息
    {
        "type":"chat.GetHistoryPrivateMsg",
        "reqId":2,
        "payload":{
            "from_user_id":2069443741298462720,
            "user_id":2072248047311523840,
            "start_msg_id":2,
            "limit":50
        }
    }

"updateFile.Picture/Voice/Audio/Video/File"——获取文件上传路径
    {
        "type":"updateFile.Picture",
        "reqId":3,
        "payload":{
            "file_id":"123456789",
            "file_name":"test.jpg",
            "file_size":200,
            "file_type":"Jpeg"
        }
    }

"ack.PrivateMsgRead"——私聊信息已读响应
    {
        "type":"ack.PrivateMsgRead",
        "reqId":4,
        "payload":{
            "target_id":456,    // 好友ID
            "msg_id":123
        }
    }

"ack.GroupMsgRead"——群聊信息已读响应
    {
        "type":"ack.GroupMsgRead",
        "reqId":4,
        "payload":{
            "target_id":456,    // 群组ID
            "msg_id":123
        }
    }


# Redis
Redis用户全局未读消息数，"im:chat:unread:%d" —— K: ::接收方 | V: {接收方1:+1},{接收方2:+1}
key := fmt.Sprintf("im:chat:unread:%d", in.ToUserId)
field := fmt.Sprintf("%d", in.FromUserId)
_, err = l.svcCtx.Redis.Hincrby(key, field, 1)


# 待实现
群聊消息收发
Reids群组在线成员