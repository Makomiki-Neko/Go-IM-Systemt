// Redis
package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Message struct {
	Type      string          `json:"type"` // 对应协议中的 type
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

type Storage interface {
	Save(userId uint64, msg *Message) error
	Fetch(userId uint64, infoType string, limit int) ([]*Message, error)
	Clear(userId uint64, infoType string) error // 拉取后清理
}

const (
	redisKeyPrefix  = "offline:%s:%d" // 消息类型（event、chat）：Uid
	maxOfflineCount = 1000            // 防止恶意刷爆内存，最多存1000条
	expireTime      = 3 * 24 * time.Hour
)

type RedisStorage struct {
	Rdb *redis.Client
	Ctx context.Context
}

func (s *RedisStorage) Save(userId uint64, msg *Message) error {
	key := fmt.Sprintf(redisKeyPrefix, msg.Type, userId)
	data, _ := json.Marshal(msg)
	// LPush 将新消息插入头部，这样 LTrim 保留最近 N 条，List形式
	pipe := s.Rdb.Pipeline()
	pipe.LPush(s.Ctx, key, data)
	pipe.LTrim(s.Ctx, key, 0, maxOfflineCount-1) // 只保留最近 1000 条
	pipe.Expire(s.Ctx, key, expireTime)          // 无人拉取自动过期
	_, err := pipe.Exec(s.Ctx)
	return err
}

func (s *RedisStorage) Fetch(userId uint64, infoType string, limit int) ([]*Message, error) {
	key := fmt.Sprintf(redisKeyPrefix, infoType, userId)
	// 从头部开始拉取最新的 limit 条（按时间倒序，需反转）
	results, err := s.Rdb.LRange(s.Ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	var msgs []*Message
	for _, item := range results {
		var msg Message
		if err := json.Unmarshal([]byte(item), &msg); err == nil {
			msgs = append(msgs, &msg)
		}
	}
	// 反转顺序，使最早的放在前面（符合聊天时序）
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (s *RedisStorage) Clear(userId uint64, infoType string) error {
	key := fmt.Sprintf(redisKeyPrefix, infoType, userId)
	return s.Rdb.Del(s.Ctx, key).Err()
}
