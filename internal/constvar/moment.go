package constvar

const (
	MomentTypeText      uint8 = 1 // 纯文字
	MomentTypeTextVideo uint8 = 2 // 文字 + 视频
	MomentTypeVideo     uint8 = 3 // 纯视频
	MomentTypeTextImage uint8 = 4 // 文字 + 图片
	MomentTypeImage     uint8 = 5 // 纯图片
)

// 排序类型
const (
	SortByCreatedAt uint8 = 1 // 按创建时间（最新）
	SortByHot       uint8 = 2 // 按热度
)

// 热度分值权重
// 公式：LikeNum×LikeWeight + CommentNum×CommentWeight + ShareNum×ShareWeight + Unix(created_at) / Decay
const (
	MomentHotScoreLike    = 3       // 点赞权重
	MomentHotScoreComment = 5       // 评论权重
	MomentHotScoreShare   = 8       // 分享权重
	MomentHotScoreDecay   = 45000.0 // 时间衰减系数，值越大衰减越慢
)
