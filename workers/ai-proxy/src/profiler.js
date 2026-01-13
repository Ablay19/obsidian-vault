// Performance Profiler for Cloudflare Workers
export class WorkerProfiler {
  constructor(options = {}) {
    this.enabled = options.enabled !== false;
    this.maxProfiles = options.maxProfiles || 1000;
    this.profiles = new Map();
    this.completedProfiles = [];
    this.currentProfile = null;
  }

  // Start profiling a named operation
  startProfile(name, metadata = {}) {
    if (!this.enabled) return;

    const profileId = `${name}-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

    this.currentProfile = {
      id: profileId,
      name,
      startTime: performance.now(),
      startMemory: this.getMemoryUsage(),
      metadata,
      checkpoints: []
    };

    this.profiles.set(profileId, this.currentProfile);

    // Cleanup old profiles if we exceed limit
    if (this.profiles.size > this.maxProfiles) {
      this.cleanupOldProfiles();
    }

    return profileId;
  }

  // Add a checkpoint to the current profile
  addCheckpoint(label, data = {}) {
    if (!this.enabled || !this.currentProfile) return;

    const checkpoint = {
      label,
      timestamp: performance.now(),
      memoryDelta: this.getMemoryUsage() - this.currentProfile.startMemory,
      data
    };

    this.currentProfile.checkpoints.push(checkpoint);
  }

  // End profiling and return results
  endProfile(profileId) {
    if (!this.enabled) return null;

    let profile;
    if (profileId) {
      profile = this.profiles.get(profileId);
    } else {
      profile = this.currentProfile;
    }

    if (!profile) return null;

    const endTime = performance.now();
    const endMemory = this.getMemoryUsage();

    const result = {
      id: profile.id,
      name: profile.name,
      duration: endTime - profile.startTime,
      memoryDelta: endMemory - profile.startMemory,
      startTime: profile.startTime,
      endTime: endTime,
      checkpoints: profile.checkpoints,
      metadata: profile.metadata,
      memoryPeak: this.calculateMemoryPeak(profile),
      efficiency: this.calculateEfficiency(profile, endTime - profile.startTime)
    };

    // Store completed profile
    this.completedProfiles.push(result);
    if (this.completedProfiles.length > this.maxProfiles) {
      this.completedProfiles.shift();
    }

    // Remove from active profiles
    this.profiles.delete(profile.id);
    this.currentProfile = null;

    return result;
  }

  // Get memory usage (Cloudflare Workers specific)
  getMemoryUsage() {
    // In Cloudflare Workers, we don't have direct memory access
    // This is a placeholder for actual memory tracking
    try {
      // Use performance.memory if available (not standard in CF Workers)
      if (performance.memory) {
        return performance.memory.usedJSHeapSize || 0;
      }
    } catch (e) {
      // Memory API not available
    }

    // Fallback: estimate based on operation count or return 0
    return 0;
  }

  // Calculate memory peak from checkpoints
  calculateMemoryPeak(profile) {
    let peak = profile.startMemory || 0;

    for (const checkpoint of profile.checkpoints) {
      const checkpointMemory = (profile.startMemory || 0) + checkpoint.memoryDelta;
      if (checkpointMemory > peak) {
        peak = checkpointMemory;
      }
    }

    return peak;
  }

  // Calculate efficiency score (lower is better)
  calculateEfficiency(profile, totalDuration) {
    const memoryEfficiency = profile.memoryDelta / Math.max(totalDuration, 1);
    const checkpointEfficiency = profile.checkpoints.length / Math.max(totalDuration / 1000, 1); // checkpoints per second

    // Weighted score
    return (memoryEfficiency * 0.7) + (checkpointEfficiency * 0.3);
  }

  // Get profiling statistics
  getStats() {
    const completedCount = this.completedProfiles.length;
    const activeCount = this.profiles.size;

    if (completedCount === 0) {
      return {
        enabled: this.enabled,
        activeProfiles: activeCount,
        completedProfiles: 0,
        avgDuration: 0,
        avgMemoryDelta: 0,
        totalProfiles: 0
      };
    }

    const totalDuration = this.completedProfiles.reduce((sum, p) => sum + p.duration, 0);
    const totalMemoryDelta = this.completedProfiles.reduce((sum, p) => sum + p.memoryDelta, 0);

    return {
      enabled: this.enabled,
      activeProfiles: activeCount,
      completedProfiles: completedCount,
      avgDuration: totalDuration / completedCount,
      avgMemoryDelta: totalMemoryDelta / completedCount,
      totalProfiles: activeCount + completedCount
    };
  }

  // Get recent profiles
  getRecentProfiles(limit = 10) {
    return this.completedProfiles.slice(-limit);
  }

  // Get profile by ID
  getProfile(profileId) {
    return this.profiles.get(profileId) ||
           this.completedProfiles.find(p => p.id === profileId);
  }

  // Export profiling data
  exportProfiles(format = 'json') {
    const data = {
      stats: this.getStats(),
      activeProfiles: Array.from(this.profiles.values()),
      completedProfiles: this.completedProfiles,
      exportTime: Date.now()
    };

    switch (format) {
      case 'json':
        return JSON.stringify(data, null, 2);
      case 'csv':
        return this.convertToCSV(data.completedProfiles);
      default:
        return JSON.stringify(data);
    }
  }

  // Convert profiles to CSV format
  convertToCSV(profiles) {
    if (profiles.length === 0) return 'No profiles available';

    const headers = ['ID', 'Name', 'Duration', 'Memory Delta', 'Checkpoints', 'Start Time', 'End Time'];
    let csv = headers.join(',') + '\n';

    for (const profile of profiles) {
      const row = [
        profile.id,
        profile.name,
        profile.duration.toFixed(2),
        profile.memoryDelta,
        profile.checkpoints.length,
        new Date(profile.startTime).toISOString(),
        new Date(profile.endTime).toISOString()
      ];
      csv += row.join(',') + '\n';
    }

    return csv;
  }

  // Cleanup old profiles
  cleanupOldProfiles() {
    const cutoffTime = Date.now() - (24 * 60 * 60 * 1000); // 24 hours ago

    // Remove old completed profiles
    this.completedProfiles = this.completedProfiles.filter(
      profile => profile.endTime > cutoffTime
    );

    // Remove very old active profiles (shouldn't happen normally)
    for (const [id, profile] of this.profiles) {
      if (profile.startTime < cutoffTime - (60 * 60 * 1000)) { // 1 hour grace period
        this.profiles.delete(id);
      }
    }
  }

  // Enable or disable profiling
  setEnabled(enabled) {
    this.enabled = enabled;
  }

  // Clear all profiling data
  clear() {
    this.profiles.clear();
    this.completedProfiles.length = 0;
    this.currentProfile = null;
  }
}

// Utility function for timing function execution
export function timeFunction(fn, name, profiler) {
  const profileId = profiler ? profiler.startProfile(name) : null;

  try {
    const result = fn();
    if (result && typeof result.then === 'function') {
      // Handle promises
      return result.finally(() => {
        if (profiler && profileId) {
          profiler.endProfile(profileId);
        }
      });
    } else {
      // Handle synchronous functions
      if (profiler && profileId) {
        profiler.endProfile(profileId);
      }
      return result;
    }
  } catch (error) {
    if (profiler && profileId) {
      profiler.endProfile(profileId);
    }
    throw error;
  }
}

// Decorator for profiling class methods
export function profileMethod(profiler, methodName) {
  return function(target, propertyKey, descriptor) {
    const originalMethod = descriptor.value;

    descriptor.value = function(...args) {
      const profileName = `${methodName}-${propertyKey}`;
      const profileId = profiler ? profiler.startProfile(profileName) : null;

      try {
        const result = originalMethod.apply(this, args);
        if (profiler && profileId) {
          profiler.endProfile(profileId);
        }
        return result;
      } catch (error) {
        if (profiler && profileId) {
          profiler.endProfile(profileId);
        }
        throw error;
      }
    };

    return descriptor;
  };
}