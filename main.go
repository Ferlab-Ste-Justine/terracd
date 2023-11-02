package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Ferlab-Ste-Justine/terracd/cmd"
	"github.com/Ferlab-Ste-Justine/terracd/config"
	"github.com/Ferlab-Ste-Justine/terracd/fs"
	"github.com/Ferlab-Ste-Justine/terracd/hook"
	"github.com/Ferlab-Ste-Justine/terracd/metrics"
	"github.com/Ferlab-Ste-Justine/terracd/state"
)

func main() {
	conf, configErr := config.GetConfig()
	if configErr != nil {
		fmt.Println(configErr.Error())
		os.Exit(1)
	}

	paths := fs.GetPaths(conf.WorkingDirectory)

	var skipped bool
	execErr := state.WrapInState(func(st state.State) (state.State, error) {
		var newSt state.State
		var err error
		newSt, skipped, err = cmd.RunConfig(paths, conf, st)
		return newSt, err
	}, conf.StateStore, paths)
	

	var opResult hook.OpResult
	if execErr != nil {
		fmt.Println(execErr.Error())
		opResult = hook.OpFailure
	} else if skipped {
		opResult = hook.OpSkip
	} else {
		opResult = hook.OpSuccess
	}

	now := time.Now()
	hookErr := conf.TerminationHooks.Run(opResult)
	metricsErr := metrics.PushMetrics(conf.Metrics, conf.Command, opResult.ToString(), now)
	
	if hookErr != nil {
		fmt.Println(hookErr.Error())
	}

	if metricsErr != nil {
		fmt.Println(metricsErr.Error())
	}

	if hookErr != nil || metricsErr != nil || execErr != nil {
		os.Exit(1)
	}
}
