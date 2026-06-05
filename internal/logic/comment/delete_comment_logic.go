// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package comment

import (
	"context"
	"errors"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	commentRepo "github.com/luyb177/silent-sign-backend/internal/repo/comment"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除评论
func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCommentLogic) DeleteComment(req *types.DeleteCommentReq) (resp *types.Response, err error) {
	// 1. 参数校验
	if req.CommentID == 0 {
		return nil, errorx.WrapBadRequest("评论ID不能为空", nil)
	}

	// 2. 从 token 获取当前用户
	authUser := middleware.GetAuthUser(l.ctx)
	if authUser == nil {
		return nil, errorx.ErrUnauthorized
	}

	// 3. 查找评论
	comment, err := l.svcCtx.Repos.Comment.FindByID(l.ctx, req.CommentID)
	if err != nil {
		l.Errorf("find comment failed: %v", err)
		return nil, errorx.WrapDBQuery("查询评论失败", err)
	}
	if comment == nil {
		return nil, errorx.WrapBadRequest("评论不存在", nil)
	}
	if comment.CreatorID != authUser.UserID {
		return nil, errorx.ErrForbidden
	}

	// 4. 计算需要从 moment 扣减的评论数
	delta := 1
	if comment.FatherID == 0 {
		delta += int(comment.SubNum) // 父评论：连带所有子评论
	}

	// 5. 事务内：级联删除子评论 → 删除评论 → 维护父SubNum → 扣减目标计数
	err = l.svcCtx.Repos.Transaction(func(tx *gorm.DB) error {
		// 如果是父评论，先级联删除所有子评论
		if comment.FatherID == 0 && comment.SubNum > 0 {
			if err := l.svcCtx.Repos.Comment.DeleteByFatherID(l.ctx, comment.ID, tx); err != nil {
				return err
			}
		}
		// 如果是子评论，减少父评论的 SubNum
		if comment.FatherID != 0 {
			if err := l.svcCtx.Repos.Comment.AdjustSubNum(l.ctx, comment.FatherID, -1, tx); err != nil {
				return err
			}
		}
		if err := l.svcCtx.Repos.Comment.Delete(l.ctx, authUser.UserID, req.CommentID, tx); err != nil {
			return err
		}
		return l.afterDelete(comment, delta, tx)
	})
	if err != nil {
		l.Errorf("delete comment failed: %v", err)
		return nil, errorx.WrapDBDelete("删除评论失败", err)
	}

	return &types.Response{}, nil
}

// afterDelete 评论删除后的后置处理（事务内），根据目标类型分发
func (l *DeleteCommentLogic) afterDelete(comment *commentRepo.Comment, delta int, tx *gorm.DB) error {
	switch comment.TargetType {
	case constvar.TargetTypeMoment:
		return l.decrMomentComment(comment.TargetID, delta, tx)
	}
	return nil
}

// decrMomentComment 原子减少动态评论数并刷新热度
func (l *DeleteCommentLogic) decrMomentComment(momentID uint64, delta int, tx *gorm.DB) error {
	err := l.svcCtx.Repos.Moment.AdjustCommentNum(l.ctx, momentID, -delta, tx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errorx.WrapBadRequest("动态不存在", nil)
	}
	return err
}
