package k8sdeployments

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/distribution/reference"
)

var registryManifestAccept = strings.Join([]string{
	"application/vnd.oci.image.manifest.v1+json",
	"application/vnd.oci.image.index.v1+json",
	"application/vnd.docker.distribution.manifest.v2+json",
	"application/vnd.docker.distribution.manifest.list.v2+json",
}, ", ")

func (a *Activities) ImageExists(ctx context.Context, imageRef string) (bool, error) {
	repository, tag, baseURL, err := resolveRegistryManifestTarget(a.config.RegistryHost, imageRef)
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf("%s/v2/%s/manifests/%s", baseURL, repository, tag)

	client := &http.Client{Timeout: 8 * time.Second}
	headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false, fmt.Errorf("create HEAD request: %w", err)
	}
	headReq.Header.Set("Accept", registryManifestAccept)

	resp, err := client.Do(headReq)
	if err != nil {
		return false, fmt.Errorf("registry HEAD manifest: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	case http.StatusMethodNotAllowed:
		// Fallback for registries that don't support HEAD.
	default:
		return false, fmt.Errorf("registry HEAD manifest unexpected status: %d", resp.StatusCode)
	}

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("create GET request: %w", err)
	}
	getReq.Header.Set("Accept", registryManifestAccept)

	getResp, err := client.Do(getReq)
	if err != nil {
		return false, fmt.Errorf("registry GET manifest: %w", err)
	}
	defer getResp.Body.Close()
	_, _ = io.Copy(io.Discard, getResp.Body)

	switch getResp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("registry GET manifest unexpected status: %d", getResp.StatusCode)
	}
}

func resolveRegistryManifestTarget(registryHost, imageRef string) (repository, tag, baseURL string, _ error) {
	named, err := reference.ParseNormalizedNamed(imageRef)
	if err != nil {
		return "", "", "", fmt.Errorf("parse image reference %q: %w", imageRef, err)
	}

	tagged, ok := named.(reference.Tagged)
	if !ok {
		return "", "", "", fmt.Errorf("image reference has no tag: %s", imageRef)
	}

	repository = reference.Path(named)
	tag = tagged.Tag()

	baseURL = strings.TrimSpace(registryHost)
	if baseURL == "" {
		baseURL = reference.Domain(named)
	}
	if baseURL == "" {
		return "", "", "", fmt.Errorf("cannot determine registry host for image %s", imageRef)
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return repository, tag, baseURL, nil
}
