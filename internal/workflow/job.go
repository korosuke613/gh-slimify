package workflow

import (
	"regexp"
	"strings"
)

var (
	// containerCommandPatterns lists regex patterns that match container commands
	// Each pattern is compiled and checked against run commands.
	// Future additions could include: podman commands, containerd commands, etc.
	containerCommandPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\bdocker[\s-](?:build|run|exec|ps|pull|push|tag|login)\b`),
		regexp.MustCompile(`\bdocker-compose\b`),
		regexp.MustCompile(`\bdocker\s+compose\b`),
	}

	// containerActionPrefixes lists prefixes that indicate container-based GitHub Actions
	// This covers:
	// - docker:// image syntax (e.g., "docker://alpine:latest")
	// - docker/ organization actions (e.g., "docker/build-push-action@v6")
	// Future additions could include: "container://", "podman/", etc.
	containerActionPrefixes = []string{"docker"}
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
		if step.Run == "" {
			continue
		}

		runLower := strings.ToLower(step.Run)
		// Check if run command matches any container command pattern
		for _, pattern := range containerCommandPatterns {
			if pattern.MatchString(runLower) {
				return true
			}
		}
	}
	return false
}

// HasContainerActions checks if a job uses container-based GitHub Actions
// It detects actions that use container prefixes defined in containerActionPrefixes:
// - docker:// image syntax (e.g., "docker://alpine:latest")
// - docker/ organization actions (e.g., "docker/build-push-action@v6")
// Future container tools can be added by extending containerActionPrefixes.
func (j *Job) HasContainerActions() bool {
	for _, step := range j.Steps {
		if step.Uses == "" {
			continue
		}
		uses := step.Uses
		// Check if uses starts with any container action prefix
		for _, prefix := range containerActionPrefixes {
			if strings.HasPrefix(uses, prefix) {
				return true
			}
		}
	}
	return false
}

// HasServices checks if a job uses services
// Services are containers that are shared between jobs.
// Since ubuntu-slim runs itself inside a container and does not provide dockerd,
// nested container jobs are not supported.
func (j *Job) HasServices() bool {
	return j.Services != nil
}

// HasContainer checks if a job uses the container: syntax
// Jobs with container: run steps inside a Docker container, which requires
// access to the Docker daemon. Since ubuntu-slim runs itself inside a container
// and does not provide dockerd, nested container jobs are not supported.
func (j *Job) HasContainer() bool {
	return j.Container != nil
}
