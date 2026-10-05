#!/bin/sh
envsubst '$PORT, $ADDR' < /etc/nginx/conf.d/configfile.template > /etc/nginx/conf.d/default.conf

# Start Go backend in the background
/app/api &
GO_PID=$!

# Start Nginx in the background
nginx -g 'daemon off;' &
NGINX_PID=$!

# Trap termination signals and forward them to both processes
trap "kill -TERM $GO_PID $NGINX_PID; exit 0" SIGTERM SIGINT

# Polling loop to check both processes simultaneously
while true; do
    # Check if Go backend is still alive
    if ! kill -0 $GO_PID 2>/dev/null; then
        wait $GO_PID
        GO_EXIT_CODE=$?
        echo "Go backend died (exit code $GO_EXIT_CODE). Stopping Nginx..."
        kill -TERM $NGINX_PID 2>/dev/null
        wait $NGINX_PID
        exit $GO_EXIT_CODE
    fi

    # Check if Nginx is still alive
    if ! kill -0 $NGINX_PID 2>/dev/null; then
        wait $NGINX_PID
        NGINX_EXIT_CODE=$?
        echo "Nginx died (exit code $NGINX_EXIT_CODE). Stopping Go backend..."
        kill -TERM $GO_PID 2>/dev/null
        wait $GO_PID
        exit $NGINX_EXIT_CODE
    fi

    # Brief pause to keep CPU usage at 0%
    sleep 1
done