// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package message

import (
	"context"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	messageRepo "github.com/luyb177/silent-sign-backend/internal/repo/message"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// SendMessageLogic 发送消息
func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMessageLogic) SendMessage(req *types.SendMessageReq) (resp *types.Response, err error) {
	// 1. 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 校验接收者存在
	receiver, err := l.svcCtx.Repos.User.FindByID(l.ctx, req.ReceiverID)
	if err != nil {
		return nil, errorx.WrapDBQuery("查询用户失败", err)
	}
	if receiver == nil {
		return nil, errorx.WrapBadRequest("接收者不存在", nil)
	}

	// 3. 创建消息
	m := &messageRepo.Message{
		SenderID:   authUser.UserID,
		ReceiverID: req.ReceiverID,
		Type:       req.Type,
		Content:    req.Content,
	}
	if err := l.svcCtx.Repos.Message.Create(l.ctx, m); err != nil {
		l.Errorf("send message failed: %v", err)
		return nil, errorx.WrapDBInsert("发送消息失败", err)
	}

	// todo 推送消息

	return &types.Response{}, nil
}
