const esbuild = require("esbuild");

(async () => {
  await esbuild.build({
    entryPoints: ["src/*.tsx", "src/components/*.tsx"],
    outdir: "public/generated",
    bundle: true,
  });
})();
