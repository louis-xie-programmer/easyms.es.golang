package db

import (
	"context"
	"easyms-es/config"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

var EasyRedis *redis.Client

// InitRedis 初始化redis client
func InitRedis() {
	var (
		address  = config.GetSyncConfig("", "common.redis.address")
		password = config.GetSyncConfig("", "common.redis.password")
		dbNum    = config.GetSyncConfig_Type[int]("", "common.redis.db")
	)

	EasyRedis = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       dbNum,    // use default DB
	})
}

// ExistsKey 判断key是否存在
func ExistsKey(key string) bool {
	ctx := context.Background()
	res := EasyRedis.Exists(ctx, key).Val()
	return res == 1
}

// SetCache 设置缓存值
func SetCache(key string, value interface{}) error {
	ctx := context.Background()
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return EasyRedis.Set(ctx, key, val, 0).Err()
}

// SetHashCache 设置hash值
func SetHashCache(key string, values map[string]interface{}) error {
	ctx := context.Background()
	if ExistsKey(key) {
		err := EasyRedis.Del(ctx, key).Err()
		if err != nil {
			return err
		}
	}
	return EasyRedis.HSet(ctx, key, values).Err()
}

// RemoveParentKeyCache 删除父键中的所有子键
func RemoveParentKeyCache(parentKeyPattern string) error {
	ctx := context.Background()
	iter := EasyRedis.Scan(ctx, 0, parentKeyPattern, 0).Iterator()
	for iter.Next(ctx) {
		err := EasyRedis.Del(ctx, iter.Val()).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveKeyCache 删除
func RemoveKeyCache(key string) error {
	ctx := context.Background()
	return EasyRedis.Del(ctx, key).Err()
}

// GetCache 获取缓存值
func GetCache(key string) (string, error) {
	ctx := context.Background()
	return EasyRedis.Get(ctx, key).Result()
}

// GetHashCache 获取hash缓存值
func GetHashCache(key string, field string) (map[string]string, error) {
	ctx := context.Background()
	vals, err := EasyRedis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return vals, nil
}

// SetHashSetCache 在集合中插入数据
func SetHashSetCache(key string, value interface{}) error {
	ctx := context.Background()

	val, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = EasyRedis.SAdd(ctx, key, val).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetHashSetCache 获取集合值
func GetHashSetCache(key string, field string) ([]string, error) {
	ctx := context.Background()
	vals, err := EasyRedis.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return vals, nil
}
