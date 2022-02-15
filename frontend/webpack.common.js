var HtmlWebpackPlugin = require("html-webpack-plugin");
const Dotenv = require("dotenv-webpack");
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin');

module.exports = {
    entry: ["regenerator-runtime/runtime.js", './src/index.jsx'],
    resolve: {
        extensions: [".js", ".jsx"],
    },
    module: {
        rules: [
            {
                test: /\.jsx?$/,
                use: "babel-loader",
            },
            {
                test: /\.css/,
                use: ["style-loader", "css-loader"],
            },
            {
                test: /\.log$/i,
                use: "raw-loader",
            },
        ],
    },
    optimization: {
        runtimeChunk: true,
        splitChunks: {
            chunks: "all",
        },
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: "./src/index.html",
        }),
        new MonacoWebpackPlugin({
            // available options are documented at https://github.com/Microsoft/monaco-editor-webpack-plugin#options
        }),
    ],
    devServer: {
        historyApiFallback: true,
        host: "0.0.0.0",
        port: 23409,
        hot: 'only',
    },
    externals: {
        // global app config object
        config: JSON.stringify({
            apiUrl: "http://localhost:4000",
        }),
    },
    watchOptions: {
        aggregateTimeout: 500, // delay before reloading
        poll: 1000 // enable polling since fsevents are not supported in docker
    }
};
