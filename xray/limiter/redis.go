package limiter

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
	"x-ui/logger"

	"github.com/go-redis/redis/v8"
)

var (
	client     *redis.Client
	clientOnce sync.Once
	ctx        = context.Background()
)

// Config Redis配置
type Config struct {
	Enabled      bool   // 是否启用IP限制
	RedisAddr    string // Redis服务器地址
	RedisPort    int    // Redis端口
	RedisPass    string // Redis密码
	RedisDB      int    // Redis数据库索引
	MaxIPLimit   int    // 每个UUID最大IP数量限制
	MaxIpPerConn int    // 每个IP最大连接数量
}

// 全局配置实例
var config = &Config{
	Enabled:      false,
	RedisAddr:    "127.0.0.1",
	RedisPort:    6379,
	RedisPass:    "",
	RedisDB:      0,
	MaxIPLimit:   2,
	MaxIpPerConn: 10,
}

// InitRedisClient 初始化Redis客户端
func InitRedisClient() error {
	var initErr error
	clientOnce.Do(func() {
		if !config.Enabled {
			return
		}

		addr := fmt.Sprintf("%s:%d", config.RedisAddr, config.RedisPort)
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: config.RedisPass,
			DB:       config.RedisDB,
		})

		_, err := client.Ping(ctx).Result()
		if err != nil {
			logger.Warning("Redis连接失败:", err)
			initErr = err
			return
		}

		logger.Info("Redis连接成功，IP限制功能已启用")
	})
	return initErr
}

// GetConfig 获取当前配置
func GetConfig() *Config {
	return config
}

// SetConfig 设置配置
func SetConfig(cfg *Config) {
	config = cfg
}

// isEnabled 检查限制功能是否启用
func isEnabled() bool {
	return config.Enabled && client != nil
}

// CheckConnection 检查UUID的连接是否超过限制
func CheckConnection(uuid string, clientIP string) bool {
	if !isEnabled() {
		return true // 如果未启用，则总是允许连接
	}

	// 检查UUID的IP数量是否达到限制
	ipsKey := fmt.Sprintf("user:%s:ips", uuid)
	ipCountKey := fmt.Sprintf("user:%s:ipcount:%s", uuid, clientIP)

	// 检查当前IP数
	ips, err := client.SMembers(ctx, ipsKey).Result()
	if err != nil && err != redis.Nil {
		logger.Warning("Redis获取UUID的IP列表失败:", err)
		return true // 出错时放行
	}

	// 如果IP数达到限制且当前IP不在列表中，拒绝连接
	if len(ips) >= config.MaxIPLimit && !client.SIsMember(ctx, ipsKey, clientIP).Val() {
		logger.Info(fmt.Sprintf("UUID %s 已达到最大IP限制 %d，当前IP: %s 被拒绝", uuid, config.MaxIPLimit, clientIP))
		return false
	}

	// 检查单个IP的连接数是否超过限制
	ipConn, err := client.Get(ctx, ipCountKey).Result()
	if err != nil && err != redis.Nil {
		logger.Warning("Redis获取IP连接数失败:", err)
		return true
	}

	connCount := 0
	if err != redis.Nil {
		connCount, _ = strconv.Atoi(ipConn)
	}

	if connCount >= config.MaxIpPerConn {
		logger.Info(fmt.Sprintf("IP %s 已达到最大连接限制 %d", clientIP, config.MaxIpPerConn))
		return false
	}

	return true
}

// AddConnection 添加UUID的连接
func AddConnection(uuid string, clientIP string) {
	if !isEnabled() {
		return
	}

	ipsKey := fmt.Sprintf("user:%s:ips", uuid)
	ipCountKey := fmt.Sprintf("user:%s:ipcount:%s", uuid, clientIP)

	// 使用管道执行多个命令
	pipe := client.Pipeline()

	// 添加IP到集合
	pipe.SAdd(ctx, ipsKey, clientIP)

	// 增加IP连接计数
	pipe.Incr(ctx, ipCountKey)

	// 设置过期时间（24小时），避免过期的连接永远存在
	pipe.Expire(ctx, ipsKey, 24*time.Hour)
	pipe.Expire(ctx, ipCountKey, 24*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Warning("Redis添加连接失败:", err)
	}
}

// RemoveConnection 移除UUID的连接
func RemoveConnection(uuid string, clientIP string) {
	if !isEnabled() {
		return
	}

	ipsKey := fmt.Sprintf("user:%s:ips", uuid)
	ipCountKey := fmt.Sprintf("user:%s:ipcount:%s", uuid, clientIP)

	// 减少IP连接计数
	count, err := client.Decr(ctx, ipCountKey).Result()
	if err != nil {
		logger.Warning("Redis减少连接计数失败:", err)
		return
	}

	// 如果连接数为0，从IP集合中移除
	if count <= 0 {
		client.SRem(ctx, ipsKey, clientIP)
		client.Del(ctx, ipCountKey)
	}
}

// ClearAllConnections 清除所有连接记录
func ClearAllConnections() {
	if !isEnabled() {
		return
	}

	// 查找所有用户的键
	keys, err := client.Keys(ctx, "user:*").Result()
	if err != nil {
		logger.Warning("Redis查找所有连接记录失败:", err)
		return
	}

	if len(keys) > 0 {
		err = client.Del(ctx, keys...).Err()
		if err != nil {
			logger.Warning("Redis清除所有连接记录失败:", err)
		}
	}
}

// GetActiveIPs 获取UUID的活跃IP列表
func GetActiveIPs(uuid string) []string {
	if !isEnabled() {
		return []string{}
	}

	ipsKey := fmt.Sprintf("user:%s:ips", uuid)

	ips, err := client.SMembers(ctx, ipsKey).Result()
	if err != nil {
		logger.Warning("Redis获取活跃IP列表失败:", err)
		return []string{}
	}

	return ips
}
