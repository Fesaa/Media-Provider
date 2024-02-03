const esbuild = require("esbuild");

(async () => {
  let ctx = await esbuild.context({
    entryPoints: ["src/*.tsx", "src/components/*.tsx"],
    outdir: "public/generated",
    bundle: true,
  });
  await ctx.watch();
  console.log("waiting");
})();
