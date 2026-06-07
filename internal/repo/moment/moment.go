package moment

import (
	"math"
	"time"

	"github.com/luyb177/silent-sign-backend/internal/constvar"
	"gorm.io/plugin/soft_delete"
)

// Moment 动态（类似微信朋友圈）
type Moment struct {
	ID        uint64                `gorm:"primarykey;type:bigint unsigned auto_increment"`
	CreatedAt time.Time             `gorm:"type:datetime(3)"`
	UpdatedAt time.Time             `gorm:"type:datetime(3)"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:nano;type:bigint unsigned"`

	UserID   uint64 `gorm:"index;type:bigint unsigned"`
	Type     uint8  `gorm:"index;default:1;type:tinyint unsigned"`
	Content  string `gorm:"type:text"`
	VideoURL string `gorm:"type:varchar(512)"` // 单个视频地址
	Location string `gorm:"type:varchar(255)"` // 发布动态时ip的位置

	// 冗余计数字段
	LikeNum    uint64 `gorm:"type:bigint unsigned;default:0"`
	CommentNum uint64 `gorm:"type:bigint unsigned;default:0"`
	ShareNum   uint64 `gorm:"type:bigint unsigned;default:0"`
	ViewNum    uint64 `gorm:"type:bigint unsigned;default:0"`

	// 热度分数，按热度倒序排列时使用，定时/写入时计算
	HotScore float64 `gorm:"index;type:decimal(10,2)"`
}

func (Moment) TableName() string {
	return "moments"
}

// RefreshHotScore 刷新 m 的热度分数（点赞/评论/分享/创建后调用）
func (m *Moment) RefreshHotScore() {
	m.HotScore = CalcHotScore(m.LikeNum, m.CommentNum, m.ShareNum, m.CreatedAt)
}

// CalcHotScore 计算热度分数
func CalcHotScore(likeNum, commentNum, shareNum uint64, createdAt time.Time) float64 {
	score := float64(likeNum)*constvar.MomentHotScoreLike +
		float64(commentNum)*constvar.MomentHotScoreComment +
		float64(shareNum)*constvar.MomentHotScoreShare +
		float64(createdAt.Unix())/constvar.MomentHotScoreDecay

	return math.Round(score*100) / 100
}
