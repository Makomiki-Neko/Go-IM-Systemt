package svc

import (
	"IMM/common/models"
	"IMM/common/pkg"
	"IMM/rpc/ai/internal/config"
	"log"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config   config.Config
	LLM      *OpenAIClient
	DB       *gorm.DB
	RabbitMQ *pkg.RabbitClient
	Redis    *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Create LLM Client
	llm := NewOpenAIClient(c.LLM.Key, c.LLM.Api)

	db, err := gorm.Open(mysql.Open(c.DB.DataSource), &gorm.Config{})
	if err != nil {
		log.Fatalf("DataBases Connect Failed, Err: %v", err)
	}
	err = db.AutoMigrate(&models.LlmMessage{})
	err = db.AutoMigrate(&models.AgentInfo{})
	err = db.AutoMigrate(&models.LlmSession{})
	err = db.AutoMigrate(&models.SessionSummary{})
	if err != nil {
		log.Fatalf("DataBases Table Init Failed, Err: %v", err)
	}

	rabbitClient, err := pkg.NewRabbitClient(c.RabbitMQ.DSN, c.RabbitMQ.Exchange)
	if err != nil {
		log.Fatalf("RabbitMQ初始化失败: %v", err)
	}

	return &ServiceContext{
		Config:   c,
		LLM:      llm,
		DB:       db,
		Redis:    redis.MustNewRedis(c.RedisConf),
		RabbitMQ: rabbitClient,
	}
}
