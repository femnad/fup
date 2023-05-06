package provision

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"os/exec"
)

const knownHostsFile = "~/.ssh/known_hosts"

func addKnownHost(host string) error {
	cmd := exec.Command("ssh-keyscan", host)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf(stderr.String())
	}

	scanner := bufio.NewScanner(&stdout)
	scanner.Split(bufio.ScanLines)
	var hostKeys []string

	for scanner.Scan() {
		line := scanner.Text()
		hostKeys = append(hostKeys, line)
	}

	return common.EnsureLinesInFile(knownHostsFile, hostKeys)
}

func addKnownHosts(hosts []string) error {
	for _, host := range hosts {
		err := addKnownHost(host)
		if err != nil {
			return err
		}
	}

	return nil
}

func acceptHostKeys(config base.Config) {
	err := addKnownHosts(config.AcceptHostKeys)
	if err != nil {
		internal.Log.Errorf("error accepting host keys: %v", err)
	}
}
