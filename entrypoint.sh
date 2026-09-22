#!/bin/sh

# Start the Go backend in the background and pipe its output directly to stdout
/app/api &
GO_PID=$!

# Wait 2 seconds to give Go time to initialize or crash
sleep 2

# Check if Go process died immediately
if ! kill -0 $GO_PID 2>/dev/null; then
    echo "ERROR: Go backend failed to start or crashed immediately."
    exit 1
fi

echo "Go backend is running successfully."

# Inject Cloud Run's dynamic PORT into Nginx and start Nginx in the foreground
envsubst '$PORT' < /etc/nginx/conf.d/configfile.template > /etc/nginx/conf.d/default.conf
nginx -g 'daemon off;'