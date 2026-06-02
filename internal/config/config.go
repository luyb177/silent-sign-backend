// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	MySQLConf struct {
		DSN string
	}
	RedisConf RedisConf
	EmailConf struct {
		From     string
		Password string
		SMTPHost string
		SMTPPort int
	}
	IP2RegionConf IP2RegionConf
	JWTConf       JWTConf
}

type JWTConf struct {
	Secret  string
	ExpireS int64 // 过期时间，单位：秒
}

type RedisConf struct {
	Addr     string
	Password string
	DB       int
}

type IP2RegionConf struct {
	V4 string
	V6 string
}
