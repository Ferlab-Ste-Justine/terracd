package recurrence

import (
	"time"
	"github.com/Ferlab-Ste-Justine/terracd/source"
)

type Recurrence struct {
	MinInterval  time.Duration `yaml:"min_interval"`
	GitTriggers  bool          `yaml:"git_triggers"`
}

func (rec *Recurrence) IsDefined() bool {
	return rec.MinInterval > 0 
}

type Occurrence struct {
	CommitHashes []source.CommitHash `yaml:"commit_hashes"`
	Timestamp    time.Time
}

type CommandOccurrence struct {
	Command    string
	Occurrence Occurrence
}

func GitReposChanged(first []source.CommitHash, second []source.CommitHash) bool {
	if len(first) != len(second) {
		return true
	}

	for _, hash := range first {
		matched := false
		for _, hash2 := range second {
			if hash == hash2 {
				matched = true
				break
			}
		}

		if !matched {
			return true
		}
	}

	return false
}

func GenerateCommandOccurrence(cmd string, hashes []source.CommitHash) *CommandOccurrence {
	return &CommandOccurrence{
		Command: cmd,
		Occurrence: Occurrence{
			CommitHashes: hashes,
			Timestamp: time.Now(),
		},
	}
}

func (last *CommandOccurrence) ShouldOccur(rec *Recurrence, next *CommandOccurrence) bool {
	if last.Command != next.Command {
		return true
	}

	if last.Command == "migrate_backend" {
		return GitReposChanged(last.Occurrence.CommitHashes, next.Occurrence.CommitHashes)
	} else if last.Command == "plan" || last.Command == "apply" {
		if rec.GitTriggers && GitReposChanged(last.Occurrence.CommitHashes, next.Occurrence.CommitHashes) {
			return true
		}

		return last.Occurrence.Timestamp.Add(rec.MinInterval).Before(next.Occurrence.Timestamp)
	} else if last.Command == "destroy" {
		return false
	}

	return true
}