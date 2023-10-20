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

	var st state.State
	var stateErr error
	var store state.StateStore
	var storeErr error
	if conf.StateStore.IsDefined() {
		store, storeErr = conf.StateStore.GetStore()
		if storeErr != nil {
			fmt.Println(storeErr.Error())
			os.Exit(1)
		}
		
		initErr := store.Initialize()
		if initErr != nil {
			fmt.Println(initErr.Error())
			os.Exit(1)
		}

		defer store.Cleanup()

		st, stateErr = store.Read()
		if stateErr != nil {
			fmt.Println(stateErr.Error())
			os.Exit(1)
		}

	}

	newSt, skipped, err := cmd.RunConfig(conf, st)
	if err != nil {
		fmt.Println(err.Error())

		hookErr := conf.TerminationHooks.Run(hook.OpFailure)
		if hookErr != nil {
			fmt.Println(hookErr.Error())
		}

		os.Exit(1)
	}

	if conf.StateStore.IsDefined() {
		writeErr := store.Write(newSt)
		if writeErr != nil {
			fmt.Println(writeErr.Error())
			os.Exit(1)
		}
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
