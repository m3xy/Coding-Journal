const Dotenv = require('dotenv-webpack');
const { webpack } = require('webpack');
const { merge } = require('webpack-merge');
const common = require('./webpack.common.js');

module.exports = merge(common, {
  mode: 'development',
  devtool: 'inline-source-map',
  plugins: [
    new Dotenv({
      path: "./.env"
    }),
  ],
});
