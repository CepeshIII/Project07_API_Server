#!/bin/sh
envsubst '$PORT, $ADDR' < /etc/nginx/conf.d/configfile.template > /etc/nginx/conf.d/default.conf


# Start Go backend
/app/api &
GO_PID=$!

# Start Nginx in background
nginx -g 'daemon off;' &
NGINX_PID=$!

# Trap signals and forward them
trap "kill -TERM $GO_PID $NGINX_PID" SIGTERM SIGINT

# Wait for EITHER process to exit
wait -n

# Exit container if either process dies
exit 1