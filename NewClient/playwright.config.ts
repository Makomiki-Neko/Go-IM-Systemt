import { defineConfig } from '@playwright/test'
export default defineConfig({
  testDir: './tests/browser', fullyParallel: false, workers: 1, timeout: 30000,
  use: { channel: process.env.BROWSER_CHANNEL || 'chrome', baseURL: 'http://127.0.0.1:3001', headless: true, viewport: { width: 1440, height: 960 }, screenshot: 'only-on-failure', trace: 'retain-on-failure' },
  webServer: { command: 'npm run dev', url: 'http://127.0.0.1:3001', reuseExistingServer: !process.env.CI, timeout: 30000 },
})
