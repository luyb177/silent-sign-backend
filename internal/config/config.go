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
}

type RedisConf struct {
	Addr     string
	Password string
	DB       int
}
