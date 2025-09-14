package provision

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
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
		return fmt.Errorf("error adding known host, stderr %s: %v", stderr.String(), err)
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

func acceptHostKeys(config entity.Config) error {
	err := addKnownHosts(config.AcceptHostKeys)
	if err != nil {
		internal.Logger.Error().Stack().Err(err).Msg("Error adding known host keys")
		return err
	}

	return nil
}
