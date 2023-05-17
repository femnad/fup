package provision

import (
	"fmt"
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"strings"
)

func addUserToGroup(user, group string) error {
	usermod := fmt.Sprintf("usermod -aG %s %s", group, user)
	out, err := common.RunCmd(common.CmdIn{Command: usermod, Sudo: true})
	if err != nil {
		return fmt.Errorf("error running usermod command: %s, output %s: %v", usermod, out.Stderr, err)
	}

	return nil
}

func doEnsureUserInGroups(user string, groups []string) error {
	out, err := common.RunCmd(common.CmdIn{Command: fmt.Sprintf("groups %s", user)})
	if err != nil {
		return err
	}

	groupList := strings.Split(strings.TrimSpace(out.Stdout), " : ")
	if len(groupList) != 2 {
		return fmt.Errorf("unexpected groups output: %s", out.Stdout)
	}

	userGroups := strings.Split(groupList[1], " ")
	current := internal.SetFromList[string](userGroups)
	desired := internal.SetFromList[string](groups)
	missing := desired.Difference(current)

	missing.Each(func(missingGroup string) bool {
		err = addUserToGroup(user, missingGroup)
		if err != nil {
			return true
		}
		return false
	})
	if err != nil {
		return err
	}

	return nil
}

func ensureUserInGroups(userGroupsMap map[string][]string) error {
	for user, groups := range userGroupsMap {
		err := doEnsureUserInGroups(user, groups)
		if err != nil {
			return err
		}
	}

	return nil
}

func userInGroup(config base.Config) {
	err := ensureUserInGroups(config.UserInGroup)
	if err != nil {
		internal.Log.Errorf("error ensuring user in groups: %v", err)
	}
}
