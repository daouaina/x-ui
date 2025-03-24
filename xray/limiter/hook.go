package limiter

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"x-ui/logger"
)

var (
	// 用于存储UUID到协议的映射
	uuidMap      = make(map[string]string)
	uuidMapMutex sync.RWMutex
)

// ExtractUUIDs 从配置中提取所有的UUID和对应协议
func ExtractUUIDs(config []byte) {
	var rawConfig interface{}
	err := json.Unmarshal(config, &rawConfig)
	if err != nil {
		logger.Warning("解析Xray配置失败:", err)
		return
	}

	// 清空旧的映射
	uuidMapMutex.Lock()
	uuidMap = make(map[string]string)
	uuidMapMutex.Unlock()

	// 提取入站配置
	configMap, ok := rawConfig.(map[string]interface{})
	if !ok {
		logger.Warning("Xray配置格式错误")
		return
	}

	inbounds, ok := configMap["inbounds"].([]interface{})
	if !ok {
		logger.Warning("Xray配置中没有找到入站配置")
		return
	}

	// 遍历所有入站
	for _, inbound := range inbounds {
		inboundMap, ok := inbound.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取协议类型
		protocol, ok := inboundMap["protocol"].(string)
		if !ok {
			continue
		}

		// 获取Tag
		tag, _ := inboundMap["tag"].(string)

		// 只处理vmess, vless, trojan协议
		if protocol != "vmess" && protocol != "vless" && protocol != "trojan" {
			continue
		}

		// 获取设置
		settings, ok := inboundMap["settings"].(map[string]interface{})
		if !ok {
			continue
		}

		// 获取客户端列表
		var clients []interface{}
		if protocol == "trojan" {
			clients, _ = settings["clients"].([]interface{})
		} else {
			clients, _ = settings["clients"].([]interface{})
		}

		// 提取UUID
		for _, client := range clients {
			clientMap, ok := client.(map[string]interface{})
			if !ok {
				continue
			}

			var uuid string
			if protocol == "trojan" {
				uuid, _ = clientMap["password"].(string)
			} else {
				uuid, _ = clientMap["id"].(string)
			}

			if uuid != "" {
				uuidMapMutex.Lock()
				uuidMap[uuid] = protocol
				uuidMapMutex.Unlock()
				logger.Debug("提取到", protocol, "协议的UUID:", uuid, "Tag:", tag)
			}
		}
	}
}

// Regexp for extracting connection info from Xray logs
var (
	acceptRegex = regexp.MustCompile(`accept a ([a-zA-Z0-9\-]+) connection from (.+?):(\d+)`)
	uuidRegex   = regexp.MustCompile(`identifier: ([a-zA-Z0-9\-]+)`)
)

// ProcessLog 处理Xray日志，识别连接建立和断开事件
func ProcessLog(line string) {
	// 检查是否启用了IP限制功能
	if !isEnabled() {
		return
	}

	// 处理连接建立
	if strings.Contains(line, "accept") {
		matches := acceptRegex.FindStringSubmatch(line)
		if len(matches) == 4 {
			// 获取连接信息
			protocol := matches[1]
			clientIP := matches[2]
			logger.Debug("检测到连接：", protocol, "从", clientIP)

			// 延迟一些特殊处理，因为UUID信息会在后续日志出现
		}
	}

	// 处理UUID识别
	if strings.Contains(line, "identifier:") {
		matches := uuidRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			uuid := matches[1]

			// 查找最后一次连接信息
			acceptMatches := acceptRegex.FindStringSubmatch(line)
			if len(acceptMatches) == 4 {
				clientIP := acceptMatches[2]

				// 检查是否允许连接
				if !CheckConnection(uuid, clientIP) {
					logger.Warning("拒绝UUID:", uuid, "客户端IP:", clientIP, "连接，已达到IP数量限制")
					// 这里我们无法直接中断连接，但会在后续的流量传输中限制它
					return
				}

				// 添加连接记录
				AddConnection(uuid, clientIP)
				logger.Info("接受UUID:", uuid, "客户端IP:", clientIP, "的连接")
			}
		}
	}

	// 处理连接断开（比较复杂，Xray日志可能没有明确的断开标志）
	if strings.Contains(line, "connection closed") {
		// 实际应用中，可能需要结合上下文分析或定期清理过期连接
	}
}

// IsUUIDRegistered 检查UUID是否在配置中注册
func IsUUIDRegistered(uuid string) bool {
	uuidMapMutex.RLock()
	defer uuidMapMutex.RUnlock()
	_, exists := uuidMap[uuid]
	return exists
}

// GetUUIDProtocol 获取UUID对应的协议
func GetUUIDProtocol(uuid string) string {
	uuidMapMutex.RLock()
	defer uuidMapMutex.RUnlock()
	return uuidMap[uuid]
}
