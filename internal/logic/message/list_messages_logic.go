// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package message

import (
	"context"
	"time"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/pkg/pagetoken"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// ListMessagesLogic 消息列表
func NewListMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMessagesLogic {
	return &ListMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMessagesLogic) ListMessages(req *types.ListMessagesReq) (resp *types.ListMessagesResp, err error) {
	// 1. 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 2. 参数校验
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 50 {
		pageSize = constvar.DefaultPageSize
	}
	limit := int(pageSize) + 1

	// 3. 解析 page_token
	var pt types.MessagePageToken
	if req.PageToken != "" {
		if err := pagetoken.Decode(req.PageToken, constvar.MessagePageTokenPrefix, &pt); err != nil {
			return nil, errorx.WrapBadRequest("分页参数无效", err)
		}
	}

	// 4. 查询消息列表
	messages, err := l.svcCtx.Repos.Message.ListByReceiver(l.ctx, authUser.UserID, pt.ID, limit)
	if err != nil {
		l.Errorf("list messages failed: %v", err)
		return nil, errorx.WrapDBQuery("查询消息列表失败", err)
	}

	// 5. 判断 has_more
	hasMore := len(messages) > int(pageSize)
	if hasMore {
		messages = messages[:pageSize]
	}

	// 6. 批量查发送者信息
	senderIDs := make([]uint64, 0, len(messages))
	for _, m := range messages {
		if m.SenderID != 0 {
			senderIDs = append(senderIDs, m.SenderID)
		}
	}
	senderMap, _ := l.svcCtx.Repos.User.FindByIDs(l.ctx, senderIDs)

	// 7. 组装响应
	messageInfos := make([]types.MessageInfo, 0, len(messages))
	for _, m := range messages {
		mi := types.MessageInfo{
			MessageID: m.ID,
			CreatedAt: m.CreatedAt.Format(time.DateTime),
			Type:      m.Type,
			Content:   m.Content,
			IsRead:    m.IsRead,
		}
		if m.SenderID == 0 {
			mi.Creator = types.CreatorInfo{
				UserID: 0,
				Name:   "系统",
			}
		} else if u, ok := senderMap[m.SenderID]; ok {
			mi.Creator = types.CreatorInfo{
				UserID: u.ID,
				Name:   u.Name,
				Avatar: u.Avatar,
			}
		}
		messageInfos = append(messageInfos, mi)
	}

	// 8. 生成下一页 token
	nextToken := ""
	if hasMore {
		last := messages[len(messages)-1]
		nextPT := types.MessagePageToken{ID: last.ID}
		nextToken, _ = pagetoken.Encode(constvar.MessagePageTokenPrefix, &nextPT)
	}

	return &types.ListMessagesResp{
		Messages:  messageInfos,
		PageToken: nextToken,
		HasMore:   hasMore,
	}, nil
}
