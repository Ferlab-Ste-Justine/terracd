package state

import (
	"ferlab/terracd/recurrence"
)


type State struct {
	LastCommandOccurrence recurrence.CommandOccurrence `yaml:"last_command_occurrence"`
	//Metrics Metrics
}