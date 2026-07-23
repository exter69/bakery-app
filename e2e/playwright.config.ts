import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  retries: 1,
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
  // In CI, servers are started manually in the workflow (in-memory mode with seed data).
  // Locally, Playwright starts them automatically.
  ...(!process.env.CI && {
    webServer: [
      {
        command: 'go run ./cmd/server',
        cwd: '..',
        port: 8080,
        reuseExistingServer: true,
        timeout: 30_000,
      },
      {
        command: 'npm run dev',
        cwd: '../frontend',
        port: 5173,
        reuseExistingServer: true,
        timeout: 15_000,
      },
    ],
  }),
});
