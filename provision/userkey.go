package provision

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
)

const authorizedKeysFile = "~/.ssh/authorized_keys"

type idKeyPair struct {
	Id  int    `yaml:"id"`
	Key string `yaml:"key"`
}

func ensureUserKeys(user string) error {
	if user == "" {
		return fmt.Errorf("GitHub username is empty")
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/keys", user)
	resp, err := remote.ReadResponseBody(url)
	if err != nil {
		internal.Log.Errorf("error reading from url %s: %v", url, err)
		return err
	}
	defer resp.Body.Close()

	keyFile := internal.ExpandUser(authorizedKeysFile)
	dir, _ := path.Split(keyFile)
	if err = common.EnsureDir(dir); err != nil {
		return err
	}

	_, err = os.Stat(keyFile)
	var fd *os.File
	if os.IsNotExist(err) {
		fd, err = os.OpenFile(keyFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
	} else if err == nil {
		fd, err = os.Open(keyFile)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	scanner.Split(bufio.ScanLines)

	var userKeyPairs []idKeyPair
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&userKeyPairs)
	if err != nil {
		return err
	}

	keySet := mapset.NewSet[string]()
	for _, pair := range userKeyPairs {
		keySet.Add(pair.Key)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if keySet.Contains(line) {
			keySet.Remove(line)
		}
	}

	keySet.Each(func(key string) bool {
		line := fmt.Sprintf("%s\n", key)
		_, err = fd.WriteString(line)
		if err != nil {
			return true
		}
		return false
	})

	return err
}

func addGithubUserKeys(config base.Config) {
	user := config.GithubUserKey.User
	err := ensureUserKeys(user)
	if err != nil {
		internal.Log.Errorf("error ensuring GitHub keys for user %s: %v", user, err)
		return
	}
}
