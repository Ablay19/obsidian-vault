import { createLogger } from '../utils/logger';
import type { VideoStorageService, VideoUploadResult } from '../types';

const logger = createLogger({ level: 'info', component: 'video-storage' });

export interface EphemeralStorageConfig {
  rendererUrl: string;
  expirySeconds: number;
}

export class EphemeralVideoStorage implements VideoStorageService {
  private config: EphemeralStorageConfig;

  constructor(config: Partial<EphemeralStorageConfig> = {}) {
    this.config = {
      rendererUrl: config.rendererUrl || 'http://localhost:8080',
      expirySeconds: config.expirySeconds || 3600,
    };
  }

  async upload(videoPath: string, jobId: string): Promise<VideoUploadResult> {
    logger.info('Creating ephemeral URL for video', { jobId, rendererUrl: this.config.rendererUrl });

    try {
      const statusResponse = await fetch(`${this.config.rendererUrl}/status/${jobId}`);
      
      if (!statusResponse.ok) {
        return {
          success: false,
          error: 'Video not found on renderer',
        };
      }

      const status = await statusResponse.json() as Record<string, unknown>;

      if (status.status !== 'complete') {
        return {
          success: false,
          error: `Video not ready: ${status.status}`,
        };
      }

      const downloadUrl = `${this.config.rendererUrl}/download/${jobId}`;
      const expiresAt = new Date(Date.now() + this.config.expirySeconds * 1000);

      logger.info('Ephemeral URL created', {
        jobId,
        url: downloadUrl,
        expiresAt: expiresAt.toISOString(),
      });

      return {
        success: true,
        url: downloadUrl,
        key: jobId,
      };
    } catch (error) {
      logger.error('Failed to create ephemeral URL', error as Error, { jobId });
      return {
        success: false,
        error: (error as Error).message,
      };
    }
  }

  getUrl(key: string): string {
    return `${this.config.rendererUrl}/download/${key}`;
  }

  async delete(key: string): Promise<boolean> {
    try {
      const response = await fetch(`${this.config.rendererUrl}/cancel/${key}`, {
        method: 'POST',
      });
      return response.ok;
    } catch (error) {
      logger.error('Failed to delete video', error as Error, { key });
      return false;
    }
  }
}

export class MockVideoStorage implements VideoStorageService {
  private videos: Map<string, { url: string; expiresAt: Date }>;

  constructor() {
    this.videos = new Map();
  }

  async upload(videoPath: string, jobId: string): Promise<VideoUploadResult> {
    const url = `https://mock-storage.example.com/videos/${jobId}.mp4`;
    const expiresAt = new Date(Date.now() + 3600 * 1000);
    
    this.videos.set(jobId, { url, expiresAt });

    return {
      success: true,
      url,
      key: jobId,
    };
  }

  getUrl(key: string): string {
    return this.videos.get(key)?.url || `https://mock-storage.example.com/videos/${key}.mp4`;
  }

  async delete(key: string): Promise<boolean> {
    return this.videos.delete(key);
  }
}

export const createVideoStorage = (config?: Partial<EphemeralStorageConfig>): VideoStorageService => {
  return new EphemeralVideoStorage(config);
};
