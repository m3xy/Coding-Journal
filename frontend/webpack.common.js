var HtmlWebpackPlugin = require("html-webpack-plugin");
const Dotenv = require("dotenv-webpack");

module.exports = {
    mode: "production",
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
    ],
    devServer: {
        historyApiFallback: true,
        host: "0.0.0.0",
        port: 23409,
    },
    externals: {
        // global app config object
        config: JSON.stringify({
            apiUrl: "http://localhost:4000",
        }),
    },
};
