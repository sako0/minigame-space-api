package handler

import "time"

var retryInterval = 500 * time.Millisecond

func isValidRoomId(roomId uint) bool {
	return roomId != 0
}

func isValidUserId(fromUserID uint) bool {
	return fromUserID != 0
}

// 一時的なエラーかどうかを判定する
func isTemporary(err error) bool {
	te, ok := err.(interface {
		Temporary() bool
	})
	return ok && te.Temporary()
}
