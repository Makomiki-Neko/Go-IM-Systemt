package svc

import (
	"IMM/common/models"
	"IMM/common/pkg"
	"IMM/rpc/chat/internal/config"
	"log"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config   config.Config
	Redis    *redis.Redis
	DB       *gorm.DB
	RabbitMQ *pkg.RabbitClient
}

func NewServiceContext(c config.Config) *ServiceContext {

	db, err := gorm.Open(mysql.Open(c.DB.DataSource), &gorm.Config{})
	if err != nil {
		log.Fatalf("DataBases Connect Failed, Err: %v", err)
	}
	err = db.AutoMigrate(&models.PrivateMessage{})
	err = db.AutoMigrate(&models.GroupMessage{})
	if err != nil {
		log.Fatalf("DataBases Table Init Failed, Err: %v", err)
	}

	rabbitClient, err := pkg.NewRabbitClient(c.RabbitMQ.DSN, c.RabbitMQ.Exchange)
	if err != nil {
		log.Fatalf("RabbitMQ初始化失败: %v", err)
	}

	return &ServiceContext{
		Config:   c,
		Redis:    redis.MustNewRedis(c.RedisConf),
		DB:       db,
		RabbitMQ: rabbitClient,
	}
}
