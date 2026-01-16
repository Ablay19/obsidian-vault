import { describe, it, expect, beforeAll, afterEach } from 'vitest';

describe('Telegram Media Upload - Integration with Go Services', () => {
  const API_GATEWAY_URL = process.env.API_GATEWAY_URL || 'http://localhost:8080';
  const TELEGRAM_BOT_TOKEN = process.env.TELEGRAM_BOT_TOKEN || 'test-token';

  describe('Image Upload to Go Services', () => {
    it('should submit image processing job to Go services', async () => {
      const mockImageBuffer = new ArrayBuffer(1024);
      const mockMessage = {
        message_id: 1,
        chat: { id: 12345, type: 'private' },
        from: { id: 67890, is_bot: false },
        photo: [{ file_id: 'photo_123', file_size: 1024 }],
        date: Date.now(),
      };

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer test-auth-token`,
        },
        body: JSON.stringify({
          chat_id: 12345,
          user_id: '67890',
          file_id: 'photo_123',
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
      const largeImageBuffer = new ArrayBuffer(20 * 1024 * 1024); // 20MB

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer test-auth-token`,
        },
        body: JSON.stringify({
          chat_id: 12345,
          user_id: '67890',
          file_id: 'large_photo_123',
          file_size: largeImageBuffer.byteLength,
          processing_type: 'image-analysis',
        }),
      });

      expect(response.status).toBe(413); // Payload Too Large
      expect(response.ok).toBe(false);

      const data = await response.json();
      expect(data.error).toContain('too large');
    });

    it('should retrieve processing status for image job', async () => {
      const jobId = 'job-image-test-123';

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/jobs/${jobId}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer test-auth-token`,
        },
      });

      expect(response.ok).toBe(true);

      const data = await response.json();
      expect(data.job_id).toBe(jobId);
      expect(data.status).toBeDefined();
    });

    it('should return processed result with analysis data', async () => {
      const jobId = 'job-image-result-456';

      const response = await fetch(`${API_GATEWAY_URL}/api/v1/jobs/${jobId}/result`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer test-auth-token`,
        },
      });

      expect(response.ok).toBe(true);

      const data = await response.json();
      expect(data.status).toBe('ready');
      expect(data.result).toBeDefined();
      expect(data.result.analysis).toBeDefined();
    });
  });

  describe('Telegram Bot Integration', () => {
    it('should send processed result to user via Telegram', async () => {
      const chatId = 12345;

      const response = await fetch(
        `https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            chat_id: chatId,
            text: '✅ *Your image has been processed!*',
            parse_mode: 'Markdown',
          }),
        }
      );

      expect(response.ok).toBe(true);
    });

    it('should send error message for failed image processing', async () => {
      const chatId = 12345;

      const response = await fetch(
        `https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          batch: [
            {
              chat_id: chatId,
              text: '❌ *Image processing failed.*',
            },
            {
              chat_id: chatId,
              text: 'Please try again with a different image.',
            },
          ],
        }
      );

      expect(response.ok).toBe(true);
    });

    it('should send photo result back to user', async () => {
      const chatId = 12345;
      const photoUrl = 'https://example.com/processed-photo.jpg';

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
      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer test-auth-token`,
        },
        signal: AbortSignal.timeout(30000),
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(504);
    });

    it('should handle authentication errors from Go services', async () => {
      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer invalid-token',
        },
      });

      expect(response.status).toBe(401);
    });

    it('should handle malformed responses from Go services', async () => {
      const response = await fetch(`${API_GATEWAY_URL}/api/v1/image/process`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer test-auth-token`,
        },
      });

      expect(response.status).toBe(500);
    });
  });
});
