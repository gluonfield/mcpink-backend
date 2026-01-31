package coolify

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// LogEntry represents a single log entry from an application.
type LogEntry struct {
	Timestamp string `json:"timestamp,omitempty"`
	Message   string `json:"message,omitempty"`
	Stream    string `json:"stream,omitempty"` // stdout or stderr
}

// GetLogsOptions represents the options for retrieving application logs.
type GetLogsOptions struct {
	// Since returns logs since this many seconds ago (e.g., 3600 for last hour)
	Since int
	// Tail returns only the last N lines (default: 100)
	Tail int
	// Timestamps includes timestamps in the output
	Timestamps bool
}

// GetLogs retrieves the logs for an application.
// See: https://coolify.io/docs/api-reference/api/operations/get-application-logs-by-uuid
func (s *ApplicationsService) GetLogs(ctx context.Context, uuid string, opts *GetLogsOptions) ([]LogEntry, error) {
	if uuid == "" {
		return nil, fmt.Errorf("coolify: uuid is required")
	}

	query := url.Values{}
	if opts != nil {
		if opts.Since > 0 {
			query.Set("since", strconv.Itoa(opts.Since))
		}
		if opts.Tail > 0 {
			query.Set("tail", strconv.Itoa(opts.Tail))
		}
		if opts.Timestamps {
			query.Set("timestamps", "true")
		}
	}

	var logs []LogEntry
	if err := s.client.do(ctx, "GET", "/api/v1/applications/"+uuid+"/logs", query, nil, &logs); err != nil {
		return nil, fmt.Errorf("failed to get application logs: %w", err)
	}
	return logs, nil
}
