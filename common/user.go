package common

import "github.com/femnad/fup/internal"

const (
	rootUid = 0
)

func IsUserRoot() (bool, error) {
	userId, err := internal.GetCurrentUserId()
	if err != nil {
		return false, err
	}

	return userId == rootUid, nil
}
