package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/augustdev/autoclip/internal/storage/pg/generated/apps"
)

// GiteaPushPayload represents the incoming webhook payload from Gitea
type GiteaPushPayload struct {
	Ref        string             `json:"ref"`
	Before     string             `json:"before"`
	After      string             `json:"after"`
	Repository GiteaPushRepository `json:"repository"`
	Pusher     GiteaPushUser      `json:"pusher"`
	Sender     GiteaPushUser      `json:"sender"`
}

type GiteaPushRepository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	SSHURL   string `json:"ssh_url"`
	CloneURL string `json:"clone_url"`
	Private  bool   `json:"private"`
}

type GiteaPushUser struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

func (h *Handlers) HandleInternalGitWebhook(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-Gitea-Event")

	// Read body for signature verification
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("received internal git webhook",
		"method", r.Method,
		"event", eventType,
		"delivery", r.Header.Get("X-Gitea-Delivery"),
		"signature_present", r.Header.Get("X-Gitea-Signature") != "")

	// Verify signature
	signature := r.Header.Get("X-Gitea-Signature")
	if !h.verifyGiteaSignature(body, signature) {
		h.logger.Warn("invalid internal git webhook signature", "signature", signature)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	switch eventType {
	case "push":
		h.handleInternalGitPushWebhook(w, r, body)
		return
	default:
		h.logger.Info("ignoring unhandled internal git event", "event", eventType)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WebhookResponse{Message: "ignored event type: " + eventType})
		return
	}
}

func (h *Handlers) handleInternalGitPushWebhook(w http.ResponseWriter, r *http.Request, body []byte) {
	var payload GiteaPushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("failed to parse internal git payload", "error", err)
		http.Error(w, "failed to parse payload", http.StatusBadRequest)
		return
	}

	// Extract repo full name and branch
	repoFullName := payload.Repository.FullName
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
	after := strings.TrimSpace(payload.After)

	h.logger.Info("received internal git push webhook",
		"repo_full_name", repoFullName,
		"branch", branch,
		"after", after)

	// Look up the internal repo in database
	internalRepo, err := h.internalReposQ.GetInternalRepoByFullName(r.Context(), repoFullName)
	if err != nil {
		h.logger.Warn("internal repo not found", "full_name", repoFullName, "error", err)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WebhookResponse{Message: "repo not found"})
		return
	}

	h.logger.Info("found internal repo",
		"repo_id", internalRepo.ID,
		"user_id", internalRepo.UserID,
		"full_name", internalRepo.FullName)

	// Find apps for this repo/branch with git_provider = 'gitea'
	matchingApps, err := h.appsQ.GetAppsByRepoBranchProvider(r.Context(), apps.GetAppsByRepoBranchProviderParams{
		Repo:        repoFullName,
		Branch:      branch,
		GitProvider: "gitea",
	})
	if err != nil {
		h.logger.Error("failed to query apps", "error", err)
		http.Error(w, "failed to query apps", http.StatusInternalServerError)
		return
	}

	h.logger.Info("found matching apps",
		"repo", repoFullName,
		"branch", branch,
		"count", len(matchingApps))

	if len(matchingApps) == 0 {
		h.logger.Info("no apps found for repo/branch",
			"repo", repoFullName,
			"branch", branch)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(WebhookResponse{Message: "no apps found for this repo/branch"})
		return
	}

	// Start redeploy workflow for each app
	var deployments []DeploymentInfo
	for _, app := range matchingApps {
		workflowID, err := h.deployService.RedeployFromInternalGitPush(r.Context(), app.ID, after)
		if err != nil {
			h.logger.Error("failed to start redeploy workflow",
				"appID", app.ID,
				"error", err)
			continue
		}

		h.logger.Info("started redeploy workflow",
			"appID", app.ID,
			"workflowID", workflowID)

		deployments = append(deployments, DeploymentInfo{
			AppID:      app.ID,
			WorkflowID: workflowID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(WebhookResponse{
		Message:     "redeploy workflows started",
		Deployments: deployments,
	})
}

func (h *Handlers) verifyGiteaSignature(body []byte, signature string) bool {
	if h.giteaConfig.WebhookSecret == "" {
		h.logger.Warn("gitea webhook secret not configured, skipping verification")
		return true
	}

	if signature == "" {
		h.logger.Warn("no signature provided in internal git request")
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.giteaConfig.WebhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	valid := hmac.Equal([]byte(signature), []byte(expectedMAC))
	if !valid {
		h.logger.Warn("internal git signature mismatch",
			"expected", expectedMAC,
			"got", signature)
	}
	return valid
}
