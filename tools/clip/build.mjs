import * as esbuild from 'esbuild';

// Bundle clip.mjs into a single CommonJS file the Go backend can exec.
// defuddle and linkedom are bundled in (nothing external); the DOM shim at
// the top of clip.mjs runs before the bundled defuddle code because the
// defuddle imports in clip.mjs are dynamic and stay at their call sites.
await esbuild.build({
  entryPoints: ['clip.mjs'],
  bundle: true,
  platform: 'node',
  target: 'node20',
  format: 'cjs',
  outfile: 'dist/clip.cjs',
  logLevel: 'info',
});
