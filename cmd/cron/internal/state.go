package internal

import "github.com/luyb177/silent-sign-backend/internal/svc"

// SvcCtx 共享的服务上下文，由 main.go 初始化，子命令使用
var SvcCtx *svc.ServiceContext
