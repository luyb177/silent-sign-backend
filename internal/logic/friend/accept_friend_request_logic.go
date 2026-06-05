// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package friend

import (
	"context"
	"encoding/json"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	friendRepo "github.com/luyb177/silent-sign-backend/internal/repo/friend"
	messageRepo "github.com/luyb177/silent-sign-backend/internal/repo/message"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AcceptFriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// AcceptFriendRequestLogic 通过好友申请
func NewAcceptFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptFriendRequestLogic {
	return &AcceptFriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AcceptFriendRequestLogic) AcceptFriendRequest(req *types.HandleFriendRequestReq) (resp *types.Response, err error) {
	// 1. 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 查找申请
	fr, err := l.svcCtx.Repos.FriendRequest.FindByID(l.ctx, req.RequestID)
	if err != nil {
		return nil, errorx.WrapDBQuery("查询申请失败", err)
	}
	if fr == nil {
		return nil, errorx.WrapBadRequest("申请不存在", nil)
	}
	if fr.ToUserID != authUser.UserID {
		return nil, errorx.ErrForbidden
	}
	if fr.Status != constvar.FriendRequestStatusPending {
		return nil, errorx.WrapBadRequest("申请已处理", nil)
	}

	// 3. 事务内：更新申请状态 + 建立双向好友关系 + 发消息通知
	err = l.svcCtx.Repos.Transaction(func(tx *gorm.DB) error {
		if err := l.svcCtx.Repos.FriendRequest.UpdateStatus(l.ctx, fr.ID, constvar.FriendRequestStatusAccepted, tx); err != nil {
			return err
		}
		// 双向插入好友关系
		if err := l.svcCtx.Repos.Friend.Create(l.ctx, &friendRepo.Friend{
			UserID: fr.FromUserID, FriendID: fr.ToUserID,
		}, tx); err != nil {
			return err
		}
		if err := l.svcCtx.Repos.Friend.Create(l.ctx, &friendRepo.Friend{
			UserID: fr.ToUserID, FriendID: fr.FromUserID,
		}, tx); err != nil {
			return err
		}
		msg := &messageRepo.Message{
			SenderID:   authUser.UserID,
			ReceiverID: fr.FromUserID,
			Type:       constvar.MsgTypeFriendAccept,
			Content:    "已通过你的好友申请",
		}
		return l.svcCtx.Repos.Message.Create(l.ctx, msg, tx)
	})
	if err != nil {
		l.Errorf("accept friend request failed: %v", err)
		return nil, errorx.WrapDBUpdate("通过好友申请失败", err)
	}

	// 4. SSE 推送
	event, _ := json.Marshal(map[string]interface{}{
		"type":     "new_message",
		"msg_type": constvar.MsgTypeFriendAccept,
	})
	l.svcCtx.SSEHub.PushToUser(fr.FromUserID, event)

	return &types.Response{}, nil
}
