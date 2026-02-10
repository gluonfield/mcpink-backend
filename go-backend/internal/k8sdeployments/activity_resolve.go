package k8sdeployments

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.temporal.io/sdk/activity"
)

func (a *Activities) ResolveBuildContext(ctx context.Context, input ResolveBuildContextInput) (*ResolveBuildContextResult, error) {
	a.logger.Info("ResolveBuildContext activity started", "serviceID", input.ServiceID)
	activity.RecordHeartbeat(ctx, "resolving build context")

	app, err := a.appsQ.GetAppByID(ctx, input.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("get app: %w", err)
	}

	project, err := a.projectsQ.GetProjectByID(ctx, app.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}

	user, err := a.usersQ.GetUserByID(ctx, app.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	githubUsername := ""
	if user.GithubUsername != nil {
		githubUsername = *user.GithubUsername
	}
	if githubUsername == "" {
		return nil, fmt.Errorf("user %s has no github username", app.UserID)
	}

	namespace := NamespaceName(githubUsername, project.Ref)
	name := ServiceName(*app.Name)

	imageRef := fmt.Sprintf("%s/%s/%s:%s", a.config.RegistryAddress, namespace, name, input.CommitSHA)

	// Parse env vars from app
	envVars := make(map[string]string)
	if len(app.EnvVars) > 0 {
		if err := json.Unmarshal(app.EnvVars, &envVars); err != nil {
			a.logger.Warn("failed to parse env vars as map, trying array format", "error", err)
			// Try array of {key, value} format
			var envArr []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			if err := json.Unmarshal(app.EnvVars, &envArr); err == nil {
				for _, e := range envArr {
					envVars[e.Key] = e.Value
				}
			}
		}
	}
	envVars["PORT"] = app.Port

	// Determine build pack
	buildPack := app.BuildPack
	switch buildPack {
	case "nixpacks":
		buildPack = "railpack"
	case "dockerfile":
		// Check that Dockerfile exists
		if _, err := os.Stat(filepath.Join(input.SourcePath, "Dockerfile")); os.IsNotExist(err) {
			return nil, fmt.Errorf("build pack is 'dockerfile' but no Dockerfile found in repo")
		}
	case "static":
		// OK
	case "dockercompose":
		return nil, fmt.Errorf("dockercompose build pack is not supported on k8s")
	default:
		// Auto-detect: check for Dockerfile, else railpack
		if _, err := os.Stat(filepath.Join(input.SourcePath, "Dockerfile")); err == nil {
			buildPack = "dockerfile"
		} else {
			buildPack = "railpack"
		}
	}

	a.logger.Info("ResolveBuildContext completed",
		"serviceID", input.ServiceID,
		"namespace", namespace,
		"name", name,
		"buildPack", buildPack,
		"imageRef", imageRef)

	return &ResolveBuildContextResult{
		BuildPack: buildPack,
		ImageRef:  imageRef,
		Namespace: namespace,
		Name:      name,
		Port:      app.Port,
		EnvVars:   envVars,
	}, nil
}
