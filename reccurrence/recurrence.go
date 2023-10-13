package reccurrence

import (
	"time"
)

type Recurrence struct {
	MinInterval  time.Duration `yaml:"min_interval"`
	GitTriggers  bool          `yaml:"git_triggers"`
}

func (rec *Recurrence) IsDefined() bool {
	return rec.MinInterval > 0 
}

type Occurrence struct {
	CommitHash string        `yaml:"commit_hash"`
	Timestamp  time.Time
}

type CommandOccurrence struct {
	Command    string
	Occurrence Occurrence
}

func (last *CommandOccurrence) ShouldOccur(rec *Recurrence, next *CommandOccurrence) bool {
	if last.Command != next.Command {
		return true
	}

	if last.Command == "migrate_backend" {
		return last.Occurrence.CommitHash != next.Occurrence.CommitHash
	} else if last.Command == "plan" || last.Command == "apply" {
		if rec.GitTriggers && last.Occurrence.CommitHash != next.Occurrence.CommitHash {
			return true
		}

		return last.Occurrence.Timestamp.Add(rec.MinInterval).After(next.Occurrence.Timestamp)
	}

	return true
}