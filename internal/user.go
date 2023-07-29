package internal

import (
	"os/user"
	"strconv"
)

func GetCurrentUserId() (int64, error) {
	currentUser, err := user.Current()
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(currentUser.Uid, 10, 64)
}
