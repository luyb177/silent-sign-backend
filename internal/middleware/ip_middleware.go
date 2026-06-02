// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"context"
	"encoding/base64"
	"net"
	"net/http"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/service"
	"github.com/luyb177/silent-sign-backend/internal/config"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type IPMiddleware struct {
	logx.Logger
	ip2region *service.Ip2Region
}

func NewIPMiddleware(cfg config.IP2RegionConf) *IPMiddleware {
	l := logx.WithContext(context.Background())
	// 1. 创建 v4 配置
	v4Config, err := service.NewV4Config(service.VIndexCache, cfg.V4, 20)
	if err != nil {
		logx.Must(err)
	}

	// 2. 创建 v6 配置
	v6Config, err := service.NewV6Config(service.VIndexCache, cfg.V6, 20)
	if err != nil {
		logx.Must(err)
	}

	// 3. 创建 ip2region 实例
	ip2region, err := service.NewIp2Region(v4Config, v6Config)
	if err != nil {
		logx.Must(err)
	}

	return &IPMiddleware{
		Logger:    l,
		ip2region: ip2region,
	}
}

func (m *IPMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 获取客户端 IP 地址
		ip := getClientIP(r)

		var ipLocation *types.IPLocation
		if ip != "" && m.ip2region != nil {
			region, err := m.ip2region.Search(ip)
			if err != nil {
				m.Errorf("IP %s 地理位置查询失败: %s", ip, err.Error())
			} else if loc := parseIPRegion(region); loc != nil {
				loc.IP = ip
				ipLocation = loc
			}
		}

		// 2. 将 IP 位置信息写入 context
		ctx := context.WithValue(r.Context(), constvar.IPLocationKey, ipLocation)
		next(w, r.WithContext(ctx))
	}
}

func getClientIP(r *http.Request) string {
	// 1. X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// client, proxy1, proxy2
		ips := strings.Split(xff, ",")
		ip := strings.TrimSpace(ips[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// 2. X-Real-IP
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		if net.ParseIP(xrip) != nil {
			return xrip
		}
	}

	// 3. RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && net.ParseIP(host) != nil {
		return host
	}

	return ""
}

func parseIPRegion(region string) *types.IPLocation {
	// region 格式：国家|省份|城市|ISP|iso-alpha2-code(国家两字母简称)
	parts := strings.Split(region, "|")
	if len(parts) < 5 {
		return nil
	}

	clean := func(s string) string {
		if s == "0" {
			return "未知"
		}
		return s
	}

	return &types.IPLocation{
		Country:  encodeValue(clean(parts[0])),
		Province: encodeValue(clean(parts[1])),
		City:     encodeValue(clean(parts[2])),
		ISP:      encodeValue(clean(parts[3])),
		ISOCode:  encodeValue(clean(parts[4])),
	}
}

// encodeValue 使用 base64 编码字符串
func encodeValue(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// GetIPLocation 从 context 中获取 IP 地理位置信息，方便业务层调用
func GetIPLocation(ctx context.Context) *types.IPLocation {
	v := ctx.Value(constvar.IPLocationKey)
	if v == nil {
		return nil
	}
	loc, ok := v.(*types.IPLocation)
	if !ok {
		return nil
	}
	return loc
}
