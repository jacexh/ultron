// ref: https://umijs.org/config/
const CompressionWebpackPlugin = require('compression-webpack-plugin');
const prodGzipList = ['js', 'css'];

export default {
  treeShaking: true,
  history: 'hash',
  outputPath: '../web',
  base: './',
  publicPath: './',
  hash: true,
  theme: {
    '@primary-color': '#666666',
  },
  define: {
    //线上
    // APPID: 'dingoauyl1gaxf2oedcrdp',
    // BASE_URL: 'https://furcas.shouqianba.com/',
    // CORPID: 'dingf239b82e87a474d7',
    //测试
    // APPID: 'dingoazhtcq92uilhobijk',
    // BASE_URL: 'http://furcas.beta.iwosai.com/',
    // CORPID: 'dinge97c0113c4fe039135c2f4657eb6378f',
    //BASE_URL: 'http://localhost:8000/',
  },
  targets: {
    ie: 11,
  },
  disableRedirectHoist: true,
  ignoreMomentLocale: true,
  lessLoaderOptions: {
    javascriptEnabled: true,
  },
  manifest: {
    basePath: '/',
  },

  proxy: {
    '/api/': {
      target: 'http://127.0.0.1:8089/',
      changeOrigin: true,
      pathRewrite: {
        '^/server': '',
      },
    },
  },

  runtimePublicPath: true,
  chainWebpack: config => {
    //修改JS输出目录
    config.output.filename('[name].[hash:8].js').chunkFilename('[name].[contenthash:8].chunk.js');
    // 修改css输出目录
    config.plugin('extract-css').tap(() => [
      {
        filename: `[name].[contenthash:8].css`,
        chunkFilename: `[name].[contenthash:8].chunk.css`,
        ignoreOrder: true,
      },
    ]);

    if (process.env.NODE_ENV === 'production') {
      // 生产模式开启
      config.plugin('compression-webpack-plugin').use(
        new CompressionWebpackPlugin({
          // filename: 文件名称，这里我们不设置，让它保持和未压缩的文件同一个名称
          algorithm: 'gzip', // 指定生成gzip格式
          test: new RegExp('\\.(' + prodGzipList.join('|') + ')$'), // 匹配哪些格式文件需要压缩
          threshold: 10240, //对超过10k的数据进行压缩
          minRatio: 0.6, // 压缩比例，值为0 ~ 1
        }),
      );
    }
  },

  routes: [
    {
      path: '/',
      component: '../layouts/index',
      routes: [
        { path: '/', component: '../pages/ultronHome/' },
        { path: '/charts', component: '../pages/ultronBar/highcharttest' },
        {
          path: '/check',
          component: '../layouts/check',
        },
      ],
    },
  ],
  plugins: [
    // ref: https://umijs.org/plugin/umi-plugin-react.html
    [
      'umi-plugin-react',
      {
        dva: {
          immer: true,
        },
        antd: true,
        dynamicImport: false,
        title: 'ultron',
        dll: false,

        routes: {
          exclude: [/components\//],
        },
      },
    ],
  ],
};
