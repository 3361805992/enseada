/**
 * Copyright 2019-2020 Enseada authors
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const purgecss = require('@fullhuman/postcss-purgecss')({
  content: [
    './app/js/**/*.js',
    './app/scss/**/*.scss',
    './templates/**/*.html'
  ],
  defaultExtractor: content => content.match(/[\w-/:]+(?<!:)/g) || []
})

const path = require('path')

module.exports = (env, { mode }) => {
  return {
    entry: {
      app: './app/js/app.js',
    },
    output: {
      filename: '[name].js',
      path: path.resolve(__dirname, 'static')
    },
    optimization: {
      runtimeChunk: 'single',
    },
    module: {
      rules: [
        {
          test: /\.s[ac]ss$/i,
          use: [
            MiniCssExtractPlugin.loader,
            'css-loader',
            {
              loader: 'sass-loader',
              options: {
                // Prefer `dart-sass`
                implementation: require('sass')
              }
            },
            {
              loader: 'postcss-loader',
              options: {
                ident: 'postcss',
                plugins: [
                  require('tailwindcss')
                ].concat(mode === 'production'
                  ? [purgecss]
                  : [])
              }
            }
          ]
        }
      ]
    },
    plugins: [
      new MiniCssExtractPlugin({
        filename: '[name].css',
        chunkFilename: '[name].css'
      })
    ]
  }
}