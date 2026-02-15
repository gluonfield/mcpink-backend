package graph

import (
	"context"

	"github.com/augustdev/autoclip/internal/graph/model"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/projects"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/services"
)

func (r *Resolver) getServicesForProject(ctx context.Context, projectID string) ([]*model.Service, error) {
	dbServices, err := r.ServiceQueries.ListServicesByProjectID(ctx, services.ListServicesByProjectIDParams{
		ProjectID: projectID,
		Limit:     1000,
		Offset:    0,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*model.Service, len(dbServices))
	for i, dbSvc := range dbServices {
		result[i] = dbServiceToModel(&dbSvc)
		if dep, err := r.DeployService.GetLatestDeployment(ctx, dbSvc.ID); err == nil {
			enrichServiceWithDeployment(result[i], dep)
		}
		if cd, err := r.CustomDomainQueries.GetByServiceID(ctx, dbSvc.ID); err == nil {
			result[i].CustomDomain = &cd.Domain
			result[i].CustomDomainStatus = &cd.Status
		}
	}
	return result, nil
}

func dbProjectToModel(dbProject *projects.Project, projectServices []*model.Service) *model.Project {
	if projectServices == nil {
		projectServices = []*model.Service{}
	}
	return &model.Project{
		ID:        dbProject.ID,
		Name:      dbProject.Name,
		Ref:       dbProject.Ref,
		Services:  projectServices,
		CreatedAt: dbProject.CreatedAt.Time,
		UpdatedAt: dbProject.UpdatedAt.Time,
	}
}
