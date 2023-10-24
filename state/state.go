package state

import (
	"github.com/Ferlab-Ste-Justine/terracd/recurrence"
)

type State struct {
	LastCommandOccurrence recurrence.CommandOccurrence `yaml:"last_command_occurrence"`
	//Metrics Metrics
}

type StateScopedFn func(State) (State, error)

func WrapInState(fn StateScopedFn, conf StateStoreConfig) error {
	var st State
	var stateErr error
	var store StateStore
	var storeErr error
	if conf.IsDefined() {
		store, storeErr = conf.GetStore()
		if storeErr != nil {
			return storeErr
		}
		
		initErr := store.Initialize()
		if initErr != nil {
			return initErr
		}

		defer store.Cleanup()

		st, stateErr = store.Read()
		if stateErr != nil {
			return stateErr
		}
	}

	newSt, fnErr := fn(st)
	if fnErr != nil {
		return fnErr
	}

	if conf.IsDefined() {
		writeErr := store.Write(newSt)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}