#!/bin/bash
cd /Users/byronoconnor/kale-fi/kale-app-core

# Initialize if not already done
./kalefid init test-node --chain-id kalefi-1 || true

# Start node in background
./kalefid start --minimum-gas-prices=0.0001stake &
NODE_PID=$!
echo "Node started with PID: $NODE_PID"
echo "Waiting for node to start..."
sleep 10

# Verify node is running
echo "Checking if node is running..."
curl -s http://localhost:26657/status

# When done, you can kill the node with:
echo "To stop the node, run: kill $NODE_PID"
