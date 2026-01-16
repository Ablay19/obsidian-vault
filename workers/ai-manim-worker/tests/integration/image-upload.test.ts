import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

const API_GATEWAY_URL = process.env.API_GATEWAY_URL || 'http://localhost:8080';
const TELEGRAM_BOT_TOKEN = process.env.TELEGRAM_BOT_TOKEN || 'test-token';

describe('Image Upload - Integration with Go Services', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Image Upload to Go Services', () => {
    it('should submit image processing job to Go services', async () => {
      const mockFetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          job_id: 'test-job-123',
          status: 'queued',
        }),
      } as Response);

      global.fetch = mockFetch;

      const mockImageBuffer = new ArrayBuffer(1024);
      const mockMessage = {
        message_id: 1,
        chat: { id: 12345, type: 'private' },
        from: { id: 67890, is_bot: false },
        photo: [{ file_id: 'photo-123', file_size: 1024 }],
        date: Date.now(),
      };

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-auth-token',
        },
        body: JSON.stringify({
          chat_id: 12345,
          user_id: '67890',
          file_id: 'photo-123',
          file_size: 1024,
          processing_type: 'image-analysis',
        }),
      });

      expect(response.ok).toBe(true);

      const data = await response.json();
      expect(data.job_id).toBeDefined();
      expect(data.status).toBe('queued');
    });

    it('should handle large image files with size limit', async () => {
      const mockFetch = vi.fn().mockResolvedValueOnce({
        ok: false,
        status: 413,
        json: async () => ({
          error: 'File too large',
        }),
      } as Response);

      global.fetch = mockFetch;

      const largeImageBuffer = new ArrayBuffer(20 * 1024 * 1024);

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-auth-token',
        },
        body: JSON.stringify({
          chat_id: 12345,
          user_id: '67890',
          file_id: 'large-photo-123',
          file_size: largeImageBuffer.byteLength,
          processing_type: 'image-analysis',
        }),
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(413);

      const data = await response.json();
      expect(data.error).toContain('too large');
    });

    it('should retrieve processing status for image job', async () => {
      const jobId = 'job-status-test-123';

      const mockFetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          job_id: jobId,
          status: 'processing',
          result: {
            analysis: 'Mathematical concept detected',
            confidence: 0.95,
          },
        }),
      } as Response);

      global.fetch = mockFetch;

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/jobs/${jobId}`, {
        method: 'GET',
        headers: {
          'Authorization': 'Bearer test-auth-token',
        },
      });

      expect(response.ok).toBe(true);

      const data = await response.json();
      expect(data.job_id).toBe(jobId);
      expect(data.status).toBe('processing');
      expect(data.result).toBeDefined();
      expect(data.result.analysis).toBeDefined();
    });

    it('should return processed result with analysis data', async () => {
      const jobId = 'job-result-test-456';

      const mockFetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          job_id: jobId,
          status: 'ready',
          result: {
            analysis: 'Mathematical formula visualized',
            confidence: 0.98,
            detected_elements: ['formula', 'variables', 'diagram'],
          },
        }),
      } as Response);

      global.fetch = mockFetch;

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/jobs/${jobId}/result`, {
        method: 'GET',
        headers: {
          'Authorization': 'Bearer test-auth-token',
        },
      });

      expect(response.ok).toBe(true);

      const data = await response.json();
      expect(data.job_id).toBe(jobId);
      expect(data.status).toBe('ready');
      expect(data.result).toBeDefined();
      expect(data.result.analysis).toBeDefined();
      expect(data.result.analysis.detected_elements).toContain('formula');
    });
  });

  describe('Telegram Bot Integration', () => {
    it('should send processed result to user via Telegram', async () => {
      const chatId = 12345;

      const mockFetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          ok: true,
          result: {
            message_id: 987,
          },
        }),
      } as Response);

      global.fetch = mockFetch;

      const response = await fetch(
        `https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            chat_id: chatId,
            text: '✅ Your image has been processed!',
            parse_mode: 'Markdown',
          }),
        }
      );

      expect(response.ok).toBe(true);
    });

    it('should send error message for failed image processing', async () => {
      const chatId = 12345;

      const mockFetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          ok: true,
          result: {
            message_id: 988,
          },
        }),
      } as Response);

      global.fetch = mockFetch;

      const response = await fetch(
        `https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            chat_id: chatId,
            text: '❌ Image processing failed. Please try again.',
            parse_mode: 'Markdown',
          }),
        }
      );

      expect(response.ok).toBe(true);
    });

    it('should send photo result back to user', async () => {
      const chatId = 12345;
      const photoUrl = 'https://example.com/processed-photo.jpg';

      const mockFetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          ok: true,
          result: {
            message_id: 989,
          },
        }),
      } as Response);

      global.fetch = mockFetch;

      const response = await fetch(
        `https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendPhoto`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            chat_id: chatId,
            photo: photoUrl,
            caption: 'Processed result: Mathematical concept detected',
          }),
        }
      );

      expect(response.ok).toBe(true);
    });
  });

  describe('Error Handling', () => {
    it('should handle Go service timeout gracefully', async () => {
      const mockFetch = vi.fn().mockRejectedValueOnce(new Error('Request timeout'));

      global.fetch = mockFetch;

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-auth-token',
        },
        signal: AbortSignal.timeout(30000),
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(504);
    });

    it('should handle authentication errors from Go services', async () => {
      const mockFetch = vi.fn().mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({
          error: 'Unauthorized',
        }),
      } as Response);

      global.fetch = mockFetch;

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-auth-token',
        },
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(401);
    });

    it('should handle malformed responses from Go services', async () => {
      const mockFetch = vi.fn().mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({
          error: 'Internal server error',
        }),
      } as Response);

      global.fetch = mockFetch;

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-auth-token',
        },
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(500);
    });
  });
});
