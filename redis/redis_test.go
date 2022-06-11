package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/pokitpeng/pkg/logger"
)

const (
	redisAddr     = "192.168.50.97:30246"
	redisPassword = ""
	db            = 0
	PoolSize      = 200
	MaxRetries    = 3
	IdleTimeout   = 30
	MinIdleConns  = 5
)

// Client redis 连接池
var (
	ctx    = context.Background()
	Client *redis.Client
)

// 初始化redis链接池
func init() {
	Client = redis.NewClient(&redis.Options{
		Addr:         redisAddr,                                // Redis地址
		Password:     redisPassword,                            // Redis账号
		DB:           db,                                       // Redis库
		PoolSize:     PoolSize,                                 // Redis连接池大小
		MaxRetries:   MaxRetries,                               // 最大重试次数
		IdleTimeout:  time.Duration(IdleTimeout) * time.Second, // 空闲链接超时时间
		MinIdleConns: MinIdleConns,                             // 空闲连接数量
	})
	pong, err := Client.Ping(ctx).Result()
	if err == redis.Nil {
		log.Fatalf("Redis异常")
	} else if err != nil {
		log.Fatalf("失败:%s", err.Error())
	} else {
		log.Info(pong)
	}

}

// go test --count=1 -run TestString
func TestString(t *testing.T) { // 设置k,v
	err := Client.Set(ctx, "name", "tom", 0).Err()
	if err != nil {
		log.Error(err)
		return
	}

	val, err := Client.Get(ctx, "name").Result()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("name=", val)

	val2, err := Client.Get(ctx, "missing_key").Result()
	if err == redis.Nil {
		log.Warn("missing_key does not exist")
	} else if err != nil {
		log.Error(err)
		return
	} else {
		log.Info("missing_key=", val2)
	}
}

// go test --count=1 -run TestStringUpdate
func TestStringUpdate(t *testing.T) { // 更新k,v
	err := Client.Set(ctx, "name", "tom", 0).Err()
	if err != nil {
		log.Error(err)
		return
	}

	val, err := Client.Get(ctx, "name").Result()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("before update name=", val)

	// update
	err = Client.Set(ctx, "name", "jerry", 0).Err()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("updating...")

	val2, err := Client.Get(ctx, "name").Result()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("after update name=", val2)
}

// go test --count=1 -run TestStringExpiration
func TestStringExpiration(t *testing.T) { // k,v过期时间
	err := Client.Set(ctx, "name", "tom", 2*time.Second).Err()
	if err != nil {
		log.Error(err)
		return
	}

	val, err := Client.Get(ctx, "name").Result()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("before expiration name=", val)

	// sleep
	log.Info("sleep 3s ...")
	time.Sleep(3 * time.Second)

	val2, err := Client.Get(ctx, "name").Result()
	if err == redis.Nil {
		log.Warn("after expiration name does not exist")
	} else if err != nil {
		log.Error(err)
		return
	} else {
		log.Info("after expiration name=", val2)
	}
}

// go test --count=1 -run TestHash
func TestHash(t *testing.T) { // 设置hash
	if err := Client.HSet(ctx, "myhash1", "key1", "value1", "key2", "value2").Err(); err != nil {
		log.Error(err)
		return
	}
	if err := Client.HSet(ctx, "myhash2", []string{"key1", "value1", "key2", "value2"}).Err(); err != nil {
		log.Error(err)
		return
	}
	if err := Client.HSet(ctx, "myhash3", map[string]interface{}{"key1": "value1", "key2": "value2"}).Err(); err != nil {
		log.Error(err)
		return
	}

	myhash1, err := Client.HGet(ctx, "myhash1", "key1").Result()
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("get hash myhash1[key1]=%s", myhash1)

	myhash1all, err := Client.HGetAll(ctx, "myhash1").Result()
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("get hash myhash1all=%v", myhash1all)

	myhash2all, err := Client.HGetAll(ctx, "myhash2").Result()
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("get hash myhash2all=%v", myhash2all)

	myhash3all, err := Client.HGetAll(ctx, "myhash3").Result()
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("get hash myhash3all=%v", myhash3all)
}

// go test --count=1 -run TestList
func TestList(t *testing.T) { // 设置list
	if err := Client.LPush(ctx, "mylist", []string{"hello1", "hello2", "hello3"}).Err(); err != nil {
		log.Error(err)
		return
	}
	result, err := Client.LRange(ctx, "mylist", -1, -1).Result()
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("get list mylist=%v", result)
}
