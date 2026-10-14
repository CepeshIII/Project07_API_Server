#!/bin/sh
envsubst '$PORT, $ADDR' < /etc/nginx/conf.d/configfile.template > /etc/nginx/conf.d/default.conf

# Start Go backend in the background
/app/api &
GO_PID=$!

# Start Nginx in the background
nginx -g 'daemon off;' &
NGINX_PID=$!

# # Trap termination signals and forward them to both processes
# trap "kill -TERM $GO_PID $NGINX_PID; exit 0" SIGTERM SIGINT

# Monitor background processes using POSIX-compatible kill -0
while true; do
    if ! kill -0 "$NGINX_PID" 2>/dev/null; then
        echo "Nginx died."
        break
    fi
    
    if ! kill -0 "$GO_PID" 2>/dev/null; then
        echo "Go backend died."
        break
    fi
    
    sleep 2
done

# Cleanup any remaining processes just in case
kill "$NGINX_PID" 2>/dev/null
kill "$GO_PID" 2>/dev/null

# Check if we are in production or local
if [ "$TARGET" = "production" ] || [ -n "$K_SERVICE" ]; then
    echo "Production environment detected. Sleeping forever to prevent crash loops..."
    while true; do
        sleep 60
    done
else
    echo "Local environment detected. Exiting container to surface the error."
    exit 1
fi