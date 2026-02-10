package graph

import (
	"encoding/json"

	"github.com/augustdev/autoclip/internal/graph/model"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/resources"
)

func dbResourceToModel(dbResource *resources.Resource) *model.Resource {
	var metadata *model.ResourceMetadata
	if dbResource.Metadata != nil {
		var m map[string]string
		if err := json.Unmarshal(dbResource.Metadata, &m); err == nil {
			metadata = &model.ResourceMetadata{
				Size:     strPtr(m["size"]),
				Hostname: strPtr(m["hostname"]),
				Group:    strPtr(m["group"]),
			}
		}
	}

	return &model.Resource{
		ID:        dbResource.ID,
		Name:      dbResource.Name,
		Type:      dbResource.Type,
		Provider:  dbResource.Provider,
		Region:    dbResource.Region,
		Status:    dbResource.Status,
		Metadata:  metadata,
		ProjectID: dbResource.ProjectID,
		CreatedAt: dbResource.CreatedAt.Time,
		UpdatedAt: dbResource.UpdatedAt.Time,
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
