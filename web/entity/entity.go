package entity

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"strings"
	"time"
	"x-ui/util/common"
	"x-ui/xray"
)

type Msg struct {
	Success bool        `json:"success"`
	Msg     string      `json:"msg"`
	Obj     interface{} `json:"obj"`
}

type Pager struct {
	Current  int         `json:"current"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
	OrderBy  string      `json:"order_by"`
	Desc     bool        `json:"desc"`
	Key      string      `json:"key"`
	List     interface{} `json:"list"`
}

type AllSetting struct {
	WebListen   string `json:"webListen" form:"webListen"`
	WebPort     int    `json:"webPort" form:"webPort"`
	WebCertFile string `json:"webCertFile" form:"webCertFile"`
	WebKeyFile  string `json:"webKeyFile" form:"webKeyFile"`
	WebBasePath string `json:"webBasePath" form:"webBasePath"`

	XrayTemplateConfig string `json:"xrayTemplateConfig" form:"xrayTemplateConfig"`

	TimeLocation string `json:"timeLocation" form:"timeLocation"`

	// IP限制相关设置
	EnableIpLimit bool   `json:"enableIpLimit" form:"enableIpLimit"`
	RedisAddr     string `json:"redisAddr" form:"redisAddr"`
	RedisPort     int    `json:"redisPort" form:"redisPort"`
	RedisPassword string `json:"redisPassword" form:"redisPassword"`
	RedisDb       int    `json:"redisDb" form:"redisDb"`
	MaxIpLimit    int    `json:"maxIpLimit" form:"maxIpLimit"`
	MaxIpPerConn  int    `json:"maxIpPerConn" form:"maxIpPerConn"`
}

func (s *AllSetting) CheckValid() error {
	if s.WebListen != "" {
		ip := net.ParseIP(s.WebListen)
		if ip == nil {
			return common.NewError("web listen is not valid ip:", s.WebListen)
		}
	}

	if s.WebPort <= 0 || s.WebPort > 65535 {
		return common.NewError("web port is not a valid port:", s.WebPort)
	}

	if s.WebCertFile != "" || s.WebKeyFile != "" {
		_, err := tls.LoadX509KeyPair(s.WebCertFile, s.WebKeyFile)
		if err != nil {
			return common.NewErrorf("cert file <%v> or key file <%v> invalid: %v", s.WebCertFile, s.WebKeyFile, err)
		}
	}

	if !strings.HasPrefix(s.WebBasePath, "/") {
		s.WebBasePath = "/" + s.WebBasePath
	}
	if !strings.HasSuffix(s.WebBasePath, "/") {
		s.WebBasePath += "/"
	}

	xrayConfig := &xray.Config{}
	err := json.Unmarshal([]byte(s.XrayTemplateConfig), xrayConfig)
	if err != nil {
		return common.NewError("xray template config invalid:", err)
	}

	_, err = time.LoadLocation(s.TimeLocation)
	if err != nil {
		return common.NewError("time location not exist:", s.TimeLocation)
	}

	// 验证IP限制相关设置
	if s.EnableIpLimit {
		if s.RedisAddr == "" {
			return common.NewError("Redis address cannot be empty when IP limit is enabled")
		}
		if s.RedisPort <= 0 || s.RedisPort > 65535 {
			return common.NewError("Redis port is not valid:", s.RedisPort)
		}
		if s.MaxIpLimit <= 0 {
			return common.NewError("Max IP limit must be greater than 0")
		}
		if s.MaxIpPerConn <= 0 {
			return common.NewError("Max IP per connection must be greater than 0")
		}
	}

	return nil
}
