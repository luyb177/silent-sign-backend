// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package comment

import (
	"context"
	"time"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/pkg/pagetoken"
	commentRepo "github.com/luyb177/silent-sign-backend/internal/repo/comment"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCommentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListCommentsLogic 分页获取评论列表
func NewListCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCommentsLogic {
	return &ListCommentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCommentsLogic) ListComments(req *types.ListCommentsReq) (resp *types.ListCommentsResp, err error) {
	// 1. 参数校验与默认值
	if req.TargetID == 0 {
		return nil, errorx.WrapBadRequest("目标ID不能为空", nil)
	}
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 50 {
		pageSize = constvar.DefaultPageSize
	}
	limit := int(pageSize) + 1 // 多取一条判断 has_more

	// 2. 解析 page_token
	var pt types.CommentPageToken
	if req.PageToken != "" {
		if err := pagetoken.Decode(req.PageToken, constvar.CommentPageTokenPrefix, &pt); err != nil {
			return nil, errorx.WrapBadRequest("分页参数无效", err)
		}
	}

	// 3. 按场景查询：father_id=0 热度排序，father_id>0 时间正序
	var comments []*commentRepo.Comment
	if req.FatherID == 0 {
		cursorTime := time.Time{}
		if pt.CreatedAt != "" {
			cursorTime, _ = time.Parse(time.DateTime, pt.CreatedAt)
		}
		comments, err = l.svcCtx.Repos.Comment.ListByHot(
			l.ctx, req.TargetType, req.TargetID,
			pt.LikeNum, pt.SubNum, cursorTime, pt.ID, limit,
		)
	} else {
		cursorTime := time.Time{}
		if pt.ID != 0 && pt.CreatedAt == "" {
			return nil, errorx.WrapBadRequest("分页参数无效", nil)
		}
		if pt.CreatedAt != "" {
			var parseErr error
			cursorTime, parseErr = time.Parse(time.DateTime, pt.CreatedAt)
			if parseErr != nil {
				return nil, errorx.WrapBadRequest("分页参数无效", parseErr)
			}
		}
		comments, err = l.svcCtx.Repos.Comment.ListByCreatedAtAsc(
			l.ctx, req.TargetType, req.TargetID, req.FatherID,
			cursorTime, pt.ID, limit,
		)
	}
	if err != nil {
		l.Errorf("list comments failed: %v", err)
		return nil, errorx.WrapDBQuery("查询评论列表失败", err)
	}

	// 4. 判断 has_more
	hasMore := len(comments) > int(pageSize)
	if hasMore {
		comments = comments[:pageSize]
	}

	// 5. 组装响应
	authUser := middleware.GetAuthUser(l.ctx)
	creatorIDs := make([]uint64, 0, len(comments))
	commentIDs := make([]uint64, 0, len(comments))
	for _, c := range comments {
		creatorIDs = append(creatorIDs, c.CreatorID)
		commentIDs = append(commentIDs, c.ID)
	}

	// 批量查用户
	creatorMap, _ := l.svcCtx.Repos.User.FindByIDs(l.ctx, creatorIDs)

	// 批量查是否点赞
	likedSet := make(map[uint64]bool)
	if authUser != nil {
		for _, cid := range commentIDs {
			liked, _ := l.svcCtx.Repos.Like.IsLiked(l.ctx, constvar.TargetTypeComment, cid, authUser.UserID)
			likedSet[cid] = liked
		}
	}

	// 组装 CommentInfo 列表
	commentInfos := make([]types.CommentInfo, 0, len(comments))
	for _, c := range comments {
		ci := types.CommentInfo{
			CommentID: c.ID,
			CreatedAt: c.CreatedAt.Format(time.DateTime),
			FartherID: c.FatherID,
			Location:  c.Location,
			Content:   c.Content,
			LikeNum:   c.LikeNum,
			SubNum:    c.SubNum,
			IsLiked:   likedSet[c.ID],
		}
		if u, ok := creatorMap[c.CreatorID]; ok {
			ci.Creator = types.CreatorInfo{
				UserID: u.ID,
				Name:   u.Name,
				Avatar: u.Avatar,
			}
		}
		commentInfos = append(commentInfos, ci)
	}

	// 6. 生成下一页 token
	nextToken := ""
	if hasMore {
		last := comments[len(comments)-1]
		nextPT := types.CommentPageToken{
			PageToken: types.PageToken{
				ID:        last.ID,
				CreatedAt: last.CreatedAt.Format(time.DateTime),
			},
			LikeNum: last.LikeNum,
			SubNum:  last.SubNum,
		}
		nextToken, _ = pagetoken.Encode(constvar.CommentPageTokenPrefix, &nextPT)
	}

	return &types.ListCommentsResp{
		Comments:  commentInfos,
		PageToken: nextToken,
		HasMore:   hasMore,
	}, nil
}
