package k8sdeployments

import (
	"errors"
	"fmt"
	"os"

	"go.temporal.io/sdk/temporal"
)

func sourcePathMissingError(sourcePath string, cause error) error {
	return temporal.NewNonRetryableApplicationError(
		fmt.Sprintf("source path missing: %s", sourcePath),
		"source_path_missing",
		cause,
	)
}

func isPathMissingErr(err error) bool {
	if err == nil {
		return false
	}
	if os.IsNotExist(err) {
		return true
	}
	var pathErr *os.PathError
	return errors.As(err, &pathErr) && os.IsNotExist(pathErr)
}
