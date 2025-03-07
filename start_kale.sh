#!/bin/bash

# Reset the node data
echo "Resetting node data..."
rm -rf ~/.kalefid

# Initialize the node
echo "Initializing node..."
./kalefid init test --chain-id kale-test-1

# Update the app.toml file to set minimum gas prices
echo "Updating app.toml with minimum gas prices..."
APP_TOML="$HOME/.kalefid/config/app.toml"

if [ -f "$APP_TOML" ]; then
    # Check if the file exists and update the minimum gas prices
    sed -i '' 's/minimum-gas-prices = ".*"/minimum-gas-prices = "0.0001ukale"/' "$APP_TOML"
    echo "Updated minimum-gas-prices in $APP_TOML"
else
    echo "Error: $APP_TOML not found"
    exit 1
fi

# Create a genesis account
echo "Creating test account..."
./kalefid keys add test-user --keyring-backend test

# Get the address
TEST_USER=$(./kalefid keys show test-user -a --keyring-backend test)
echo "Test user address: $TEST_USER"

# Start the node
echo "Starting node..."
./kalefid start
