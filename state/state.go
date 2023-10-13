package state

import (
	"ferlab/terracd/reccurrence"
)


type State struct {
	LastCommandOccurrence reccurrence.CommandOccurrence `yaml:"last_command_occurrence"`
	//Metrics Metrics
}