package provision

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	marecmd "github.com/femnad/mare/cmd"
)

const knownHostsFile = "~/.ssh/known_hosts"

func addKnownHost(host string) error {
	out, err := marecmd.RunFmtErr(marecmd.Input{Command: fmt.Sprintf("ssh-keyscan %s", host)})
	if err != nil {
		return fmt.Errorf("error adding known host, stderr %s: %v", err)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer([]byte(out.Stdout)))
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
