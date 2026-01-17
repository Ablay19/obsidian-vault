package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/models"
)

// ServiceManager defines the interface for managing services
type ServiceManager interface {
	StartService(config models.ServiceConfig) error
	StopService(serviceID string) error
	GetServiceStatus(serviceID string) (*models.ServiceStatus, error)
	ListServices() ([]models.ServiceStatus, error)
	IsAvailable() bool
}

// DockerManager manages Docker-based services
type DockerManager struct {
	logger *log.Logger
}

// NewDockerManager creates a new Docker service manager
func NewDockerManager(logger *log.Logger) *DockerManager {
	return &DockerManager{logger: logger}
}

// IsAvailable checks if Docker is available on the system
func (dm *DockerManager) IsAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// StartService starts a Docker container for the given service configuration
func (dm *DockerManager) StartService(config models.ServiceConfig) error {
	if !dm.IsAvailable() {
		return fmt.Errorf("docker is not available on this system")
	}

	dm.logger.Printf("Starting Docker service: %s", config.Name)

	// Build docker run command
	args := []string{"run", "-d", "--name", config.Name}

	// Add port mappings
	if len(config.Ports) > 0 {
		for _, port := range config.Ports {
			args = append(args, "-p", fmt.Sprintf("%d:%d", port.Host, port.Container))
		}
	}

	// Add environment variables
	if len(config.Environment) > 0 {
		for key, value := range config.Environment {
			args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Add image
	args = append(args, config.Image)

	// Add command if specified
	if len(config.Command) > 0 {
		args = append(args, config.Command...)
	}

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start Docker container: %w, output: %s", err, string(output))
	}

	dm.logger.Printf("Successfully started Docker service: %s", config.Name)
	return nil
}

// StopService stops a Docker container
func (dm *DockerManager) StopService(serviceID string) error {
	if !dm.IsAvailable() {
		return fmt.Errorf("docker is not available on this system")
	}

	dm.logger.Printf("Stopping Docker service: %s", serviceID)

	cmd := exec.Command("docker", "stop", serviceID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop Docker container: %w, output: %s", err, string(output))
	}

	// Remove the container
	cmd = exec.Command("docker", "rm", serviceID)
	if err := cmd.Run(); err != nil {
		dm.logger.Printf("Warning: failed to remove container %s: %v", serviceID, err)
	}

	dm.logger.Printf("Successfully stopped Docker service: %s", serviceID)
	return nil
}

// GetServiceStatus returns the status of a Docker container
func (dm *DockerManager) GetServiceStatus(serviceID string) (*models.ServiceStatus, error) {
	if !dm.IsAvailable() {
		return nil, fmt.Errorf("docker is not available on this system")
	}

	cmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", serviceID), "--format", "{{.Names}},{{.Status}},{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get container status: %w", err)
	}

	lines := strings.TrimSpace(string(output))
	if lines == "" {
		return &models.ServiceStatus{
			ServiceID:     serviceID,
			State:         "stopped",
			Health:        models.HealthStatus{Status: "unknown"},
			LastSeen:      time.Now(),
			Uptime:        0,
			ResourceUsage: models.ResourceUsage{CPU: "0%", Memory: "0MB"},
		}, nil
	}

	parts := strings.Split(lines, ",")
	if len(parts) < 2 {
		return nil, fmt.Errorf("unexpected docker ps output format")
	}

	status := dm.parseDockerStatus(parts[1])

	return &models.ServiceStatus{
		ServiceID:     serviceID,
		State:         status,
		Health:        models.HealthStatus{Status: "healthy"}, // TODO: implement actual health checks
		LastSeen:      time.Now(),
		Uptime:        0,                                               // TODO: calculate uptime
		ResourceUsage: models.ResourceUsage{CPU: "N/A", Memory: "N/A"}, // TODO: get actual resource usage
	}, nil
}

// ListServices returns all Docker containers managed by this tool
func (dm *DockerManager) ListServices() ([]models.ServiceStatus, error) {
	if !dm.IsAvailable() {
		return nil, fmt.Errorf("docker is not available on this system")
	}

	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}},{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var services []models.ServiceStatus
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 2 {
			status := dm.parseDockerStatus(parts[1])
			services = append(services, models.ServiceStatus{
				ServiceID:     parts[0],
				State:         status,
				Health:        models.HealthStatus{Status: "healthy"},
				LastSeen:      time.Now(),
				Uptime:        0,
				ResourceUsage: models.ResourceUsage{CPU: "N/A", Memory: "N/A"},
			})
		}
	}

	return services, nil
}

// parseDockerStatus parses Docker container status string
func (dm *DockerManager) parseDockerStatus(status string) string {
	status = strings.ToLower(status)
	if strings.Contains(status, "up") {
		return "running"
	} else if strings.Contains(status, "exited") {
		return "stopped"
	} else if strings.Contains(status, "created") {
		return "created"
	}
	return "unknown"
}

// ProcessManager manages local processes (non-Docker)
type ProcessManager struct {
	logger    *log.Logger
	processes map[string]*exec.Cmd
}

// NewProcessManager creates a new process service manager
func NewProcessManager(logger *log.Logger) *ProcessManager {
	return &ProcessManager{
		logger:    logger,
		processes: make(map[string]*exec.Cmd),
	}
}

// IsAvailable checks if process management is available (always true for local processes)
func (pm *ProcessManager) IsAvailable() bool {
	return true
}

// StartService starts a local process
func (pm *ProcessManager) StartService(config models.ServiceConfig) error {
	pm.logger.Printf("Starting process service: %s", config.Name)

	if len(config.Command) == 0 {
		return fmt.Errorf("no command specified for service %s", config.Name)
	}

	cmd := exec.Command(config.Command[0], config.Command[1:]...)

	// Set environment variables
	if len(config.Environment) > 0 {
		env := os.Environ()
		for key, value := range config.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Start the process asynchronously
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	pm.processes[config.Name] = cmd

	// Monitor the process in a goroutine
	go func() {
		err := cmd.Wait()
		if err != nil {
			pm.logger.Printf("Process %s exited with error: %v", config.Name, err)
		} else {
			pm.logger.Printf("Process %s exited normally", config.Name)
		}
		delete(pm.processes, config.Name)
	}()

	pm.logger.Printf("Successfully started process service: %s", config.Name)
	return nil
}

// StopService stops a local process
func (pm *ProcessManager) StopService(serviceID string) error {
	pm.logger.Printf("Stopping process service: %s", serviceID)

	cmd, exists := pm.processes[serviceID]
	if !exists {
		return fmt.Errorf("process %s not found", serviceID)
	}

	// Try graceful shutdown first
	if runtime.GOOS == "windows" {
		cmd.Process.Kill() // Windows doesn't support SIGTERM well
	} else {
		cmd.Process.Kill() // TODO: implement graceful shutdown with SIGTERM
	}

	delete(pm.processes, serviceID)
	pm.logger.Printf("Successfully stopped process service: %s", serviceID)
	return nil
}

// GetServiceStatus returns the status of a local process
func (pm *ProcessManager) GetServiceStatus(serviceID string) (*models.ServiceStatus, error) {
	cmd, exists := pm.processes[serviceID]

	state := "stopped"
	if exists && cmd.Process != nil {
		// Check if process is still running
		if cmd.ProcessState == nil {
			state = "running"
		}
	}

	return &models.ServiceStatus{
		ServiceID:     serviceID,
		State:         state,
		Health:        models.HealthStatus{Status: "healthy"},
		LastSeen:      time.Now(),
		Uptime:        0, // TODO: calculate uptime
		ResourceUsage: models.ResourceUsage{CPU: "N/A", Memory: "N/A"},
	}, nil
}

// ListServices returns all managed processes
func (pm *ProcessManager) ListServices() ([]models.ServiceStatus, error) {
	var services []models.ServiceStatus
	for serviceID := range pm.processes {
		status, err := pm.GetServiceStatus(serviceID)
		if err != nil {
			continue
		}
		services = append(services, *status)
	}
	return services, nil
}
