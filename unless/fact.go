package precheck

import (
	"fmt"
)

func isOs(osId string) (bool, error) {
	foundOsId, err := GetOsId()
	if err != nil {
		return false, fmt.Errorf("error getting OS ID %v", err)
	}

	return foundOsId == osId, nil
}

func IsDebian() (bool, error) {
	return isOs("debian")
}

func IsFedora() (bool, error) {
	return isOs("fedora")
}

func IsUbuntu() (bool, error) {
	return isOs("ubuntu")
}

var Facts = map[string]func() (bool, error){
	"is-debian": IsDebian,
	"is-fedora": IsFedora,
	"is-ubuntu": IsUbuntu,
}
