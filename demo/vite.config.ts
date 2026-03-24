import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  resolve: {
    alias: {
      '@content-extractor/core': resolve(__dirname, '../ts/src/index.ts'),
    },
  },
  optimizeDeps: {
    exclude: ['@content-extractor/core'],
    include: ['pdfjs-dist'],
  },
});
