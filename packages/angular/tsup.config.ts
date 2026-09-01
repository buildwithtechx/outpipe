import { defineConfig } from 'tsup';

export default defineConfig({
  clean: true,
  dts: false,
  entry: ['src/index.ts'],
  external: ['@angular/common', '@angular/core', '@outpipe/sdk', 'rxjs'],
  format: ['esm'],
  sourcemap: true,
});
