package like

import (
	"context"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ToggleLikeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewToggleLikeLogic 点赞/取消点赞
func NewToggleLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ToggleLikeLogic {
	return &ToggleLikeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ToggleLikeLogic) ToggleLike(req *types.ToggleLikeReq) (resp *types.ToggleLikeResp, err error) {
	// 1. 参数校验
	if req.TargetID == 0 {
		return nil, errorx.WrapBadRequest("目标ID不能为空", nil)
	}

	// 2. 从 token 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 3. 执行点赞/取消点赞（纯 Redis 操作）
	liked, count, err := l.svcCtx.Repos.Like.Toggle(l.ctx, req.TargetType, req.TargetID, authUser.UserID)
	if err != nil {
		l.Errorf("toggle like failed: %v", err)
		return nil, errorx.ErrInternalServer
	}

	return &types.ToggleLikeResp{
		Liked:   liked,
		LikeNum: uint64(count),
	}, nil
}
