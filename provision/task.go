package provision

import (
	"errors"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
)

func runTask(task entity.Task, cfg entity.Config) error {
	if !when.ShouldRun(task) {
		internal.Log.Debugf("Skipping running task %s as when condition %s evaluated to false", task.Desc, task.When)
		return nil
	}

	if unless.ShouldSkip(task, cfg.Settings) {
		internal.Log.Debugf("Skipping running task %s as unless condition %s evaluated to true", task.Desc, task.Unless)
		return nil
	}

	internal.Log.Infof("Running task: %s", task.Name())
	return task.Run(cfg)
}

func runTaskGroup(group entity.TaskGroup, cfg entity.Config) []error {
	if !when.ShouldRun(group) {
		internal.Log.Debugf("Skipping running task group as when condition %s evaluated to false", group.When)
		return nil
	}

	var errs []error
	for _, task := range group.Tasks {
		err := runTask(task, cfg)
		errs = append(errs, err)
	}

	return errs
}

func runTasks(cfg entity.Config, tasks []entity.Task) error {
	var taskErrs []error

	for _, task := range tasks {
		err := runTask(task, cfg)
		taskErrs = append(taskErrs, err)
	}

	return errors.Join(taskErrs...)
}

func runCombinedTasks(cfg entity.Config) error {
	var taskErrs []error

	for _, task := range cfg.Tasks {
		err := runTask(task, cfg)
		taskErrs = append(taskErrs, err)
	}

	for _, group := range cfg.TaskGroups {
		errs := runTaskGroup(group, cfg)
		taskErrs = append(taskErrs, errs...)
	}

	return errors.Join(taskErrs...)
}
