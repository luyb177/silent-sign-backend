package constvar

// 消息类型
const (
	MsgTypeFriendRequest uint8 = 1 // 好友申请
	MsgTypeFriendAccept  uint8 = 2 // 好友通过
	MsgTypeSystemNotice  uint8 = 3 // 系统通知
	MsgTypeChat          uint8 = 4 // 聊天消息（预留）
)
