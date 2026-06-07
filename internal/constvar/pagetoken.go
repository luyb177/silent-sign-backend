package constvar

import "time"

const (
	MomentPageTokenPrefix        = "moment"
	CommentPageTokenPrefix       = "comment"
	MessagePageTokenPrefix       = "message"
	FriendPageTokenPrefix        = "friend"
	FriendRequestPageTokenPrefix = "friend_req"
)

const (
	DefaultPageSize = 20
)

// TimeLocation 数据库时区，与 DSN 中 loc 参数保持一致
var TimeLocation = time.FixedZone("CST", 8*3600)
