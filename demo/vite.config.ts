import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  resolve: {
    alias: {
      '@damorris25/content-extractor': resolve(__dirname, '../ts/src/index.ts'),
    },
  },
  optimizeDeps: {
    exclude: ['@damorris25/content-extractor'],
    include: ['pdfjs-dist'],
  },
});
