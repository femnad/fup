package precheck

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	quote                = `"`
	osReleaseFile        = "/etc/os-release"
	osIdField            = "ID"
	osVersionField       = "VERSION_ID"
	versionCodenameField = "VERSION_CODENAME"
)

func removeLeadingTrailingQuotes(s string) string {
	if strings.HasPrefix(s, quote) {
		s = strings.TrimPrefix(s, quote)
	}
	if strings.HasSuffix(s, quote) {
		s = strings.TrimSuffix(s, quote)
	}
	return s
}

func getOSReleaseField(f string) (string, error) {
	file, err := os.Open(osReleaseFile)
	if err != nil {
		return "", err
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
			return "", fmt.Errorf("unexpected field count when splitting line: %s", line)
		}

		field, value := fields[0], fields[1]
		if field != f {
			continue
		}

		return removeLeadingTrailingQuotes(value), nil
	}

	return "", fmt.Errorf("unable to locate OS ID line in %s", osReleaseFile)
}

func GetOSVersionCodename() (string, error) {
	return getOSReleaseField(versionCodenameField)
}

func GetOSId() (string, error) {
	return getOSReleaseField(osIdField)
}

func GetOSVersion() (float64, error) {
	versionStr, err := getOSReleaseField(osVersionField)
	if err != nil {
		return 0, err
	}

	version, err := strconv.ParseFloat(versionStr, 64)
	if err != nil {
		return 0, err
	}

	return version, nil
}
