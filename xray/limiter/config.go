package limiter

import (
	"database/sql"
	"fmt"
	"strconv"
	"x-ui/logger"
)

// LoadIPLimitConfig 从数据库加载IP限制配置
func LoadIPLimitConfig() error {
	db, err := sql.Open("sqlite3", "bin/x-ui.db")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// 检查是否启用IP限制
	enabledStr, err := getSettingValue(db, "enableIpLimit")
	if err != nil {
		// 默认不启用
		config.Enabled = false
		return nil
	}

	// 转换enableIpLimit的值
	if enabledStr == "true" {
		config.Enabled = true
	} else {
		config.Enabled = false
		// 如果未启用，不需要加载其他配置
		return nil
	}

	// 加载Redis服务器地址
	redisAddr, err := getSettingValue(db, "redisAddr")
	if err == nil && redisAddr != "" {
		config.RedisAddr = redisAddr
	}

	// 加载Redis端口
	redisPort, err := getSettingValue(db, "redisPort")
	if err == nil && redisPort != "" {
		port, err := strconv.Atoi(redisPort)
		if err == nil && port > 0 && port < 65536 {
			config.RedisPort = port
		}
	}

	// 加载Redis密码
	redisPassword, err := getSettingValue(db, "redisPassword")
	if err == nil {
		config.RedisPass = redisPassword
	}

	// 加载Redis数据库索引
	redisDb, err := getSettingValue(db, "redisDb")
	if err == nil && redisDb != "" {
		db, err := strconv.Atoi(redisDb)
		if err == nil && db >= 0 {
			config.RedisDB = db
		}
	}

	// 加载最大IP限制
	maxIpLimit, err := getSettingValue(db, "maxIpLimit")
	if err == nil && maxIpLimit != "" {
		limit, err := strconv.Atoi(maxIpLimit)
		if err == nil && limit > 0 {
			config.MaxIPLimit = limit
		}
	}

	// 加载单IP最大连接数
	maxIpPerConn, err := getSettingValue(db, "maxIpPerConn")
	if err == nil && maxIpPerConn != "" {
		limit, err := strconv.Atoi(maxIpPerConn)
		if err == nil && limit > 0 {
			config.MaxIpPerConn = limit
		}
	}

	logger.Info("IP限制配置已加载, 已启用:", config.Enabled,
		", Redis地址:", config.RedisAddr,
		", Redis端口:", config.RedisPort,
		", 最大IP数量:", config.MaxIPLimit,
		", 单IP最大连接数:", config.MaxIpPerConn)

	return nil
}

// getSettingValue 从数据库获取设置值
func getSettingValue(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM setting WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}
