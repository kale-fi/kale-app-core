#!/bin/bash

# Reset the node data
rm -rf ~/.kalefid

# Initialize the node
./kalefid init test --chain-id kale-test-1

# Update minimum gas prices in app.toml
sed -i '' 's/minimum-gas-prices = ".*"/minimum-gas-prices = "0.0001ukale"/' ~/.kalefid/config/app.toml

# Create test user
./kalefid keys add test-user --keyring-backend test 2>/dev/null || true
TEST_USER=$(./kalefid keys show test-user -a --keyring-backend test)
echo "Test user address: $TEST_USER"

# Add funds to the test user
./kalefid add-genesis-account $TEST_USER 10000000ukale

# Create validator
./kalefid gentx test-user 1000000ukale --chain-id kale-test-1 --keyring-backend test

# Collect genesis transactions
./kalefid collect-gentxs

# Start the node
./kalefid start
