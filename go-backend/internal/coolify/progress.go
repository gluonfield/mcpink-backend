package coolify

import (
	"regexp"
	"strconv"
	"strings"
)

type BuildProgress struct {
	Stage       int    `json:"stage"`
	TotalStages int    `json:"total_stages"`
	Message     string `json:"message"`
}

// Pattern: "#N [stage-0 X/Y] ..." or "#N [stage-0  X/Y] ..."
var dockerStageRegex = regexp.MustCompile(`#\d+\s+\[stage-\d+\s+(\d+)/(\d+)\]\s+(.+)`)

func ExtractBuildProgress(logs []LogEntry) *BuildProgress {
	var lastProgress *BuildProgress
	for _, log := range logs {
		if matches := dockerStageRegex.FindStringSubmatch(log.Message); matches != nil {
			stage, _ := strconv.Atoi(matches[1])
			total, _ := strconv.Atoi(matches[2])
			lastProgress = &BuildProgress{
				Stage:       stage,
				TotalStages: total,
				Message:     strings.TrimSpace(matches[3]),
			}
		}
	}
	return lastProgress
}
