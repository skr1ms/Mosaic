const webpack = require('webpack');
const path = require('path');

module.exports = function override(config, env) {
    
    config.resolve.fallback = {
        ...config.resolve.fallback,
                "url": require.resolve("url"),
        "assert": require.resolve("assert"),
        "crypto": require.resolve("crypto-browserify"),
        "http": require.resolve("stream-http"),
        "https": require.resolve("https-browserify"),
        "os": require.resolve("os-browserify/browser"),
        "buffer": require.resolve("buffer"),
        "stream": require.resolve("stream-browserify"),
        "path": require.resolve("path-browserify"),
        "fs": false,
        "net": false,
        "tls": false,
        "child_process": false,
        "async_hooks": false,
        "util": require.resolve("util"),
        "vm": require.resolve("vm-browserify"),
        "events": require.resolve("events"),
        "string_decoder": require.resolve("string_decoder"),
        "constants": require.resolve("constants-browserify"),
        "process": require.resolve("process/browser.js"),
    };

    config.plugins = [
        ...config.plugins,
        new webpack.ProvidePlugin({
            process: 'process/browser.js',
            Buffer: ['buffer', 'Buffer'],
        }),
    ].filter(Boolean);

    
    if (Array.isArray(config.plugins)) {
        config.plugins.forEach(plugin => {
            const isESLintPlugin = plugin && plugin.constructor && plugin.constructor.name === 'ESLintWebpackPlugin';
            if (isESLintPlugin) {
                plugin.options = {
                    ...(plugin.options || {}),
                    cache: true,
                    cacheLocation: path.resolve(__dirname, '.eslintcache'),
                };
            }
        });
    }

    
    config.ignoreWarnings = [
        ...(config.ignoreWarnings || []),
        
        /Failed to parse source map.*react-bootstrap-sweetalert/,
        
        /Failed to parse source map/,
        /source-map-loader/,
        
        /ENOENT.*index\.js/,
        /ENOENT.*\.tsx?/,
        
        /content-all\.js/,
        /chrome-extension/,
        /moz-extension/,
        /Could not establish connection/,
        /message channel closed/,
        
        /Module Warning.*source-map-loader/,
        
        /postcss-resolve-url.*deprecated/,
        /postcss\.plugin was deprecated/,
        
        /Deprecation.*Sass @import rules are deprecated/,
        /Deprecation.*Global built-in functions are deprecated/,
        /Deprecation.*color\.mix instead/,
        /Deprecation.*color\.channel.*deprecated/,
        /Deprecation.*math\.div.*deprecated/,
        /Deprecation.*Using \/ for division.*deprecated/,
        /Deprecation.*The legacy JS API is deprecated/,
        /More info.*sass-lang\.com/,
    ];

    
    if (config.module && config.module.rules) {
        config.module.rules.forEach(rule => {
            if (rule.oneOf) {
                rule.oneOf.forEach(oneOfRule => {
                    if (oneOfRule.use && oneOfRule.use.some(use => 
                        use.loader && use.loader.includes('source-map-loader'))) {
                        oneOfRule.exclude = [
                            ...(oneOfRule.exclude || []),
                            /node_modules\/react-bootstrap-sweetalert/,
                            /node_modules\/react-redux/,
                            /node_modules\/react-toastify/,
                            /node_modules\/react-table/,
                            /content-all\.js/,
                            /chrome-extension/,
                            /moz-extension/
                        ];
                    }

                    
                    if (oneOfRule.use && Array.isArray(oneOfRule.use)) {
                        oneOfRule.use.forEach(useItem => {
                            if (useItem.loader && useItem.loader.includes('postcss-loader')) {
                                useItem.options = {
                                    ...useItem.options,
                                    postcssOptions: {
                                        ...useItem.options?.postcssOptions,
                                        plugins: [
                                            
                                            ...(useItem.options?.postcssOptions?.plugins || []),
                                        ],
                                        
                                        hideNothingWarning: true,
                                    }
                                };
                            }
                            
                            if (useItem.loader && useItem.loader.includes('sass-loader')) {
                                useItem.options = {
                                    ...useItem.options,
                                    sassOptions: {
                                        ...useItem.options?.sassOptions,
                                        
                                        quietDeps: true,
                                        verbose: false,
                                    },
                                    
                                    additionalData: '',
                                };
                            }
                        });
                    }
                });
            }
        });
    }

    // Suppress deprecation warnings at console level
    const originalConsoleWarn = console.warn;
    console.warn = function(...args) {
        const message = args.join(' ');
        // Suppress PostCSS deprecation warnings
        if (message.includes('postcss.plugin was deprecated') || 
            message.includes('postcss-resolve-url')) {
            return;
        }
        
        if (message.includes('Deprecation') && (
            message.includes('Sass @import rules are deprecated') ||
            message.includes('Global built-in functions are deprecated') ||
            message.includes('color.mix instead') ||
            message.includes('color.channel') ||
            message.includes('math.div') ||
            message.includes('Using / for division') ||
            message.includes('The legacy JS API is deprecated') ||
            message.includes('sass-lang.com/d/')
        )) {
            return;
        }
        
        if (message.includes('repetitive deprecation warnings omitted')) {
            return;
        }
        originalConsoleWarn.apply(console, args);
    };

    
    if (env === 'development') {
        if (config.devServer) {
            config.devServer.setupMiddlewares = (middlewares, devServer) => {
                
                return middlewares;
            };
        }

        
        config.output = {
            ...config.output,
            chunkFilename: '[name].[chunkhash].chunk.js',
            chunkLoadTimeout: 120000, 
        };
        
        
        config.optimization = {
            ...config.optimization,
            splitChunks: {
                ...config.optimization.splitChunks,
                chunks: 'all',
                cacheGroups: {
                    default: {
                        minChunks: 1,
                        priority: -20,
                        reuseExistingChunk: true,
                    },
                    vendor: {
                        test: /[\\/]node_modules[\\/]/,
                        name: 'vendors',
                        priority: -10,
                        chunks: 'all',
                    },
                },
            },
        };
    }

    return config;
}