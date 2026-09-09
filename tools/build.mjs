// SPDX-License-Identifier: BSD-3-Clause
import { execFileSync } from "node:child_process";
import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { build } from "esbuild";
import { rollup } from "rollup";
import { dts } from "rollup-plugin-dts";

const root = fileURLToPath(new URL("../", import.meta.url));
const file = (name) => new URL(`../${name}`, import.meta.url);
const read = (name) => readFile(file(name), "utf8");
const { version } = JSON.parse(await read("package.json"));
const banner = `/*! BamtiGraph ${version} | BSD-3-Clause | See LICENSE */`;
await mkdir(file("dist"), { recursive: true });
try {
  execFileSync(
    process.execPath,
    ["node_modules/typescript/bin/tsc", "-p", "tsconfig.build.json"],
    { cwd: root, stdio: "inherit" },
  );
  const declarations = await rollup({
    input: fileURLToPath(file(".build/types/index.d.ts")),
    plugins: [dts()],
  });
  try {
    await declarations.write({
      file: fileURLToPath(file("dist/index.d.ts")),
      format: "es",
    });
  } finally {
    await declarations.close();
  }
  await writeFile(file("dist/index.d.cts"), await read("dist/index.d.ts"));
  const common = {
    absWorkingDir: root,
    bundle: true,
    target: "es2022",
    platform: "neutral",
    legalComments: "inline",
    banner: { js: banner },
  };
  await Promise.all([
    build({
      ...common,
      entryPoints: ["src/index.ts"],
      format: "esm",
      outfile: "dist/index.js",
      sourcemap: true,
    }),
    build({
      ...common,
      entryPoints: ["src/index.ts"],
      format: "cjs",
      outfile: "dist/index.cjs",
      sourcemap: true,
    }),
    build({
      ...common,
      stdin: {
        contents:
          'import BamtiGraph from "./src/index.ts"; globalThis.BamtiGraph = BamtiGraph;',
        resolveDir: root,
        sourcefile: "browser-entry.js",
      },
      format: "iife",
      outfile: "dist/bamtigraph.global.js",
    }),
  ]);
  let html = await read("index.html");
  html = html.replace(
    '<link rel="stylesheet" href="assets/ui.css">',
    `<style>\n${await read("assets/ui.css")}\n</style>`,
  );
  for (const match of [...html.matchAll(/<script src="([^"]+)"><\/script>/g)]) {
    const script = (await read(match[1])).replace(/<\/script/gi, "<\\/script");
    html = html.replace(match[0], `<script>\n${script}\n</script>`);
  }
  html = html.replace(
    /href="(?:docs\/index.html|examples\/[^\"]+\.html)"/g,
    'href="https://github.com/gosuda/BamtiGraph"',
  );
  await writeFile(file("standalone.html"), html);
  console.log(
    "Built ESM, CommonJS, bundled declarations, browser global and standalone.html.",
  );
} finally {
  await rm(file(".build"), { recursive: true, force: true });
}
