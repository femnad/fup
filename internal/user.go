package internal

import (
	"os/user"
	"strconv"
)

const (
	rootUid = 0
)

func GetCurrentUserId() (int64, error) {
	currentUser, err := user.Current()
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(currentUser.Uid, 10, 64)
}

func IsUserRoot() (bool, error) {
	userId, err := GetCurrentUserId()
	if err != nil {
		return false, err
	}

	return userId == rootUid, nil
}
