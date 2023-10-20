package state

import (
	"github.com/Ferlab-Ste-Justine/terracd/recurrence"
)


type State struct {
	LastCommandOccurrence recurrence.CommandOccurrence `yaml:"last_command_occurrence"`
	//Metrics Metrics
}