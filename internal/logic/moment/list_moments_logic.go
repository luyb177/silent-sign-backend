// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package moment

import (
	"context"
	"time"

	"github.com/luyb177/silent-sign-backend/common/errorx"
	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"github.com/luyb177/silent-sign-backend/internal/middleware"
	"github.com/luyb177/silent-sign-backend/internal/pkg/pagetoken"
	momentRepo "github.com/luyb177/silent-sign-backend/internal/repo/moment"
	"github.com/luyb177/silent-sign-backend/internal/svc"
	"github.com/luyb177/silent-sign-backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMomentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListMomentsLogic 分页获取动态列表
func NewListMomentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMomentsLogic {
	return &ListMomentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMomentsLogic) ListMoments(req *types.ListMomentsReq) (resp *types.ListMomentsResp, err error) {
	// 1. 参数校验与默认值
	pageSize := req.PageSize
	if pageSize <= 0 || pageSize > 50 {
		pageSize = constvar.DefaultPageSize
	}
	limit := int(pageSize) + 1 // 多取一条判断 has_more

	sortType := req.SortType
	if sortType != constvar.SortByHot {
		sortType = constvar.SortByCreatedAt // 默认按创建时间
	}

	// 2. 解析 page_token
	var pt types.MomentPageToken
	if req.PageToken != "" {
		if err := pagetoken.Decode(req.PageToken, constvar.MomentPageTokenPrefix, &pt); err != nil {
			return nil, errorx.WrapBadRequest("分页参数无效", err)
		}
		if pt.SortType != sortType {
			return nil, errorx.WrapBadRequest("排序类型与分页token不匹配", nil)
		}
	}

	// 3. 按排序类型查询
	var moments []*momentRepo.Moment
	switch sortType {
	case constvar.SortByHot:
		moments, err = l.svcCtx.Repos.Moment.ListByHot(l.ctx, pt.ID, pt.HotScore, limit)
	default:
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
		moments, err = l.svcCtx.Repos.Moment.ListByCreatedAt(l.ctx, pt.ID, cursorTime, limit)
	}
	if err != nil {
		l.Errorf("list moments failed: %v", err)
		return nil, errorx.WrapDBQuery("查询动态列表失败", err)
	}

	// 4. 判断 has_more
	hasMore := len(moments) > int(pageSize)
	if hasMore {
		moments = moments[:pageSize]
	}

	// 5. 组装响应
	authUser := middleware.GetAuthUser(l.ctx)
	creatorIDs := make([]uint64, 0, len(moments))
	momentIDs := make([]uint64, 0, len(moments))
	for _, m := range moments {
		creatorIDs = append(creatorIDs, m.UserID)
		momentIDs = append(momentIDs, m.ID)
	}

	// 批量查用户
	creatorMap, err := l.svcCtx.Repos.User.FindByIDs(l.ctx, creatorIDs)
	if err != nil {
		l.Errorf("find moment creators failed: %v", err)
		return nil, errorx.WrapDBQuery("查找用户失败", err)
	}

	// 批量查是否点赞
	likedSet := make(map[uint64]bool)
	if authUser != nil {
		for _, mid := range momentIDs {
			liked, likeErr := l.svcCtx.Repos.Like.IsLiked(l.ctx, constvar.TargetTypeMoment, mid, authUser.UserID)
			if likeErr != nil {
				l.Errorf("check moment is_liked failed (moment_id=%d): %v", mid, likeErr)
				continue
			}
			likedSet[mid] = liked
		}
	}

	// 组装 MomentInfo 列表
	momentInfos := make([]types.MomentInfo, 0, len(moments))
	for _, m := range moments {
		mi := types.MomentInfo{
			MomentID:   m.ID,
			CreatedAt:  m.CreatedAt.Format(time.DateTime),
			UpdatedAt:  m.UpdatedAt.Format(time.DateTime),
			Type:       m.Type,
			Content:    m.Content,
			VideoURL:   m.VideoURL,
			LikeNum:    m.LikeNum,
			CommentNum: m.CommentNum,
			ShareNum:   m.ShareNum,
			ViewNum:    m.ViewNum,
			IsLiked:    likedSet[m.ID],
		}
		if u, ok := creatorMap[m.UserID]; ok {
			mi.Creator = types.CreatorInfo{
				UserID: u.ID,
				Name:   u.Name,
				Avatar: u.Avatar,
			}
		}
		momentInfos = append(momentInfos, mi)
	}

	// 6. 生成下一页 token
	nextToken := ""
	if hasMore {
		last := moments[len(moments)-1]
		nextPT := types.MomentPageToken{
			PageToken: types.PageToken{
				ID:        last.ID,
				CreatedAt: last.CreatedAt.Format(time.DateTime),
				SortType:  sortType,
			},
			HotScore: last.HotScore,
		}
		nextToken, _ = pagetoken.Encode(constvar.MomentPageTokenPrefix, &nextPT)
	}

	return &types.ListMomentsResp{
		Moments:   momentInfos,
		PageToken: nextToken,
		HasMore:   hasMore,
	}, nil
}
