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

type ListFriendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// ListFriendsLogic 好友列表
func NewListFriendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFriendsLogic {
	return &ListFriendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFriendsLogic) ListFriends(req *types.ListFriendsReq) (resp *types.ListFriendsResp, err error) {
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

	// 3. 查询好友列表
	friends, err := l.svcCtx.Repos.Friend.ListByUser(l.ctx, authUser.UserID, limit, cursorID)
	if err != nil {
		l.Errorf("list friends failed: %v", err)
		return nil, errorx.WrapDBQuery("查询好友列表失败", err)
	}

	// 4. 判断 has_more
	hasMore := len(friends) > int(pageSize)
	if hasMore {
		friends = friends[:pageSize]
	}

	// 5. 批量查好友用户信息
	friendIDs := make([]uint64, 0, len(friends))
	for _, f := range friends {
		friendIDs = append(friendIDs, f.FriendID)
	}
	userMap, _ := l.svcCtx.Repos.User.FindByIDs(l.ctx, friendIDs)

	// 6. 组装响应
	friendInfos := make([]types.FriendInfo, 0, len(friends))
	for _, f := range friends {
		fi := types.FriendInfo{
			FriendID:  f.FriendID,
			CreatedAt: f.CreatedAt.Format(time.DateTime),
		}
		if u, ok := userMap[f.FriendID]; ok {
			fi.User = types.CreatorInfo{UserID: u.ID, Name: u.Name, Avatar: u.Avatar}
		}
		friendInfos = append(friendInfos, fi)
	}

	// 7. 下一页 token
	nextToken := ""
	if hasMore {
		last := friends[len(friends)-1]
		nextToken = strconv.FormatUint(last.ID, 10)
	}

	return &types.ListFriendsResp{
		Friends:   friendInfos,
		PageToken: nextToken,
		HasMore:   hasMore,
	}, nil
}
