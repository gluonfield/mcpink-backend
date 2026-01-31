/*
Package coolify provides a Go SDK for the Coolify API.

Coolify is an open-source, self-hostable alternative to Heroku/Netlify/Vercel.
This SDK provides a type-safe interface to Coolify's REST API.

# Authentication

All API requests require a Bearer token. Generate one in the Coolify dashboard
under Keys & Tokens > API tokens.

# Quick Start

Create a client and deploy an application from a private GitHub repository:

	client, err := coolify.NewClient(coolify.Config{
		BaseURL: "https://coolify.example.com",
		Token:   "your-api-token",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create application from private GitHub repo
	resp, err := client.Applications.CreatePrivateGitHubApp(ctx, &coolify.CreatePrivateGitHubAppRequest{
		// Required fields
		ProjectUUID:     "project-uuid",
		ServerUUID:      "server-uuid",
		EnvironmentName: "production",
		GitHubAppUUID:   "github-app-uuid",
		GitRepository:   "github.com/user/repo",
		GitBranch:       "main",
		BuildPack:       coolify.BuildPackNixpacks,
		PortsExposes:    "3000",

		// Optional fields - use helper functions for pointers
		InstantDeploy:       coolify.Bool(true),  // Deploy immediately
		IsAutoDeployEnabled: coolify.Bool(true),  // Auto-deploy on push (default: true)
		LimitsMemory:        "512m",
		LimitsCPUs:          "0.5",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created application:", resp.UUID)

# Available Services

The client provides access to the following services:

  - Applications: Create, update, delete, start, stop, restart applications
  - Applications.Envs: Manage environment variables

# Endpoints Implemented

Applications:
  - GET    /api/v1/applications                          - List all applications
  - GET    /api/v1/applications/{uuid}                   - Get application by UUID
  - POST   /api/v1/applications/public                   - Create from public repo
  - POST   /api/v1/applications/private-github-app       - Create from private repo (GitHub App)
  - PATCH  /api/v1/applications/{uuid}                   - Update application
  - DELETE /api/v1/applications/{uuid}                   - Delete application
  - GET    /api/v1/applications/{uuid}/start             - Start/deploy application
  - GET    /api/v1/applications/{uuid}/stop              - Stop application
  - GET    /api/v1/applications/{uuid}/restart           - Restart application
  - GET    /api/v1/applications/{uuid}/logs              - Get application logs

Environment Variables:
  - GET    /api/v1/applications/{uuid}/envs              - List environment variables
  - POST   /api/v1/applications/{uuid}/envs              - Create environment variable
  - PATCH  /api/v1/applications/{uuid}/envs              - Update environment variable
  - PATCH  /api/v1/applications/{uuid}/envs/bulk         - Bulk update environment variables
  - DELETE /api/v1/applications/{uuid}/envs              - Delete environment variable

# Error Handling

The SDK returns structured errors for API failures:

	app, err := client.Applications.Get(ctx, "invalid-uuid")
	if err != nil {
		var apiErr *coolify.Error
		if errors.As(err, &apiErr) {
			fmt.Printf("API error: %s (status %d)\n", apiErr.Message, apiErr.StatusCode)
		}
	}

# API Reference

For the complete API documentation, see: https://coolify.io/docs/api-reference
*/
package coolify
