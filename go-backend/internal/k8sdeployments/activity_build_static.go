package k8sdeployments

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.temporal.io/sdk/activity"
)

func (a *Activities) StaticBuild(ctx context.Context, input BuildImageInput) (*BuildImageResult, error) {
	a.logger.Info("StaticBuild activity started",
		"imageRef", input.ImageRef,
		"sourcePath", input.SourcePath)
	defer os.RemoveAll(input.SourcePath)

	activity.RecordHeartbeat(ctx, "starting static build")

	lokiLogger := a.newBuildLokiLogger(input.Name, input.Namespace)

	// Generate a Caddy-based Dockerfile for static content
	dockerfile := `FROM caddy:2-alpine
COPY . /srv
`
	if err := os.WriteFile(filepath.Join(input.SourcePath, "Dockerfile"), []byte(dockerfile), 0o644); err != nil {
		return nil, fmt.Errorf("write static Dockerfile: %w", err)
	}

	cacheRef := ""
	if a.config.RegistryHost != "" {
		cacheRef = fmt.Sprintf("%s/cache/%s/%s:buildcache", a.config.RegistryHost, input.Namespace, input.Name)
	}

	lokiLogger.Log("Building static site image with Caddy...")
	activity.RecordHeartbeat(ctx, "building image")

	err := buildWithDockerfile(ctx, buildkitSolveOpts{
		BuildkitHost: a.config.BuildkitHost,
		SourcePath:   input.SourcePath,
		ImageRef:     input.ImageRef,
		CacheRef:     cacheRef,
		LokiLogger:   lokiLogger,
	})
	if err != nil {
		lokiLogger.Log(fmt.Sprintf("BUILD FAILED: %v", err))
		_ = lokiLogger.Flush(ctx)
		return nil, fmt.Errorf("static build: %w", err)
	}

	lokiLogger.Log(fmt.Sprintf("BUILD SUCCESS: %s", input.ImageRef))
	_ = lokiLogger.Flush(ctx)

	return &BuildImageResult{ImageRef: input.ImageRef}, nil
}
