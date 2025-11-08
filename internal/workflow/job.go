package workflow

import (
	"regexp"
	"strings"
)

var (
	// TODO: more efficient way to check for docker commands
	dockerCommandPattern = regexp.MustCompile(`\b(?:docker[\s-](?:build|run|exec|ps|pull|push|tag|login)|docker-compose|docker\s+compose)\b`)
)

// IsUbuntuLatest checks if a job runs on ubuntu-latest
func (j *Job) IsUbuntuLatest() bool {
	if j.RunsOn == nil {
		return false
	}

	switch v := j.RunsOn.(type) {
	case string:
		return v == "ubuntu-latest"
	case []any:
		// runs-on can be a matrix or array
		for _, item := range v {
			if str, ok := item.(string); ok && str == "ubuntu-latest" {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// HasDockerCommands checks if a job uses Docker commands
// It checks if the job uses any Docker commands in the run commands.
// Matches patterns like "docker build", "docker-compose", "sudo docker run", etc.
func (j *Job) HasDockerCommands() bool {
	for _, step := range j.Steps {
		if step.Run != "" {
			runLower := strings.ToLower(step.Run)
			if dockerCommandPattern.MatchString(runLower) {
				return true
			}
		}
	}

	return false
}

// HasDockerActions checks if a job uses Docker-based GitHub Actions
// It detects:
// - docker:// image syntax (e.g., "docker://alpine:latest")
// - docker/ organization actions (e.g., "docker/build-push-action@v6")
// This covers all Docker-related actions without needing to enumerate specific action names.
func (j *Job) HasDockerActions() bool {
	for _, step := range j.Steps {
		if step.Uses == "" {
			continue
		}

		uses := step.Uses

		// Check for docker:// image syntax
		// TODO: more efficient way to check for docker:// image syntax
		if strings.HasPrefix(uses, "docker://") {
			return true
		}

		// Check if uses starts with docker/ organization
		// This covers all docker/* actions including:
		// - docker/build-push-action
		// - docker/login-action
		// - docker/setup-buildx-action
		// - docker/setup-qemu-action
		// - and any other docker/* actions
		// TODO: more efficient way to check for docker/ organization actions
		if strings.HasPrefix(uses, "docker/") {
			return true
		}
	}

	return false
}

// HasServices checks if a job uses services
func (j *Job) HasServices() bool {
	return j.Services != nil
}
