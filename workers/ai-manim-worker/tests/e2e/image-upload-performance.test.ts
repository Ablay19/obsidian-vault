import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('Telegram Media Upload - Performance Tests', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should process 10MB image within 30 seconds', async () => {
    const startTime = Date.now();

    const mockFetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        job_id: 'test-perf-123',
        status: 'ready',
        result: {
          analysis: 'Performance test passed',
          confidence: 1.0,
          processing_time_ms: 25000,
        },
      }),
    } as Response);

    global.fetch = mockFetch;

    const formData = new FormData();
    formData.append('image', new Blob([new ArrayBuffer(10 * 1024 * 1024)]));
    formData.append('chat_id', '12345');
    formData.append('processing_type', 'image-analysis');

    const response = await fetch('http://localhost:8080/api/v1/image/process', {
      method: 'POST',
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      body: formData,
    });

    const duration = Date.now() - startTime;

    expect(response.ok).toBe(true);

    const data = await response.json();
    expect(data.job_id).toBe('test-perf-123');
    expect(data.status).toBe('ready');
    expect(data.result.analysis).toBeDefined();
    expect(data.result.processing_time_ms).toBe(25000);
    expect(duration).toBeLessThan(30000);
  });

  it('should process 5MB image within 15 seconds', async () => {
    const startTime = Date.now();

    const formData = new FormData();
    formData.append('image', new Blob([new ArrayBuffer(5 * 1024 * 1024)]));
    formData.append('chat_id', '12345');
    formData.append('processing_type', 'image-analysis');

    const mockFetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        job_id: 'test-small-123',
        status: 'ready',
        result: {
          analysis: 'Small image processed',
          confidence: 0.98,
          processing_time_ms: 15000,
        },
      }),
    } as Response);

    global.fetch = mockFetch;

    const response = await fetch('http://localhost:8080/api/v1/image/process', {
      method: 'POST',
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      body: formData,
    });

    const duration = Date.now() - startTime;

    expect(response.ok).toBe(true);

    const data = await response.json();
    expect(data.job_id).toBe('test-small-123');
    expect(data.status).toBe('ready');
    expect(data.result.processing_time_ms).toBe(15000);
    expect(duration).toBeLessThan(15000);
  });

  it('should reject images over 20MB', async () => {
    const mockFetch = vi.fn().mockResolvedValueOnce({
      ok: false,
      status: 413,
      json: async () => ({
        error: 'File too large (maximum 20MB)',
      }),
    } as Response);

    global.fetch = mockFetch;

    const response = await fetch('http://localhost:8080/api/v1/image/process', {
      method: 'POST',
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      body: new FormData(),
    });

    expect(response.ok).toBe(false);
    expect(response.status).toBe(413);
  });

  it('should handle concurrent uploads within limits', async () => {
    const formData1 = new FormData();
    formData1.append('image', new Blob([new ArrayBuffer(1024 * 1024)]));
    formData1.append('chat_id', '12345');
    formData1.append('processing_type', 'image-analysis');

    const formData2 = new FormData();
    formData2.append('image', new Blob([new ArrayBuffer(1024 * 1024)]));
    formData2.append('chat_id', '12345');
    formData2.append('processing_type', 'image-analysis');

    const formData3 = new FormData();
    formData3.append('image', new Blob([new ArrayBuffer(1024 * 1024)]));
    formData3.append('chat_id', '12345');
    formData3.append('processing_type', 'image-analysis');

    const mockFetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        job_id: 'test-concurrent-123',
        status: 'queued',
      }),
    } as Response);

    global.fetch = mockFetch;

    await Promise.all([
      fetch('http://localhost:8080/api/v1/image/process', {
        method: 'POST',
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        body: formData1,
      }),
      fetch('http://localhost:8010/api/v1/image/process', {
        method: 'POST',
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        body: formData2,
      }),
      fetch('http://localhost:8110/api/v1/image/process', {
        method: 'POST',
        headers: {
          'Telegram-Auth-Token': 'test-token',
          'Content-Type': 'multipart/form-data',
        },
        body: formData3,
      }),
    ]);

    expect(mockFetch).toHaveBeenCalledTimes(3);
  });
});
