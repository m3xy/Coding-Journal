var HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
    mode: 'development',
    resolve: {
        extensions: ['.js', '.jsx']
    },
    module: {
        rules: [
            {
                test: /\.jsx?$/,
                loader: 'babel-loader'
            },
            {
                test: /\.css/,
                loader: 'style-loader!css-loader'
            },
            {
                test: /\.log$/i,
                use: 'raw-loader',
            },
        ]
    },
    plugins: [new HtmlWebpackPlugin({
        template: './src/index.html'
    })],
    devServer: {
        historyApiFallback: true,
        port: 23409,
        proxy: {
            '/api/*': {
                target: 'http://localhost:3333',
                secure: false,
                changeOrigin: true,
                headers: {
                  'Access-Control-Allow-Origin': '\*',
                  'Access-Control-Allow-Headers': 'X-Requested-With, X-FOREIGNJOURNAL-SECURITY-TOKEN',
                  'Access-Control-Allow-Methods': 'GET, HEAD, POST, PUT, OPTIONS'
                }
            }
        },
        headers: {
            'Access-Control-Allow-Origin': '\*',
            'Access-Control-Allow-Headers': 'X-Requested-With, X-FOREIGNJOURNAL-SECURITY-TOKEN',
            'Access-Control-Allow-Methods': 'GET, HEAD, POST, PUT, OPTIONS'
        }
    },
    externals: {
        // global app config object
        config: JSON.stringify({
            apiUrl: 'http://localhost:4000'
        })
    },
}
