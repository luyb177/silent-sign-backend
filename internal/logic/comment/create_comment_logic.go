// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package comment

import (
	"context"
	"errors"
	"strings"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	commentRepo "github.com/luyb177/silent-sign-backend/internal/repo/comment"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateCommentLogic 创建评论
func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCommentLogic) CreateComment(req *types.CreateCommentReq) (resp *types.Response, err error) {
	// 1. 参数校验
	if err = l.validReq(req); err != nil {
		return nil, err
	}

	// 2. 从 token 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 3. 获取 IP 位置
	location := l.getLocation()

	// 4. 构建评论
	c := &commentRepo.Comment{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		FatherID:   req.FatherID,
		CreatorID:  authUser.UserID,
		Content:    req.Content,
		Location:   location,
	}

	// 5. 在事务中：校验父评论 → 创建评论 → 后置处理
	err = l.svcCtx.Repos.Transaction(func(tx *gorm.DB) error {
		if req.FatherID != 0 {
			if err := l.validateParent(req.FatherID, req.TargetType, req.TargetID, tx); err != nil {
				return err
			}
		}
		if err := l.svcCtx.Repos.Comment.Create(l.ctx, c, tx); err != nil {
			return err
		}
		return l.afterCreate(req, tx)
	})
	if err != nil {
		l.Errorf("create comment failed: %v", err)
		return nil, errorx.WrapDBInsert("创建评论失败", err)
	}

	return &types.Response{}, nil
}

// validateParent 校验父评论存在且属于同一目标
func (l *CreateCommentLogic) validateParent(fatherID uint64, targetType uint8, targetID uint64, tx *gorm.DB) error {
	parent, err := l.svcCtx.Repos.Comment.FindByID(l.ctx, fatherID, tx)
	if err != nil {
		return err
	}
	if parent == nil {
		return errorx.WrapBadRequest("父评论不存在", nil)
	}
	if parent.TargetType != targetType || parent.TargetID != targetID {
		return errorx.WrapBadRequest("父评论与目标不匹配", nil)
	}
	if parent.FatherID != 0 {
		return errorx.WrapBadRequest("不支持三级评论", nil)
	}
	return nil
}

// afterCreate 评论创建后的后置处理（事务内），根据目标类型分发
func (l *CreateCommentLogic) afterCreate(req *types.CreateCommentReq, tx *gorm.DB) error {
	switch req.TargetType {
	case constvar.TargetTypeMoment:
		return l.incrMomentComment(req.TargetID, tx)
	}
	return nil
}

// incrMomentComment 原子增加动态评论数并刷新热度（避免 read-modify-write 覆盖）
func (l *CreateCommentLogic) incrMomentComment(momentID uint64, tx *gorm.DB) error {
	err := l.svcCtx.Repos.Moment.AdjustCommentNum(l.ctx, momentID, 1, tx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errorx.WrapBadRequest("动态不存在", nil)
	}
	return err
}

func (l *CreateCommentLogic) validReq(req *types.CreateCommentReq) error {
	if strings.TrimSpace(req.Content) == "" {
		return errorx.WrapBadRequest("评论内容不能为空", nil)
	}
	if req.TargetID == 0 {
		return errorx.WrapBadRequest("目标ID不能为空", nil)
	}
	return nil
}

func (l *CreateCommentLogic) getLocation() string {
	loc := middleware.GetIPLocation(l.ctx)
	if loc == nil {
		return "未知"
	}
	return middleware.ShortLocation(loc)
}
