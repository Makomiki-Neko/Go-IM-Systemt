package svc

import (
	"IMM/common/models"
	"IMM/rpc/user/internal/config"
	"log"

	goRedis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config  config.Config
	DB      *gorm.DB
	Redis   *redis.Redis
	GoRedis *goRedis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DB.DataSource), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&models.User{})
	err = db.AutoMigrate(&models.UserInfo{})
	if err != nil {
		log.Fatal(err)
	}

	GoRedis := goRedis.NewClient(&goRedis.Options{
		Addr:     c.RedisConf.Host,
		Password: c.RedisConf.Pass,
		DB:       0, // 默认 DB 0

	})

	//fmt.Printf("Config: %+v\n", c)
	//fmt.Printf("RedisConf: %+v\n", c.RedisConf)

	return &ServiceContext{
		Config:  c,
		DB:      db,
		Redis:   redis.MustNewRedis(c.RedisConf),
		GoRedis: GoRedis,
	}
}
