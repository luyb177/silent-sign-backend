package constvar

// 好友申请状态
const (
	FriendRequestStatusPending  uint8 = 1 // 待处理
	FriendRequestStatusAccepted uint8 = 2 // 已通过
	FriendRequestStatusRejected uint8 = 3 // 已拒绝
)
