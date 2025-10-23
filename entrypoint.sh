#!/bin/sh
set -e

# Generate Caddyfile if it does not exist
if [ ! -f /config/caddy/Caddyfile ]; then
    CURRENT_DIR="$(pwd)"

    # Create temporary directory
    TMP_DIR="/tmp/auto-https"
    mkdir -p "$TMP_DIR"
    cd "$TMP_DIR"

    # Generate Caddyfile
    /usr/bin/caddyfile-generate
    mkdir -p /config/caddy
    cp "$TMP_DIR/Caddyfile" /config/caddy/Caddyfile

    # Go to previous directory and clean up
    cd "$CURRENT_DIR"
    rm -rf /tmp/caddy-local
fi

# Start Caddy
exec "$@"