// ref: https://umijs.org/config/
const CompressionWebpackPlugin = require('compression-webpack-plugin');
const prodGzipList = ['js', 'css'];
const assetDir = "static";


export default {
	treeShaking: true,
	history: 'hash',
	outputPath: '../web',
	base: './',
	publicPath: './',
  hash: true,
  exportStatic: {
    htmlSuffix: true,
    dynamicRoot: true,
  },
	theme: {
		'@primary-color': '#666666',
	},
	define: {},
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

	runtimePublicPath: true,
	chainWebpack: config => {
		//修改JS输出目录
		config.output.filename(assetDir+'/[name].[hash:8].js').chunkFilename(assetDir+'/[name].[contenthash:8].chunk.js');
		// 修改css输出目录
		config.plugin('extract-css').tap(() => [
			{
				filename: assetDir+`/[name].[contenthash:8].css`,
				chunkFilename: assetDir+`/[name].[contenthash:8].chunk.css`,
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

	proxy: {
		'/api': {
			target: 'http://127.0.0.1:2017',
			changeOrigin: true,
			pathRewrite: {
				'^/server': '',
			},
		},
		'/metrics': {
			target: 'http://127.0.0.1:2017',
			changeOrigin: true,
			pathRewrite: {
				'^/server': '',
			},
		},
	},

	routes: [
		{
			path: '/',
			component: '../layouts/index',
			routes: [
				{ path: '/', component: '../pages/ultronHome/' },
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
