
const env = {
    REDIS_SERVER: "119.82.141.213:6379",
    REDIS_PASSWORD:"MkZ5bc9zUC9ZTsiKR3gC",
    REDIS_DB: "1",
    PORT: "8888",

    SERVER_ENDPOINT: "103.162.31.207",
    SERVER_PORT: "443"
};


module.exports = {
    apps: [{
        name: "control-server",
        script: "./scripts/start.sh",
        log_date_format: 'YYYY-MM-DD HH:mm:ss.SSS',
        env: {
            ...env
        },
        error_file: './logs/control-server.err',
        out_file: './logs/control-server.log'
    }]
}