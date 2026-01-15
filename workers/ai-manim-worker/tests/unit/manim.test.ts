import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ManimRendererService } from '../../src/services/manim';

describe('ManimRendererService', () => {
  let service: ManimRendererService;

  beforeEach(() => {
    service = new ManimRendererService({
      endpoint: 'http://localhost:8080',
      timeout: 30000,
      maxRetries: 3,
    });
  });

  describe('validateCode', () => {
    it('should accept valid Manim code', () => {
      const code = `
from manim import *

class Scene(Scene):
    def construct(self):
        pass
      `;

      const result = service.validateCode(code);

      expect(result.valid).toBe(true);
    });

    it('should reject empty code', () => {
      const result = service.validateCode('');

      expect(result.valid).toBe(false);
      expect(result.error).toBe('Code is empty');
    });

    it('should reject code without manim import', () => {
      const code = `
class Scene(Scene):
    def construct(self):
        pass
      `;

      const result = service.validateCode(code);

      expect(result.valid).toBe(false);
      expect(result.error).toBe('Code must import manim');
    });

    it('should reject code without Scene class', () => {
      const code = `
from manim import *

def main():
    print("Hello")
      `;

      const result = service.validateCode(code);

      expect(result.valid).toBe(false);
      expect(result.error).toBe('Code must define a Scene class');
    });

    it('should reject code without manim import', () => {
      const code = `
class Scene(Scene):
    def construct(self):
        pass
      `;

      const result = service.validateCode(code);

      expect(result.valid).toBe(false);
      expect(result.error).toBe('Code must import manim');
    });
  });

  describe('submitRender', () => {
    it('should submit render job successfully', async () => {
      const mockFetch = vi.spyOn(global, 'fetch').mockResolvedValueOnce({
        ok: true,
        json: async () => ({ job_id: 'test-123', status: 'queued' }),
      } as Response);

      const request = {
        jobId: 'test-123',
        code: 'from manim import *\nclass Scene(Scene):\n    pass',
        problem: 'Test problem',
      };

      const result = await service.submitRender(request);

      expect(result.jobId).toBe('test-123');
      expect(result.status).toBe('queued');
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/render',
        expect.objectContaining({
          method: 'POST',
          body: expect.any(String),
        })
      );

      mockFetch.mockRestore();
    });

    it('should handle fetch errors', async () => {
      const mockFetch = vi.spyOn(global, 'fetch').mockRejectedValueOnce(new Error('Network error'));

      const request = {
        jobId: 'test-123',
        code: 'from manim import *\nclass Scene(Scene):\n    pass',
        problem: 'Test problem',
      };

      const result = await service.submitRender(request);

      expect(result.status).toBe('failed');
      expect(result.error).toBe('Network error');

      mockFetch.mockRestore();
    });
  });

  describe('getStatus', () => {
    it('should return job status', async () => {
      await service.submitRender({
        jobId: 'test-123',
        code: 'from manim import *\nclass Scene(Scene):\n    pass',
        problem: 'Test problem',
      });

      const mockFetch = vi.spyOn(global, 'fetch').mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          job_id: 'test-123',
          status: 'rendering',
        }),
      } as Response);

      const result = await service.getStatus('test-123');

      expect(result.jobId).toBe('test-123');
      expect(result.status).toBe('rendering');

      mockFetch.mockRestore();
    });

    it('should return not found for unknown job', async () => {
      const mockFetch = vi.spyOn(global, 'fetch').mockResolvedValueOnce({
        ok: true,
        json: async () => ({ job_id: 'unknown', status: 'failed' }),
      } as Response);

      const result = await service.getStatus('unknown');

      expect(result.status).toBe('failed');
      expect(result.error).toBe('Job not found');

      mockFetch.mockRestore();
    });
  });

  describe('cancelRender', () => {
    it('should cancel render job', async () => {
      const mockFetch = vi.spyOn(global, 'fetch').mockResolvedValueOnce({
        ok: true,
      } as Response);

      const result = await service.cancelRender('test-123');

      expect(result).toBe(true);
      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/cancel/test-123',
        expect.objectContaining({ method: 'POST' })
      );

      mockFetch.mockRestore();
    });
  });
});
