package main

import (
	"fmt"
	"os"

	"github.com/Ferlab-Ste-Justine/terracd/cmd"
	"github.com/Ferlab-Ste-Justine/terracd/config"
	"github.com/Ferlab-Ste-Justine/terracd/hook"
	"github.com/Ferlab-Ste-Justine/terracd/state"
)

func main() {
	conf, configErr := config.GetConfig()
	if configErr != nil {
		fmt.Println(configErr.Error())
		os.Exit(1)
	}

	var skipped bool
	execErr := state.WrapInState(func(st state.State) (state.State, error) {
		var newSt state.State
		var err error
		newSt, skipped, err = cmd.RunConfig(conf, st)
		return newSt, err
	}, conf.StateStore)

	if execErr != nil {
		fmt.Println(execErr.Error())

		hookErr := conf.TerminationHooks.Run(hook.OpFailure)
		if hookErr != nil {
			fmt.Println(hookErr.Error())
		}

		os.Exit(1)
	}

	if skipped {
		hookErr := conf.TerminationHooks.Run(hook.OpSkip)
		if hookErr != nil {
			fmt.Println(hookErr.Error())
			os.Exit(1)
		}
		return
	}

	hookErr := conf.TerminationHooks.Run(hook.OpSuccess)
	if hookErr != nil {
		fmt.Println(hookErr.Error())
		os.Exit(1)
	}
}
