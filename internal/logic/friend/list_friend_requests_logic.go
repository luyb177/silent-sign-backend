// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package friend

import (
	"context"
	"strconv"
	"time"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFriendRequestsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// ListFriendRequestsLogic 待处理申请列表
func NewListFriendRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFriendRequestsLogic {
	return &ListFriendRequestsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFriendRequestsLogic) ListFriendRequests(req *types.ListFriendRequestsReq) (resp *types.ListFriendRequestsResp, err error) {
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

	var cursorID uint64
	if req.PageToken != "" {
		cursorID, err = strconv.ParseUint(req.PageToken, 10, 64)
		if err != nil {
			return nil, errorx.WrapBadRequest("分页参数无效", err)
		}
	}

	// 3. 查询
	requests, err := l.svcCtx.Repos.FriendRequest.ListPendingToUser(l.ctx, authUser.UserID, limit, cursorID)
	if err != nil {
		l.Errorf("list friend requests failed: %v", err)
		return nil, errorx.WrapDBQuery("查询好友申请失败", err)
	}

	// 4. 判断 has_more
	hasMore := len(requests) > int(pageSize)
	if hasMore {
		requests = requests[:pageSize]
	}

	// 5. 批量查用户信息
	userIDs := make([]uint64, 0, len(requests)*2)
	for _, r := range requests {
		userIDs = append(userIDs, r.FromUserID, r.ToUserID)
	}
	userMap, _ := l.svcCtx.Repos.User.FindByIDs(l.ctx, userIDs)

	// 6. 组装响应
	requestInfos := make([]types.FriendRequestInfo, 0, len(requests))
	for _, r := range requests {
		ri := types.FriendRequestInfo{
			RequestID: r.ID,
			CreatedAt: r.CreatedAt.Format(time.DateTime),
			Status:    r.Status,
		}
		if u, ok := userMap[r.FromUserID]; ok {
			ri.FromUser = types.CreatorInfo{UserID: u.ID, Name: u.Name, Avatar: u.Avatar}
		}
		if u, ok := userMap[r.ToUserID]; ok {
			ri.ToUser = types.CreatorInfo{UserID: u.ID, Name: u.Name, Avatar: u.Avatar}
		}
		requestInfos = append(requestInfos, ri)
	}

	// 7. 下一页 token
	nextToken := ""
	if hasMore {
		last := requests[len(requests)-1]
		nextToken = strconv.FormatUint(last.ID, 10)
	}

	return &types.ListFriendRequestsResp{
		Requests:  requestInfos,
		PageToken: nextToken,
		HasMore:   hasMore,
	}, nil
}
