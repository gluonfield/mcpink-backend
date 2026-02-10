package k8sdeployments

import (
	"context"
	"fmt"
	"os"

	"github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"go.temporal.io/sdk/activity"
)

func (a *Activities) RailpackBuild(ctx context.Context, input BuildImageInput) (*BuildImageResult, error) {
	a.logger.Info("RailpackBuild activity started",
		"imageRef", input.ImageRef,
		"sourcePath", input.SourcePath)
	defer os.RemoveAll(input.SourcePath)

	activity.RecordHeartbeat(ctx, "starting railpack build")

	lokiLogger := a.newBuildLokiLogger(input.Name, input.Namespace)
	lokiLogger.Log("Generating build plan with railpack...")

	// 1. Generate build plan
	activity.RecordHeartbeat(ctx, "generating railpack plan")

	userApp, err := app.NewApp(input.SourcePath)
	if err != nil {
		lokiLogger.Log(fmt.Sprintf("RAILPACK ERROR: failed to create app: %v", err))
		_ = lokiLogger.Flush(ctx)
		return nil, fmt.Errorf("railpack new app: %w", err)
	}

	env := app.NewEnvironment(&input.EnvVars)
	result := core.GenerateBuildPlan(userApp, env, &core.GenerateBuildPlanOptions{})

	if !result.Success {
		errMsg := "railpack plan generation failed"
		for _, l := range result.Logs {
			lokiLogger.Log(fmt.Sprintf("[railpack] %s", l.Msg))
		}
		lokiLogger.Log(fmt.Sprintf("BUILD FAILED: %s", errMsg))
		_ = lokiLogger.Flush(ctx)
		return nil, fmt.Errorf(errMsg)
	}

	for _, provider := range result.DetectedProviders {
		lokiLogger.Log(fmt.Sprintf("Detected provider: %s", provider))
	}
	for _, l := range result.Logs {
		lokiLogger.Log(fmt.Sprintf("[railpack] %s", l.Msg))
	}

	// 2. Build with BuildKit using railpack's own BuildKit integration
	activity.RecordHeartbeat(ctx, "building image with railpack+buildkit")
	lokiLogger.Log("Building image with railpack + BuildKit...")

	cacheRef := ""
	if a.config.RegistryHost != "" {
		cacheRef = fmt.Sprintf("%s/cache/%s/%s:buildcache", a.config.RegistryHost, input.Namespace, input.Name)
	}

	// Railpack reads BUILDKIT_HOST from env
	os.Setenv("BUILDKIT_HOST", a.config.BuildkitHost)

	buildOpts := buildkit.BuildWithBuildkitClientOptions{
		ImageName:   input.ImageRef,
		ImportCache: cacheRef,
		ExportCache: cacheRef,
		CacheKey:    fmt.Sprintf("%s/%s", input.Namespace, input.Name),
	}

	if err := buildkit.BuildWithBuildkitClient(input.SourcePath, result.Plan, buildOpts); err != nil {
		lokiLogger.Log(fmt.Sprintf("BUILD FAILED: %v", err))
		_ = lokiLogger.Flush(ctx)
		return nil, fmt.Errorf("railpack build: %w", err)
	}

	lokiLogger.Log(fmt.Sprintf("BUILD SUCCESS: %s", input.ImageRef))
	_ = lokiLogger.Flush(ctx)

	return &BuildImageResult{ImageRef: input.ImageRef}, nil
}

func (a *Activities) newBuildLokiLogger(name, namespace string) *LokiLogger {
	return NewLokiLogger(a.config.LokiPushURL, map[string]string{
		"job":       "build",
		"service":   name,
		"namespace": namespace,
	})
}
