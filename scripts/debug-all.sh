#!/bin/bash

# Configuration
SERVICES=("api" "user" "catalog" "merchant")
declare -A DLV_PORTS
DLV_PORTS["api"]=4001
DLV_PORTS["user"]=4002
DLV_PORTS["catalog"]=4003
DLV_PORTS["merchant"]=4004

# Get the absolute path of the workspace root
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Function to clean up background processes
cleanup() {
    echo -e "\n\033[1;33mStopping all services...\033[0m"
    # Kill all dlv processes and their children
    pkill -P $$ 2>/dev/null
    # Sometimes dlv processes don't die with pkill -P, so we search for them
    for port in "${DLV_PORTS[@]}"; do
        fuser -k -n tcp $port 2>/dev/null
    done
    echo "\033[1;32mStopped.\033[0m"
    exit 0
}

# Trap SIGINT and SIGTERM
trap cleanup SIGINT SIGTERM

# Check if dlv is installed
if ! command -v dlv &> /dev/null; then
    echo -e "\033[1;31mError: delve (dlv) is not installed.\033[0m"
    echo "Install it with: go install github.com/go-delve/delve/cmd/dlv@latest"
    exit 1
fi

echo -e "\033[1;34mStarting infrastructure...\033[0m"
cd "$ROOT_DIR/infrastructure" && docker compose up -d

echo -e "\033[1;34mStarting services in debug mode...\033[0m"

# Colors for log prefixing
COLORS=(31 32 33 34 35 36)

i=0
for svc in "${SERVICES[@]}"; do
    color=${COLORS[i % ${#COLORS[@]}]}
    port=${DLV_PORTS[$svc]}
    
    echo -e "\033[1;${color}m[INIT] Starting $svc on port $port...\033[0m"
    
    # Run dlv debug in background with prefixing
    (
        cd "$ROOT_DIR/service/$svc"
        # Load .env if it exists
        if [ -f .env ]; then
            export $(grep -v '^#' .env | xargs)
        fi
        
        dlv debug ./cmd/main.go --headless --listen=:$port --api-version=2 --accept-multiclient 2>&1 | \
        while read line; do 
            echo -e "\033[0;${color}m[$svc]\033[0m $line"
        done
    ) &
    
    ((i++))
done

echo -e "\033[1;32mAll services started in debug mode!\033[0m"
echo -e "\033[1;36mDelve ports:\033[0m"
for svc in "${SERVICES[@]}"; do
    echo -e "  - $svc: \033[1;37m${DLV_PORTS[$svc]}\033[0m"
done
echo -e "\033[1;33mUse Ctrl+C to stop all services.\033[0m"

# Wait for all background processes
wait
