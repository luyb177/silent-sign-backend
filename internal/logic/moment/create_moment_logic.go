// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package moment

import (
	"context"
	"strings"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/repo/image"
	momentRepo "github.com/luyb177/silent-sign-backend/internal/repo/moment"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMomentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateMomentLogic 创建动态
func NewCreateMomentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMomentLogic {
	return &CreateMomentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMomentLogic) CreateMoment(req *types.CreateMomentReq) (resp *types.Response, err error) {
	// 1. 参数校验
	if err := l.validReq(req); err != nil {
		return nil, err
	}

	// 2. 从 token 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 3. 获取 IP 位置
	location := l.getLocation()

	// 4. 构建 Moment
	m := &momentRepo.Moment{
		UserID:   authUser.UserID,
		Type:     req.Type,
		Content:  req.Content,
		VideoURL: req.VideoURL,
		Location: location,
	}
	m.RefreshHotScore()

	// 5. 创建动态
	if err = l.svcCtx.Repos.Moment.Create(l.ctx, m); err != nil {
		l.Errorf("create moment failed: %v", err)
		return nil, errorx.WrapDBInsert("创建动态失败", err)
	}

	// 6. 批量写入图片（如有）
	if len(req.ImageURLs) > 0 {
		if err := l.createImages(m.ID, req.ImageURLs); err != nil {
			l.Errorf("create images failed: %v", err)
			// 图片写入失败不阻塞主流程，仅记录日志
		}
	}

	return &types.Response{}, nil
}

func (l *CreateMomentLogic) createImages(momentID uint64, urls []string) error {
	images := make([]*image.Image, 0, len(urls))
	for i, url := range urls {
		images = append(images, &image.Image{
			TargetType: constvar.TargetTypeMoment,
			TargetID:   momentID,
			URL:        url,
			Sort:       uint8(i),
		})
	}
	return l.svcCtx.Repos.Image.CreateBatch(l.ctx, images)
}

func (l *CreateMomentLogic) getLocation() string {
	loc := middleware.GetIPLocation(l.ctx)
	if loc == nil {
		return "未知"
	}
	return middleware.ShortLocation(loc)
}

func (l *CreateMomentLogic) validReq(req *types.CreateMomentReq) error {
	switch req.Type {
	case constvar.MomentTypeText:
		if strings.TrimSpace(req.Content) == "" {
			return errorx.WrapBadRequest("文字内容不能为空", nil)
		}
	case constvar.MomentTypeVideo:
		if req.VideoURL == "" {
			return errorx.WrapBadRequest("视频地址不能为空", nil)
		}
	case constvar.MomentTypeTextVideo:
		if strings.TrimSpace(req.Content) == "" {
			return errorx.WrapBadRequest("文字内容不能为空", nil)
		}
		if req.VideoURL == "" {
			return errorx.WrapBadRequest("视频地址不能为空", nil)
		}
	case constvar.MomentTypeTextImage:
		if strings.TrimSpace(req.Content) == "" {
			return errorx.WrapBadRequest("文字内容不能为空", nil)
		}
		if len(req.ImageURLs) == 0 {
			return errorx.WrapBadRequest("至少需要上传一张图片", nil)
		}
	case constvar.MomentTypeImage:
		if len(req.ImageURLs) == 0 {
			return errorx.WrapBadRequest("至少需要上传一张图片", nil)
		}
	default:
		return errorx.WrapBadRequest("无效的动态类型", nil)
	}
	return nil
}
