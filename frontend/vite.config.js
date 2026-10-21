import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    // The dev-only mock reads the built-in quests from the Go package.
    fs: { allow: ['..', '../internal/quest/builtin'] },
  },
});
