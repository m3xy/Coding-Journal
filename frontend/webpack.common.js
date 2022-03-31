var HtmlWebpackPlugin = require("html-webpack-plugin")
const Dotenv = require("dotenv-webpack")
const { DefinePlugin } = require("webpack")
const MonacoWebpackPlugin = require("monaco-editor-webpack-plugin")

module.exports = {
	entry: ["regenerator-runtime/runtime.js", "./src/index.jsx"],
	resolve: {
		extensions: [".js", ".jsx"]
	},
	module: {
		rules: [
			{
				test: /\.jsx?$/,
				use: "babel-loader"
			},
			{
				test: /\.css/,
				use: ["style-loader", "css-loader"]
			},
			{
				test: /\.log$/i,
				use: "raw-loader"
			}
		]
	},
	optimization: {
		//makes sure you only have a single runtime (with module cache) and modules are not instantiated twice.
		runtimeChunk: "single",
		splitChunks: {
			chunks: "all"
		}
	},
	plugins: [
		new HtmlWebpackPlugin({
			template: "./src/index.html"
		}),
		new MonacoWebpackPlugin({
			// available options are documented at https://github.com/Microsoft/monaco-editor-webpack-plugin#options
		}),
		new DefinePlugin({
			MONACO_SUPPORTED_LANGUAGES: [
				"c",
				"'clojure'",
				"'cpp'",
				"csharp",
				"css",
				"dart",
				"dockerfile",
				"elixir",
				"fsharp",
				"go",
				"graphql",
				"html",
				"ini",
				"java",
				"javascript",
				"json",
				"julia",
				"kotlin",
				"lua",
				"markdown",
				"mips",
				"mysql",
				"objective-c",
				"pascal",
				"perl",
				"pgsql",
				"php",
				"plaintext",
				"powerquery",
				"powershell",
				"pug",
				"python",
				"qsharp",
				"r",
				"razor",
				"redis",
				"ruby",
				"rust",
				"scala",
				"scheme",
				"scss",
				"shell",
				"sql",
				"swift",
				"typescript",
				"vb",
				"xml",
				"yaml"
			].map((lang) => JSON.stringify(lang))
		})
	],
	devServer: {
		historyApiFallback: true,
		host: "0.0.0.0",
		port: 23409,
		hot: "only"
	},
	externals: {
		// global app config object
		config: JSON.stringify({
			apiUrl: "http://localhost:4000"
		})
	},
	watchOptions: {
		aggregateTimeout: 500, // delay before reloading
		poll: 1000 // enable polling since fsevents are not supported in docker
	},
	output: {
		publicPath: "/"
	}
}
