package workflow

import (
	"testing"
)

func TestJob_IsUbuntuLatest_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		job      *Job
		expected bool
	}{
		{
			name: "empty string",
			job: &Job{
				RunsOn: "",
			},
			expected: false,
		},
		{
			name: "ubuntu-latest with whitespace",
			job: &Job{
				RunsOn: "  ubuntu-latest  ",
			},
			expected: false, // Exact match required
		},
		{
			name: "ubuntu-latest-extra",
			job: &Job{
				RunsOn: "ubuntu-latest-extra",
			},
			expected: false,
		},
		{
			name: "empty array",
			job: &Job{
				RunsOn: []interface{}{},
			},
			expected: false,
		},
		{
			name: "array with non-string items",
			job: &Job{
				RunsOn: []interface{}{123, true, nil},
			},
			expected: false,
		},
		{
			name: "array with ubuntu-latest at end",
			job: &Job{
				RunsOn: []interface{}{"ubuntu-22.04", "macos-latest", "ubuntu-latest"},
			},
			expected: true,
		},
		{
			name: "unsupported type - int",
			job: &Job{
				RunsOn: 123,
			},
			expected: false,
		},
		{
			name: "unsupported type - map",
			job: &Job{
				RunsOn: map[string]interface{}{"os": "ubuntu-latest"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.IsUbuntuLatest()
			if got != tt.expected {
				t.Errorf("IsUbuntuLatest() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJob_HasDockerCommands_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		job      *Job
		expected bool
	}{
		{
			name: "empty steps",
			job: &Job{
				Steps: []Step{},
			},
			expected: false,
		},
		{
			name: "step with empty run",
			job: &Job{
				Steps: []Step{{Run: ""}},
			},
			expected: false,
		},
		{
			name: "docker command in comment",
			job: &Job{
				Steps: []Step{{Run: "# docker build should not match"}},
			},
			expected: true, // Current implementation detects docker in comments too
		},
		{
			name: "docker command with prefix",
			job: &Job{
				Steps: []Step{{Run: "sudo docker build -t app ."}},
			},
			expected: true,
		},
		{
			name: "docker command in multi-line script",
			job: &Job{
				Steps: []Step{{
					Run: `#!/bin/bash
set -e
echo "Building"
docker build -t app .
echo "Done"`,
				}},
			},
			expected: true,
		},
		{
			name: "docker command in variable",
			job: &Job{
				Steps: []Step{{Run: "CMD='docker build'; $CMD"}},
			},
			expected: true,
		},
		{
			name: "docker exec",
			job: &Job{
				Steps: []Step{{Run: "docker exec container echo hello"}},
			},
			expected: true,
		},
		{
			name: "docker ps",
			job: &Job{
				Steps: []Step{{Run: "docker ps -a"}},
			},
			expected: true,
		},
		{
			name: "docker pull",
			job: &Job{
				Steps: []Step{{Run: "docker pull alpine:latest"}},
			},
			expected: true,
		},
		{
			name: "docker push",
			job: &Job{
				Steps: []Step{{Run: "docker push myregistry/app:latest"}},
			},
			expected: true,
		},
		{
			name: "docker tag",
			job: &Job{
				Steps: []Step{{Run: "docker tag app:latest app:v1"}},
			},
			expected: true,
		},
		{
			name: "docker login",
			job: &Job{
				Steps: []Step{{Run: "docker login -u user -p pass registry.com"}},
			},
			expected: true,
		},
		{
			name: "word containing docker but not command",
			job: &Job{
				Steps: []Step{{Run: "echo 'This is not a docker command'"}},
			},
			expected: false,
		},
		{
			name: "dockerhub",
			job: &Job{
				Steps: []Step{{Run: "echo 'dockerhub.com'"}},
			},
			expected: false,
		},
		{
			name: "multiple steps - first has docker",
			job: &Job{
				Steps: []Step{
					{Run: "docker build -t app ."},
					{Run: "echo 'done'"},
				},
			},
			expected: true,
		},
		{
			name: "multiple steps - last has docker",
			job: &Job{
				Steps: []Step{
					{Run: "echo 'start'"},
					{Run: "docker build -t app ."},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.HasDockerCommands()
			if got != tt.expected {
				t.Errorf("HasDockerCommands() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJob_HasContainerActions_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		job      *Job
		expected bool
	}{
		{
			name: "empty steps",
			job: &Job{
				Steps: []Step{},
			},
			expected: false,
		},
		{
			name: "step with empty uses",
			job: &Job{
				Steps: []Step{{Uses: ""}},
			},
			expected: false,
		},
		{
			name: "step without uses field",
			job: &Job{
				Steps: []Step{{Run: "echo hello"}},
			},
			expected: false,
		},
		{
			name: "docker:// with tag",
			job: &Job{
				Steps: []Step{{Uses: "docker://alpine:3.18"}},
			},
			expected: true,
		},
		{
			name: "docker:// without tag",
			job: &Job{
				Steps: []Step{{Uses: "docker://alpine"}},
			},
			expected: true,
		},
		{
			name: "docker:// with digest",
			job: &Job{
				Steps: []Step{{Uses: "docker://alpine@sha256:abc123"}},
			},
			expected: true,
		},
		{
			name: "docker/ with version",
			job: &Job{
				Steps: []Step{{Uses: "docker/build-push-action@v6.0.0"}},
			},
			expected: true,
		},
		{
			name: "docker/ with branch",
			job: &Job{
				Steps: []Step{{Uses: "docker/build-push-action@main"}},
			},
			expected: true,
		},
		{
			name: "docker/ with commit SHA",
			job: &Job{
				Steps: []Step{{Uses: "docker/build-push-action@abc123def456"}},
			},
			expected: true,
		},
		{
			name: "docker/build-push-action without version",
			job: &Job{
				Steps: []Step{{Uses: "docker/build-push-action"}},
			},
			expected: true,
		},
		{
			name: "docker/login-action with version",
			job: &Job{
				Steps: []Step{{Uses: "docker/login-action@v3.0.0"}},
			},
			expected: true,
		},
		{
			name: "docker/setup-buildx-action",
			job: &Job{
				Steps: []Step{{Uses: "docker/setup-buildx-action@v3"}},
			},
			expected: true,
		},
		{
			name: "docker/setup-qemu-action",
			job: &Job{
				Steps: []Step{{Uses: "docker/setup-qemu-action@v3"}},
			},
			expected: true,
		},
		{
			name: "action containing docker but not docker/",
			job: &Job{
				Steps: []Step{{Uses: "myorg/docker-helper-action@v1"}},
			},
			expected: false,
		},
		{
			name: "action with docker in name but not prefix",
			job: &Job{
				Steps: []Step{{Uses: "myorg/build-docker-action@v1"}},
			},
			expected: false,
		},
		{
			name: "multiple steps - first has docker action",
			job: &Job{
				Steps: []Step{
					{Uses: "docker/build-push-action@v6"},
					{Uses: "actions/checkout@v4"},
				},
			},
			expected: true,
		},
		{
			name: "multiple steps - last has docker action",
			job: &Job{
				Steps: []Step{
					{Uses: "actions/checkout@v4"},
					{Uses: "docker/build-push-action@v6"},
				},
			},
			expected: true,
		},
		{
			name: "docker action in middle of steps",
			job: &Job{
				Steps: []Step{
					{Uses: "actions/checkout@v4"},
					{Uses: "docker/login-action@v3"},
					{Run: "echo done"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.HasContainerActions()
			if got != tt.expected {
				t.Errorf("HasContainerActions() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJob_HasServices_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		job      *Job
		expected bool
	}{
		{
			name: "nil services",
			job: &Job{
				Services: nil,
			},
			expected: false,
		},
		{
			name: "empty map",
			job: &Job{
				Services: map[string]any{},
			},
			expected: true, // Empty map is still defined
		},
		{
			name: "services with single service",
			job: &Job{
				Services: map[string]any{
					"postgres": map[string]any{
						"image": "postgres:14",
					},
				},
			},
			expected: true,
		},
		{
			name: "services with multiple services",
			job: &Job{
				Services: map[string]any{
					"postgres": map[string]any{
						"image": "postgres:14",
					},
					"redis": map[string]any{
						"image": "redis:7",
					},
				},
			},
			expected: true,
		},
		{
			name: "services with empty service config",
			job: &Job{
				Services: map[string]any{
					"postgres": map[string]any{},
				},
			},
			expected: true,
		},
		{
			name: "services as array",
			job: &Job{
				Services: []interface{}{"postgres", "redis"},
			},
			expected: true, // Non-nil means services are defined
		},
		{
			name: "services as string",
			job: &Job{
				Services: "postgres",
			},
			expected: true, // Non-nil means services are defined
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.HasServices()
			if got != tt.expected {
				t.Errorf("HasServices() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJob_HasContainer_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		job      *Job
		expected bool
	}{
		{
			name: "nil container",
			job: &Job{
				Container: nil,
			},
			expected: false,
		},
		{
			name: "container with image string",
			job: &Job{
				Container: "node:18",
			},
			expected: true,
		},
		{
			name: "container with image map",
			job: &Job{
				Container: map[string]any{
					"image": "node:18",
				},
			},
			expected: true,
		},
		{
			name: "container with image and options",
			job: &Job{
				Container: map[string]any{
					"image": "node:18",
					"env": map[string]string{
						"NODE_ENV": "test",
					},
					"ports": []int{3000},
				},
			},
			expected: true,
		},
		{
			name: "container with credentials",
			job: &Job{
				Container: map[string]any{
					"image": "ghcr.io/private/image:latest",
					"credentials": map[string]string{
						"username": "${{ secrets.USERNAME }}",
						"password": "${{ secrets.PASSWORD }}",
					},
				},
			},
			expected: true,
		},
		{
			name: "empty container map",
			job: &Job{
				Container: map[string]any{},
			},
			expected: true, // Non-nil means container is defined
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.HasContainer()
			if got != tt.expected {
				t.Errorf("HasContainer() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJob_CombinedChecks(t *testing.T) {
	tests := []struct {
		name           string
		job            *Job
		wantUbuntu     bool
		wantDockerCmd  bool
		wantDockerAct  bool
		wantServices   bool
		wantContainer  bool
	}{
		{
			name: "fully eligible job",
			job: &Job{
				RunsOn:    "ubuntu-latest",
				Steps:     []Step{{Run: "echo hello"}},
				Services:  nil,
				Container: nil,
			},
			wantUbuntu:    true,
			wantDockerCmd: false,
			wantDockerAct: false,
			wantServices:  false,
			wantContainer: false,
		},
		{
			name: "job with docker command",
			job: &Job{
				RunsOn:    "ubuntu-latest",
				Steps:     []Step{{Run: "docker build -t app ."}},
				Services:  nil,
				Container: nil,
			},
			wantUbuntu:    true,
			wantDockerCmd: true,
			wantDockerAct: false,
			wantServices:  false,
			wantContainer: false,
		},
		{
			name: "job with docker action",
			job: &Job{
				RunsOn:    "ubuntu-latest",
				Steps:     []Step{{Uses: "docker/build-push-action@v6"}},
				Services:  nil,
				Container: nil,
			},
			wantUbuntu:    true,
			wantDockerCmd: false,
			wantDockerAct: true,
			wantServices:  false,
			wantContainer: false,
		},
		{
			name: "job with services",
			job: &Job{
				RunsOn:    "ubuntu-latest",
				Steps:     []Step{{Run: "echo hello"}},
				Services: map[string]any{
					"postgres": map[string]any{},
				},
				Container: nil,
			},
			wantUbuntu:    true,
			wantDockerCmd: false,
			wantDockerAct: false,
			wantServices:  true,
			wantContainer: false,
		},
		{
			name: "job with container",
			job: &Job{
				RunsOn:    "ubuntu-latest",
				Steps:     []Step{{Run: "node --version"}},
				Services:  nil,
				Container: "node:18",
			},
			wantUbuntu:    true,
			wantDockerCmd: false,
			wantDockerAct: false,
			wantServices:  false,
			wantContainer: true,
		},
		{
			name: "job with all disqualifiers",
			job: &Job{
				RunsOn: "ubuntu-latest",
				Steps: []Step{
					{Run: "docker build -t app ."},
					{Uses: "docker/build-push-action@v6"},
				},
				Services: map[string]any{
					"postgres": map[string]any{},
				},
				Container: "node:18",
			},
			wantUbuntu:    true,
			wantDockerCmd: true,
			wantDockerAct: true,
			wantServices:  true,
			wantContainer: true,
		},
		{
			name: "non-ubuntu runner",
			job: &Job{
				RunsOn:    "ubuntu-22.04",
				Steps:     []Step{{Run: "echo hello"}},
				Services:  nil,
				Container: nil,
			},
			wantUbuntu:    false,
			wantDockerCmd: false,
			wantDockerAct: false,
			wantServices:  false,
			wantContainer: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.job.IsUbuntuLatest(); got != tt.wantUbuntu {
				t.Errorf("IsUbuntuLatest() = %v, want %v", got, tt.wantUbuntu)
			}
			if got := tt.job.HasDockerCommands(); got != tt.wantDockerCmd {
				t.Errorf("HasDockerCommands() = %v, want %v", got, tt.wantDockerCmd)
			}
			if got := tt.job.HasContainerActions(); got != tt.wantDockerAct {
				t.Errorf("HasContainerActions() = %v, want %v", got, tt.wantDockerAct)
			}
			if got := tt.job.HasServices(); got != tt.wantServices {
				t.Errorf("HasServices() = %v, want %v", got, tt.wantServices)
			}
			if got := tt.job.HasContainer(); got != tt.wantContainer {
				t.Errorf("HasContainer() = %v, want %v", got, tt.wantContainer)
			}
		})
	}
}

