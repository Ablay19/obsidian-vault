import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'miniflare',
    compatibilityDate: '2024-01-01',
  },
});
