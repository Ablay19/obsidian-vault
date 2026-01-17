package utils

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PlatformType represents the detected platform
type PlatformType string

const (
	PlatformDesktopLinux   PlatformType = "desktop-linux"
	PlatformDesktopMacOS   PlatformType = "desktop-macos"
	PlatformDesktopWindows PlatformType = "desktop-windows"
	PlatformTermux         PlatformType = "termux"
	PlatformWSL            PlatformType = "wsl"
	PlatformUnknown        PlatformType = "unknown"
)

// PlatformInfo contains platform detection results
type PlatformInfo struct {
	Type           PlatformType
	IsMobile       bool
	IsTermux       bool
	HasDocker      bool
	HasKubectl     bool
	ResourceLimits ResourceLimits
}

// ResourceLimits defines resource constraints for different platforms
type ResourceLimits struct {
	MaxMemoryMB       int
	MaxCPU            float64
	NetworkTimeoutSec int
}

// DetectPlatform detects the current platform and capabilities
func DetectPlatform() *PlatformInfo {
	info := &PlatformInfo{
		Type:     detectPlatformType(),
		IsMobile: false,
		IsTermux: isTermux(),
	}

	info.IsMobile = info.IsTermux

	// Check for available tools
	info.HasDocker = hasCommand("docker")
	info.HasKubectl = hasCommand("kubectl")

	// Set resource limits based on platform
	info.ResourceLimits = getResourceLimits(info)

	return info
}

// detectPlatformType determines the specific platform type
func detectPlatformType() PlatformType {
	if isTermux() {
		return PlatformTermux
	}

	if isWSL() {
		return PlatformWSL
	}

	switch runtime.GOOS {
	case "linux":
		return PlatformDesktopLinux
	case "darwin":
		return PlatformDesktopMacOS
	case "windows":
		return PlatformDesktopWindows
	default:
		return PlatformUnknown
	}
}

// isTermux checks if running in Termux environment
func isTermux() bool {
	// Check for Termux-specific environment variables
	if os.Getenv("TERMUX_VERSION") != "" {
		return true
	}

	// Check for Termux-specific paths
	termuxPaths := []string{
		"/data/data/com.termux",
		"$PREFIX",
	}

	for _, path := range termuxPaths {
		if expanded := os.ExpandEnv(path); expanded != path {
			if _, err := os.Stat(expanded); err == nil {
				return true
			}
		}
	}

	// Check for Android-specific properties
	if _, err := os.Stat("/system/build.prop"); err == nil {
		return true
	}

	return false
}

// isWSL checks if running in Windows Subsystem for Linux
func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	// Check for WSL-specific files
	wslFiles := []string{
		"/proc/version",
		"/proc/sys/kernel/osrelease",
	}

	for _, file := range wslFiles {
		if content, err := os.ReadFile(file); err == nil {
			if strings.Contains(strings.ToLower(string(content)), "microsoft") ||
				strings.Contains(strings.ToLower(string(content)), "wsl") {
				return true
			}
		}
	}

	return false
}

// hasCommand checks if a command is available in PATH
func hasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// getResourceLimits returns appropriate resource limits for the platform
func getResourceLimits(info *PlatformInfo) ResourceLimits {
	switch info.Type {
	case PlatformTermux:
		return ResourceLimits{
			MaxMemoryMB:       512, // Termux typically has 1-2GB RAM, limit to 512MB
			MaxCPU:            0.5, // Limit to 50% CPU to preserve battery
			NetworkTimeoutSec: 30,  // Mobile networks can be slow
		}
	case PlatformWSL:
		return ResourceLimits{
			MaxMemoryMB:       2048, // WSL can access host RAM but limit for performance
			MaxCPU:            0.8,  // Allow higher CPU usage
			NetworkTimeoutSec: 10,   // Desktop network is faster
		}
	case PlatformDesktopLinux, PlatformDesktopMacOS:
		return ResourceLimits{
			MaxMemoryMB:       4096, // Allow higher memory usage on desktop
			MaxCPU:            1.0,  // Full CPU access
			NetworkTimeoutSec: 10,   // Fast network
		}
	case PlatformDesktopWindows:
		return ResourceLimits{
			MaxMemoryMB:       2048, // Windows might have more constraints
			MaxCPU:            0.8,  // Reserve some for system
			NetworkTimeoutSec: 10,
		}
	default:
		return ResourceLimits{
			MaxMemoryMB:       1024,
			MaxCPU:            0.5,
			NetworkTimeoutSec: 15,
		}
	}
}

// IsLowResource returns true if the platform has limited resources
func (pi *PlatformInfo) IsLowResource() bool {
	return pi.Type == PlatformTermux || pi.ResourceLimits.MaxMemoryMB < 1024
}

// ShouldUseMobileOptimizations returns true if mobile-specific optimizations should be used
func (pi *PlatformInfo) ShouldUseMobileOptimizations() bool {
	return pi.IsMobile || pi.IsLowResource()
}

// GetRecommendedConcurrency returns recommended number of concurrent operations
func (pi *PlatformInfo) GetRecommendedConcurrency() int {
	if pi.IsLowResource() {
		return 1 // Sequential execution on low-resource platforms
	}
	return 3 // Parallel execution on desktop
}
