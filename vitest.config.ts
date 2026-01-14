import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'miniflare',
    format: ['verbose'],
    outputFile: {
      verbose: 'test-results/verbose.txt',
    },
  },
});