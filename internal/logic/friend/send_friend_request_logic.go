// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package friend

import (
	"context"
	"encoding/json"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	friendrequestRepo "github.com/luyb177/silent-sign-backend/internal/repo/friend_request"
	messageRepo "github.com/luyb177/silent-sign-backend/internal/repo/message"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SendFriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// SendFriendRequestLogic 发送好友申请
func NewSendFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendFriendRequestLogic {
	return &SendFriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendFriendRequestLogic) SendFriendRequest(req *types.SendFriendRequestReq) (resp *types.IDResponse, err error) {
	// 1. 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 不能加自己
	if authUser.UserID == req.ToUserID {
		return nil, errorx.WrapBadRequest("不能添加自己为好友", nil)
	}

	// 3. 校验目标用户存在
	target, err := l.svcCtx.Repos.User.FindByID(l.ctx, req.ToUserID)
	if err != nil {
		return nil, errorx.WrapDBQuery("查询用户失败", err)
	}
	if target == nil {
		return nil, errorx.WrapBadRequest("用户不存在", nil)
	}

	// 4. 检查是否已是好友
	isFriend, err := l.svcCtx.Repos.Friend.Exists(l.ctx, authUser.UserID, req.ToUserID)
	if err != nil {
		return nil, errorx.WrapDBQuery("查询好友关系失败", err)
	}
	if isFriend {
		return nil, errorx.WrapBadRequest("已是好友", nil)
	}

	// 5. 检查是否有待处理的申请
	existing, err := l.svcCtx.Repos.FriendRequest.FindPending(l.ctx, authUser.UserID, req.ToUserID)
	if err != nil {
		return nil, errorx.WrapDBQuery("查询申请记录失败", err)
	}
	if existing != nil {
		return nil, errorx.WrapBadRequest("已有待处理的好友申请", nil)
	}

	// 6. 事务内：创建申请 + 发消息通知
	fr := &friendrequestRepo.FriendRequest{
		FromUserID: authUser.UserID,
		ToUserID:   req.ToUserID,
		Status:     constvar.FriendRequestStatusPending,
	}

	err = l.svcCtx.Repos.Transaction(func(tx *gorm.DB) error {
		if err := l.svcCtx.Repos.FriendRequest.Create(l.ctx, fr, tx); err != nil {
			return err
		}

		msg := &messageRepo.Message{
			SenderID:   authUser.UserID,
			ReceiverID: req.ToUserID,
			Type:       constvar.MsgTypeFriendRequest,
			Content:    "请求添加你为好友",
		}
		return l.svcCtx.Repos.Message.Create(l.ctx, msg, tx)
	})
	if err != nil {
		l.Errorf("send friend request failed: %v", err)
		return nil, errorx.WrapDBInsert("发送好友申请失败", err)
	}

	// 7. SSE 推送
	event, _ := json.Marshal(map[string]interface{}{
		"type":     "new_message",
		"msg_type": constvar.MsgTypeFriendRequest,
	})
	l.svcCtx.SSEHub.PushToUser(req.ToUserID, event)

	return &types.IDResponse{
		ID: fr.ID,
	}, nil
}
