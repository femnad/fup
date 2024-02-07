package provision

import (
	"errors"
	"fmt"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
)

func runTask(task entity.Task, cfg entity.Config) error {
	name := task.Name()
	hintMsg := ""
	if task.Hint != "" {
		hintMsg = fmt.Sprintf("Hint for running task `%s`: %s", name, task.Hint)
	}

	if !when.ShouldRun(task) {
		if hintMsg != "" && !unless.ShouldSkip(task, cfg.Settings) {
			internal.Log.Warning(hintMsg)
		}
		internal.Log.Debugf("Skipping running task %s as when condition %s evaluated to false", name, task.When)
		return nil
	}

	if unless.ShouldSkip(task, cfg.Settings) {
		internal.Log.Debugf("Skipping running task %s as unless condition %s evaluated to true", name, task.Unless)
		return nil
	}

	internal.Log.Infof("Running task: %s", name)
	return task.Run(cfg)
}

func runTasks(cfg entity.Config, tasks []entity.Task) error {
	var taskErrs []error

	for _, task := range tasks {
		err := runTask(task, cfg)
		taskErrs = append(taskErrs, err)
	}

	return errors.Join(taskErrs...)
}
