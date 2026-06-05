// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package moment

import (
	"context"
	"time"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMomentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetMomentLogic 获取动态详情
func NewGetMomentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMomentLogic {
	return &GetMomentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMomentLogic) GetMoment(req *types.GetMomentReq) (resp *types.GetMomentResp, err error) {
	// 1. 查找动态
	m, err := l.svcCtx.Repos.Moment.FindByID(l.ctx, req.MomentID)
	if err != nil {
		l.Errorf("find moment failed: %v", err)
		return nil, errorx.WrapDBQuery("查找动态失败", err)
	}
	if m == nil {
		return nil, errorx.ErrNotFound
	}

	// 2. 查找创建者信息
	creator, err := l.svcCtx.Repos.User.FindByID(l.ctx, m.UserID)
	if err != nil {
		l.Errorf("find creator failed: %v", err)
		return nil, errorx.WrapDBQuery("查找用户失败", err)
	}

	// 3. 查找图片
	imageURLs, err := l.getImageURLs(m.ID)
	if err != nil {
		l.Errorf("find images failed: %v", err)
		// 图片加载失败不阻塞，继续返回
	}

	// 4. 检查当前用户是否已点赞
	isLiked := false
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser != nil {
		liked, err := l.svcCtx.Repos.Like.IsLiked(l.ctx, constvar.TargetTypeMoment, m.ID, authUser.UserID)
		if err != nil {
			l.Errorf("check is_liked failed: %v", err)
			// 非关键操作，忽略错误
		} else {
			isLiked = liked
		}
	}

	// 5. 组装响应
	momentInfo := types.MomentInfo{
		MomentID:   m.ID,
		CreatedAt:  m.CreatedAt.Format(time.DateTime),
		UpdatedAt:  m.UpdatedAt.Format(time.DateTime),
		Type:       m.Type,
		Content:    m.Content,
		VideoURL:   m.VideoURL,
		ImageURLs:  imageURLs,
		LikeNum:    m.LikeNum,
		CommentNum: m.CommentNum,
		ShareNum:   m.ShareNum,
		ViewNum:    m.ViewNum,
		IsLiked:    isLiked,
	}

	if creator != nil {
		momentInfo.Creator = types.CreatorInfo{
			UserID: creator.ID,
			Name:   creator.Name,
			Avatar: creator.Avatar,
		}
	}

	return &types.GetMomentResp{
		Moment: momentInfo,
	}, nil
}

func (l *GetMomentLogic) getImageURLs(momentID uint64) ([]string, error) {
	images, err := l.svcCtx.Repos.Image.FindByTarget(l.ctx, constvar.TargetTypeMoment, momentID)
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, len(images))
	for _, img := range images {
		urls = append(urls, img.URL)
	}
	return urls, nil
}
