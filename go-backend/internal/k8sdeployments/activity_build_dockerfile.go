package k8sdeployments

import (
	"context"
	"fmt"
	"os"

	"go.temporal.io/sdk/activity"
)

func (a *Activities) DockerfileBuild(ctx context.Context, input BuildImageInput) (*BuildImageResult, error) {
	a.logger.Info("DockerfileBuild activity started",
		"imageRef", input.ImageRef,
		"sourcePath", input.SourcePath)
	defer os.RemoveAll(input.SourcePath)

	activity.RecordHeartbeat(ctx, "starting dockerfile build")

	lokiLogger := a.newBuildLokiLogger(input.Name, input.Namespace)

	cacheRef := ""
	if a.config.RegistryHost != "" {
		cacheRef = fmt.Sprintf("%s/cache/%s/%s:buildcache", a.config.RegistryHost, input.Namespace, input.Name)
	}

	lokiLogger.Log("Building image from Dockerfile with BuildKit...")
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
		return nil, fmt.Errorf("dockerfile build: %w", err)
	}

	lokiLogger.Log(fmt.Sprintf("BUILD SUCCESS: %s", input.ImageRef))
	_ = lokiLogger.Flush(ctx)

	return &BuildImageResult{ImageRef: input.ImageRef}, nil
}
