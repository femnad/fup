package precheck

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	osReleaseFile = "/etc/os-release"
	osIdField     = "ID"
)

func isOs(osId string) (bool, error) {
	file, err := os.Open(osReleaseFile)
	if err != nil {
		return false, err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "=")
		if len(fields) != 2 {
			return false, fmt.Errorf("unexpected field count when splitting line: %s", line)
		}
		field, value := fields[0], fields[1]
		if field != osIdField {
			continue
		}
		return value == osId, nil
	}

	return false, fmt.Errorf("unable to locate OS ID line in %s", osReleaseFile)
}

func IsFedora() (bool, error) {
	return isOs("fedora")
}

func IsUbuntu() (bool, error) {
	return isOs("ubuntu")
}

var Facts = map[string]func() (bool, error){
	"is-fedora": IsFedora,
	"is-ubuntu": IsUbuntu,
}
