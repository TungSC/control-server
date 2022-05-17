
const env = {
    REDIS_SERVER: "119.82.141.213:6379",
    REDIS_PASSWORD:"MkZ5bc9zUC9ZTsiKR3gC",
    REDIS_DB: "1",
    PORT: "2577",

    SERVER_ENDPOINT: "119.82.141.213",
    SERVER_PORT: "1953"
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