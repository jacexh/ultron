import { defineConfig } from "umi";
const assetDir = "static";

export default defineConfig({
  nodeModulesTransform: {
    type: "none",
  },
  outputPath: "../web",
  publicPath: "./",
  history: { type: "hash" },
  routes: [{ path: "/", component: "@/pages/ultronHome/index" }],

  chainWebpack: (config) => {
    //修改JS输出目录
    config.output
      .filename(assetDir + "/[name].[hash:8].js")
      .chunkFilename(assetDir + "/[name].[contenthash:8].chunk.js");
    // 修改css输出目录
    config.plugin("extract-css").tap(() => [
      {
        filename: assetDir + `/[name].[contenthash:8].css`,
        chunkFilename: assetDir + `/[name].[contenthash:8].chunk.css`,
        ignoreOrder: true,
      },
    ]);
  },

  proxy: {
    "/api": {
      target: "http://127.0.0.1:2017",
      changeOrigin: true,
      pathRewrite: {
        "^/server": "",
      },
    },
    "/metrics": {
      target: "http://127.0.0.1:2017",
      changeOrigin: true,
      pathRewrite: {
        "^/server": "",
      },
    },
  },
  mfsu: {},
  webpack5: {},
  dynamicImport: {},
  fastRefresh: {},
});
