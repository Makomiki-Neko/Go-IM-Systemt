// pkg/snowflake/snowflake.go
package pkg

import (
	"errors"

	"github.com/bwmarrin/snowflake"
)

// 全局节点实例
var node *snowflake.Node

// Init 初始化雪花ID生成器
// machineID: 机器/实例编号，范围 0~1023，分布式部署时每个实例必须唯一
func Init(machineID int64) error {
	if machineID < 0 || machineID > 1023 {
		return errors.New("machineID must be between 0 and 1023")
	}

	var err error
	node, err = snowflake.NewNode(machineID)
	return err
}

// GenUint64 生成 uint64 类型的雪花ID
// 对应 MySQL bigint unsigned 类型，适合作为数据库主键
func GenUint64() uint64 {
	if node == nil {
		panic("snowflake node not initialized")
	}
	return uint64(node.Generate().Int64())
}

// GenInt64 生成 int64 类型的雪花ID
func GenInt64() int64 {
	if node == nil {
		panic("snowflake node not initialized")
	}
	return node.Generate().Int64()
}
