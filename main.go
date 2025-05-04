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

func MainNoExit() int {
	conf, configErr := config.GetConfig()
	if configErr != nil {
		fmt.Println(configErr.Error())
		return 1
	}

	paths := fs.GetPaths(conf.WorkingDirectory)

	var skipped bool
	usedProviders := []metrics.Provider{}
	execErr := state.WrapInState(func(st state.State) (state.State, error) {
		var newSt state.State
		var err error
		newSt, skipped, usedProviders, err = cmd.RunConfig(paths, conf, st)
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
	metricsErr := metrics.PushMetrics(conf.Metrics, conf.Command, opResult.ToString(), usedProviders, now)

	if hookErr != nil {
		fmt.Println(hookErr.Error())
	}

	if metricsErr != nil {
		fmt.Println(metricsErr.Error())
	}

	if hookErr != nil || metricsErr != nil || execErr != nil {
		return 1
	}

	return 0
}

func main() {
	code := MainNoExit()
	os.Exit(code)
}
