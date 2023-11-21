package provision

import (
	"errors"
	"fmt"
	"os/user"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/mare"
)

func addUserToGroup(user, group string) error {
	internal.Log.Info("Adding user %s to group %s", user, group)
	usermod := fmt.Sprintf("usermod -aG %s %s", group, user)
	_, err := marecmd.RunFormatError(marecmd.Input{Command: usermod, Sudo: true})
	if err != nil {
		return err
	}

	return nil
}

func groupAdd(group entity.Group) error {
	groupadd := "groupadd"
	if group.System {
		groupadd += " -r"
	}
	groupadd += " " + group.Name

	internal.Log.Info("Creating group %s", group.Name)
	_, err := marecmd.RunFormatError(marecmd.Input{Command: groupadd, Sudo: true})
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

func doEnsureUserInGroups(username string, groups []entity.Group) error {
	u, err := user.Lookup(username)
	if err != nil {
		return err
	}

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
		groupName, err := user.LookupGroupId(gid)
		if err != nil {
			return err
		}
		userGroups = append(userGroups, groupName.Name)
	}

	desiredGroups := mare.MapToString(groups, func(group entity.Group) string {
		return group.Name
	})
	desired := internal.SetFromList[string](desiredGroups)
	current := internal.SetFromList[string](userGroups)
	missing := desired.Difference(current)

	missing.Each(func(missingGroup string) bool {
		err = addUserToGroup(username, missingGroup)
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

func ensureUserInGroups(userGroupsMap base.UserInGroupSpec) error {
	for u, groups := range userGroupsMap {
		err := doEnsureUserInGroups(u, groups)
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
