// Redis事件过期监听
package redisListen

import (
	"IMM/common/models"
	"IMM/rpc/user/internal/svc"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// StartRedisExpiredListener 启动 Redis 过期事件监听
func StartRedisExpiredListener(svcCtx *svc.ServiceContext) {
	ctx := context.Background()
	// 订阅 Redis 的过期事件频道（假设使用 DB 0）
	err := svcCtx.GoRedis.Do(ctx, "CONFIG", "SET", "notify-keyspace-events", "Ex").Err()
	if err != nil {
		logx.Errorf("Set Redis Expired Notify Failed, %v", err)
		return
	}
	pubsub := svcCtx.GoRedis.Subscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	logx.Info("Redis expired event listener started.")

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			logx.Errorf("Receive expired event error: %v", err)
			// 可以加入重试逻辑或休眠后继续
			time.Sleep(time.Second)
			continue
		}

		// 过滤出 user:status: 前缀的 key
		key := msg.Payload
		if strings.HasPrefix(key, "user:status:") {
			account := strings.TrimPrefix(key, "user:status:")
			// 异步处理，避免阻塞事件循环
			go handleExpiredStatus(svcCtx, account)
		}
	}
}

// handleExpiredStatus 处理用户状态过期事件（含分布式锁）
func handleExpiredStatus(svcCtx *svc.ServiceContext, account string) {
	ctx := context.Background()
	lockKey := fmt.Sprintf("lock:write_last_online:%s", account)

	// 1. 尝试获取分布式锁（5秒过期），防止多实例重复写入
	ok, err := svcCtx.Redis.SetnxCtx(ctx, lockKey, "1")
	if err != nil || !ok {
		// 获取锁失败，说明其他实例正在处理
		return
	}
	// 为锁设置过期时间
	if err := svcCtx.Redis.ExpireCtx(ctx, lockKey, 5); err != nil {
		logx.Errorf("Set lock expire error: %v", err)
	}
	defer svcCtx.Redis.DelCtx(ctx, lockKey)

	// 2. 二次确认：检查用户是否真的已经离线（statusKey 是否已不存在）
	statusKey := fmt.Sprintf("user:status:%s", account)
	exists, err := svcCtx.Redis.ExistsCtx(ctx, statusKey)
	if err == nil && exists {
		// 用户可能又重新上线了，不记录离线时间
		logx.Infof("User %s is online again, skip write last online time.", account)
		return
	}

	// 3. 写入数据库（记录最后在线时间）
	t := time.Now()
	_, err = gorm.G[models.UserInfo](svcCtx.DB).Where("user_id = (SELECT uid FROM users WHERE account = ?)", account).Updates(ctx, models.UserInfo{LastOnline: &t})
	if err != nil {
		logx.Errorf("更新最后在线时间失败: %v", err)
	}
	if err != nil {
		logx.Errorf("Write last online time for user %s failed: %v", account, err)
	} else {
		logx.Infof("Recorded last online time for user %s", account)
	}
}
