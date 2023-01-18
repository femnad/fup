package provision

import (
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
)

func runTask(task base.Task, cfg base.Config) {
	if !when.ShouldRun(task) {
		internal.Log.Debugf("Skipping running task %s as when condition %s evaluated to false", task.Desc, task.When)
		return
	}

	if unless.ShouldSkip(task, cfg.Settings) {
		internal.Log.Debugf("Skipping running task %s as unless condition %s evaluated to true", task.Desc, task.Unless)
		return
	}

	internal.Log.Infof("Running task: %s", task.Name())
	task.Run(cfg)
}

func runTasks(cfg base.Config) {
	for _, task := range cfg.Tasks {
		runTask(task, cfg)
	}
}
