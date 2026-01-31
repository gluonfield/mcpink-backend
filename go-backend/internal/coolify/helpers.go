package coolify

// Helper functions for creating pointers to primitive types.
// These are useful when setting optional fields in request structs.
//
// Example:
//
//	req := &coolify.CreatePrivateGitHubAppRequest{
//		// ... required fields ...
//		InstantDeploy:       coolify.Bool(true),
//		IsAutoDeployEnabled: coolify.Bool(false), // Override default (true)
//		HealthCheckRetries:  coolify.Int(5),
//	}

// Bool returns a pointer to the given bool value.
func Bool(v bool) *bool {
	return &v
}

// Int returns a pointer to the given int value.
func Int(v int) *int {
	return &v
}

// String returns a pointer to the given string value.
func String(v string) *string {
	return &v
}

// BuildPackPtr returns a pointer to the given BuildPack value.
func BuildPackPtr(v BuildPack) *BuildPack {
	return &v
}

// RedirectTypePtr returns a pointer to the given RedirectType value.
func RedirectTypePtr(v RedirectType) *RedirectType {
	return &v
}
