package provision

import (
	"errors"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
	marecmd "github.com/femnad/mare/cmd"
)

func runUpdateCmds(cmds mapset.Set[string]) []error {
	var errs []error

	isRoot, err := internal.IsUserRoot()
	if err != nil {
		return []error{err}
	}

	cmds.Each(func(cmd string) bool {
		input := marecmd.Input{Command: cmd, Sudo: !isRoot}
		err = marecmd.RunErrOnly(input)
		errs = append(errs, err)
		return false
	})

	return errs
}

func addRepos(config entity.Config) error {
	var errs []error

	var repos []entity.OSRepo
	for _, repo := range config.AptRepos {
		repos = append(repos, repo)
	}
	for _, repo := range config.DnfRepos {
		repos = append(repos, repo)
	}

	updateCmds := mapset.NewSet[string]()
	for _, repo := range repos {
		if !when.ShouldRun(repo) {
			continue
		}

		if unless.ShouldSkip(repo, config.Settings) {
			continue
		}

		internal.Log.Infof("Adding repo %s", repo.Name())

		err := repo.Install()
		if err != nil && repo.UpdateCmd() != "" {
			internal.Log.Errorf("Error installing repo %s: %v", repo.Name(), err)
			updateCmds.Add(repo.UpdateCmd())
		}
		errs = append(errs, err)
	}

	updateErrs := runUpdateCmds(updateCmds)
	errs = append(errs, updateErrs...)

	return errors.Join(errs...)
}
