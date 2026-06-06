// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"time"

	"github.com/luyb177/silent-sign-backend/common/cache"
	"github.com/luyb177/silent-sign-backend/common/database"
	"github.com/luyb177/silent-sign-backend/common/jwtx"
	"github.com/luyb177/silent-sign-backend/common/mail"
	"github.com/luyb177/silent-sign-backend/internal/config"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/pkg/email"
	"github.com/luyb177/silent-sign-backend/internal/repo"
	"github.com/luyb177/silent-sign-backend/internal/sse"
	"github.com/luyb177/silent-sign-backend/internal/ws"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config        config.Config
	Mailer        *mail.Mailer
	RedisClient   *cache.RedisClient
	MySQLClient   *database.MySQLClient
	EmailSender   email.EmailSender
	Repos         *repo.Repositories
	SSEHub        *sse.Hub
	WSHub         *ws.Hub
	JWTHandler    jwtx.Handler
	IPMiddleware  rest.Middleware
	JWTMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	m := mail.NewMailer(mail.EmailConfig{
		From:     c.EmailConf.From,
		Password: c.EmailConf.Password,
		SMTPHost: c.EmailConf.SMTPHost,
		SMTPPort: c.EmailConf.SMTPPort,
	})

	emailSender := email.NewEmailSender(m)

	rc, err := cache.NewRedisClient(c.RedisConf.Addr, c.RedisConf.Password, c.RedisConf.DB)
	if err != nil {
		panic(err)
	}

	mc, err := database.NewMySQLClient(c.MySQLConf.DSN)
	if err != nil {
		panic(err)
	}

	jwtHandler, err := jwtx.NewHandler(
		c.JWTConf.Secret,
		time.Duration(c.JWTConf.ExpireS)*time.Second,
		time.Duration(c.JWTConf.RefreshExpireS)*time.Second,
	)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:        c,
		Mailer:        m,
		RedisClient:   rc,
		MySQLClient:   mc,
		EmailSender:   emailSender,
		Repos:         repo.NewRepositories(rc.Client, mc.DB),
		SSEHub:        sse.NewHub(),
		WSHub:         ws.NewHub(),
		JWTHandler:    jwtHandler,
		IPMiddleware:  middleware.NewIPMiddleware(c.IP2RegionConf).Handle,
		JWTMiddleware: middleware.NewJWTMiddleware(jwtHandler).Handle,
	}
}
