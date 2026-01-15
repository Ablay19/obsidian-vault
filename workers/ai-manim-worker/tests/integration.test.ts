import { describe, it, expect, beforeAll, afterAll } from 'vitest';

const WORKER_URL = process.env.WORKER_URL || 'https://ai-manim-worker-staging.abdoullahelvogani.workers.dev';
const TELEGRAM_BOT_TOKEN = process.env.TELEGRAM_BOT_TOKEN || '8220404781:AAEGmTzJqLD_xBiA7gB9zIX5Ax5JIuxvnRA';

describe('Integration Tests', () => {
  describe('Worker Health', () => {
    it('should return healthy status', async () => {
      const response = await fetch(`${WORKER_URL}/health`);
      expect(response.ok).toBe(true);
      
      const data = await response.json();
      expect(data.status).toBe('healthy');
      expect(data.version).toBeDefined();
    });
  });

  describe('Telegram Webhook', () => {
    it('should process valid webhook requests', async () => {
      const response = await fetch(`${WORKER_URL}/telegram/webhook`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          update_id: 1,
          message: {
            message_id: 1,
            chat: { id: 12345, type: 'private' },
            text: '/start',
            date: Date.now(),
          },
        }),
      });
      
      expect(response.ok).toBe(true);
    });

    it('should return 404 for unknown routes', async () => {
      const response = await fetch(`${WORKER_URL}/unknown`);
      expect(response.status).toBe(404);
    });
  });

  describe('Video Handler', () => {
    it('should accept valid render submission', async () => {
      const response = await fetch(`${WORKER_URL}/video/submit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jobId: `test-${Date.now()}`,
          manimCode: `from manim import *
class Scene(Scene):
    def construct(self):
        self.add(Text("Hello World"))
`,
          problem: 'Test integration',
          userId: 'test-user',
          chatId: 12345,
        }),
      });
      
      expect(response.ok).toBe(true);
      
      const data = await response.json();
      expect(data.jobId).toBeDefined();
      expect(data.status).toBe('queued');
    });
  });

  describe('Telegram API', () => {
    it('should return bot info', async () => {
      const response = await fetch(
        `https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getMe`
      );
      
      expect(response.ok).toBe(true);
      
      const data = await response.json();
      expect(data.ok).toBe(true);
      expect(data.result).toBeDefined();
      expect(data.result.id).toBeDefined();
    });

    it('should return webhook info', async () => {
      const response = await fetch(
        `https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getWebhookInfo`
      );
      
      expect(response.ok).toBe(true);
      
      const data = await response.json();
      expect(data.ok).toBe(true);
      expect(data.result.url).toBeDefined();
    });
  });

  describe('Full Pipeline Simulation', () => {
    it('should process a complete request flow', async () => {
      const testProblem = 'Explain the Pythagorean theorem';
      const testJobId = `integration-test-${Date.now()}`;
      
      const response = await fetch(`${WORKER_URL}/video/submit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jobId: testJobId,
          manimCode: `from manim import *
class PythagoreanScene(Scene):
    def construct(self):
        self.add(Text("${testProblem.replace(/"/g, '\\"')}"))
`,
          problem: testProblem,
          userId: 'integration-test',
          chatId: 12345,
        }),
      });
      
      expect(response.ok).toBe(true);
      
      const data = await response.json();
      expect(data.jobId).toBe(testJobId);
      expect(data.status).toBe('queued');
      expect(data.estimatedTime).toBeDefined();
    });
  });
});
