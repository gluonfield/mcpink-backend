package coolify

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type LogEntry struct {
	Timestamp string `json:"timestamp,omitempty"`
	Message   string `json:"message,omitempty"`
	Stream    string `json:"stream,omitempty"`
}

type GetLogsOptions struct {
	Since      int
	Tail       int
	Timestamps bool
}

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
