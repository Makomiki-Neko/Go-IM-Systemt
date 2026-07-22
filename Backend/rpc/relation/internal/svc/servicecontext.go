package svc

import (
	"IMM/common/models"
	"IMM/common/pkg"
	"IMM/rpc/relation/internal/config"
	"log"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config   config.Config
	DB       *gorm.DB
	Redis    *redis.Redis
	RabbitMQ *pkg.RabbitClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DB.DataSource), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&models.UserFriend{})
	err = db.AutoMigrate(&models.GroupInfo{})
	err = db.AutoMigrate(&models.GroupMember{})
	if err != nil {
		log.Fatalf("DataBases Init Failed, Err: %v", err)
	}

	rabbitClient, err := pkg.NewRabbitClient(c.RabbitMQ.DSN, c.RabbitMQ.Exchange)
	if err != nil {
		log.Fatalf("RabbitMQ初始化失败: %v", err)
	}

	return &ServiceContext{
		Config:   c,
		DB:       db,
		Redis:    redis.MustNewRedis(c.RedisConf),
		RabbitMQ: rabbitClient,
	}
}
