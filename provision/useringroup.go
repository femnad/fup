package provision

import (
	"errors"
	"fmt"
	"os/user"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/mare"
)

func ensureUser(userName string) error {
	return internal.MaybeRunWithSudo(fmt.Sprintf("useradd -m %s", userName))
}

func addUserToGroup(user, group string) error {
	isRoot, err := internal.IsUserRoot()
	if err != nil {
		return err
	}

	internal.Log.Infof("Adding user %s to group %s", user, group)
	usermod := fmt.Sprintf("usermod -aG %s %s", group, user)
	err = marecmd.RunErrOnly(marecmd.Input{Command: usermod, Sudo: !isRoot})
	if err != nil {
		return err
	}

	return nil
}

func groupAdd(group entity.Group) error {
	isRoot, err := internal.IsUserRoot()
	if err != nil {
		return err
	}

	cmd := "groupadd"
	if group.System {
		cmd += " -r"
	}
	cmd += " " + group.Name

	internal.Log.Infof("Creating group %s", group.Name)
	err = marecmd.RunErrOnly(marecmd.Input{Command: cmd, Sudo: !isRoot})
	return err
}

func ensureGroup(group entity.Group) error {
	var unknownGroupError user.UnknownGroupError
	_, err := user.LookupGroup(group.Name)
	if err == nil {
		return nil
	} else if !errors.As(err, &unknownGroupError) {
		return err
	}

	return groupAdd(group)
}

func doEnsureUserInGroups(spec entity.UserGroupSpec) error {
	userName := spec.Name
	u, err := user.Lookup(userName)
	if err != nil && !spec.Ensure {
		return err
	}

	err = ensureUser(userName)
	if err != nil {
		return err
	}

	groups := spec.Groups
	for _, g := range groups {
		if !g.Ensure {
			continue
		}

		err = ensureGroup(g)
		if err != nil {
			return err
		}
	}

	groupIds, err := u.GroupIds()
	if err != nil {
		return err
	}

	var userGroups []string
	for _, gid := range groupIds {
		var group *user.Group
		group, err = user.LookupGroupId(gid)
		if err != nil {
			return err
		}
		userGroups = append(userGroups, group.Name)
	}

	desiredGroups := mare.MapToString(groups, func(group entity.Group) string {
		return group.Name
	})
	desired := internal.SetFromList[string](desiredGroups)
	current := internal.SetFromList[string](userGroups)
	missing := desired.Difference(current)

	missing.Each(func(missingGroup string) bool {
		err = addUserToGroup(userName, missingGroup)
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

func ensureUserInGroups(userGroupSpecs entity.UserInGroupSpec) error {
	for _, spec := range userGroupSpecs {
		err := doEnsureUserInGroups(spec)
		if err != nil {
			return err
		}
	}

	return nil
}

func userInGroup(config entity.Config) error {
	err := ensureUserInGroups(config.UserInGroup)
	if err != nil {
		internal.Log.Errorf("error ensuring user in groups: %v", err)
		return err
	}

	return nil
}
