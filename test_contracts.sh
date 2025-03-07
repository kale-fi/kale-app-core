#!/bin/bash
cd /Users/byronoconnor/kale-fi/kale-app-core

# Initialize node if not already done
if [ ! -d "$HOME/.kalefid/config" ]; then
  echo "Initializing node..."
  ./kalefid init test --chain-id kale-test-1
fi

# Update minimum gas prices in app.toml
echo "Updating app.toml..."
sed -i '' 's/minimum-gas-prices = ".*"/minimum-gas-prices = "0.0001ukale"/' $HOME/.kalefid/config/app.toml

# Create test user if not exists
echo "Setting up test user..."
./kalefid keys add test-user --keyring-backend test 2>/dev/null || true
TEST_USER=$(./kalefid keys show test-user -a --keyring-backend test)
echo "Test user address: $TEST_USER"

# Add funds to the test user
echo "Adding funds to test user..."
echo '{"address": "'"$TEST_USER"'", "coins": [{"denom": "ukale", "amount": "10000000"}]}' > $HOME/.kalefid/config/genesis_account.json
./kalefid add-genesis-account $(cat $HOME/.kalefid/config/genesis_account.json)

# Create validator
echo "Creating validator..."
./kalefid gentx test-user 1000000ukale --chain-id kale-test-1 --keyring-backend test

# Collect genesis transactions
echo "Collecting genesis transactions..."
./kalefid collect-gentxs

# Start node in background
echo "Starting node..."
./kalefid start --minimum-gas-prices="0.0001ukale" > node.log 2>&1 &
NODE_PID=$!
echo "Node started with PID: $NODE_PID"

# Wait for node to start
echo "Waiting for node to start..."
sleep 10

# Check if node is running
if ps -p $NODE_PID > /dev/null; then
  echo "Node is running."
else
  echo "Node failed to start. Check node.log for details."
  exit 1
fi

# Store contracts
echo "Storing kale-amm..."
AMM_CODE_TX=$(./kalefid tx wasm store contracts/kale-amm/target/wasm32-unknown-unknown/release/kale_amm.wasm --from test-user --chain-id kale-test-1 --node tcp://localhost:26657 --gas auto --fees 5000ukale --keyring-backend test -y --output json | jq -r '.txhash')
echo "Storing kale-social..."
SOCIAL_CODE_TX=$(./kalefid tx wasm store contracts/kale-social/target/wasm32-unknown-unknown/release/kale_social.wasm --from test-user --chain-id kale-test-1 --node tcp://localhost:26657 --gas auto --fees 5000ukale --keyring-backend test -y --output json | jq -r '.txhash')
echo "Storing kale-rewards..."
REWARDS_CODE_TX=$(./kalefid tx wasm store contracts/kale-rewards/target/wasm32-unknown-unknown/release/kale_rewards.wasm --from test-user --chain-id kale-test-1 --node tcp://localhost:26657 --gas auto --fees 5000ukale --keyring-backend test -y --output json | jq -r '.txhash')

# Get code IDs
sleep 5  # Wait for tx confirmation
AMM_CODE_ID=$(./kalefid query tx $AMM_CODE_TX --node tcp://localhost:26657 --output json | jq -r '.raw_log' | grep -o '"code_id":"[0-9]*"' | cut -d'"' -f4)
SOCIAL_CODE_ID=$(./kalefid query tx $SOCIAL_CODE_TX --node tcp://localhost:26657 --output json | jq -r '.raw_log' | grep -o '"code_id":"[0-9]*"' | cut -d'"' -f4)
REWARDS_CODE_ID=$(./kalefid query tx $REWARDS_CODE_TX --node tcp://localhost:26657 --output json | jq -r '.raw_log' | grep -o '"code_id":"[0-9]*"' | cut -d'"' -f4)
echo "Code IDs: AMM=$AMM_CODE_ID, Social=$SOCIAL_CODE_ID, Rewards=$REWARDS_CODE_ID"

# Instantiate
echo "Instantiating kale-amm..."
AMM_INST_TX=$(./kalefid tx wasm instantiate $AMM_CODE_ID '{"owner":"'"$TEST_USER"'","fee_percent":2,"yield_percent":50,"lp_percent":30,"treasury_percent":20,"token_a":"ukale","token_b":"uusdc","reserves_a":"0","reserves_b":"0"}' --from test-user --chain-id kale-test-1 --node tcp://localhost:26657 --label "kale-amm" --gas auto --fees 5000ukale --keyring-backend test -y --output json | jq -r '.txhash')
echo "Instantiating kale-social..."
SOCIAL_INST_TX=$(./kalefid tx wasm instantiate $SOCIAL_CODE_ID '{"owner":"'"$TEST_USER"'"}' --from test-user --chain-id kale-test-1 --node tcp://localhost:26657 --label "kale-social" --gas auto --fees 5000ukale --keyring-backend test -y --output json | jq -r '.txhash')
echo "Instantiating kale-rewards..."
REWARDS_INST_TX=$(./kalefid tx wasm instantiate $REWARDS_CODE_ID '{"owner":"'"$TEST_USER"'","fee_percent":2,"yield_percent":50,"lp_percent":30,"treasury_percent":20,"token_a":"ukale","token_b":"uusdc","reserves_a":"0","reserves_b":"0"}' --from test-user --chain-id kale-test-1 --node tcp://localhost:26657 --label "kale-rewards" --gas auto --fees 5000ukale --keyring-backend test -y --output json | jq -r '.txhash')

# Get contract addresses
sleep 5
AMM_ADDR=$(./kalefid query tx $AMM_INST_TX --node tcp://localhost:26657 --output json | jq -r '.raw_log' | grep -o '"contract_address":"kale1[^"]*"' | cut -d'"' -f4)
SOCIAL_ADDR=$(./kalefid query tx $SOCIAL_INST_TX --node tcp://localhost:26657 --output json | jq -r '.raw_log' | grep -o '"contract_address":"kale1[^"]*"' | cut -d'"' -f4)
REWARDS_ADDR=$(./kalefid query tx $REWARDS_INST_TX --node tcp://localhost:26657 --output json | jq -r '.raw_log' | grep -o '"contract_address":"kale1[^"]*"' | cut -d'"' -f4)
echo "Contract addresses: AMM=$AMM_ADDR, Social=$SOCIAL_ADDR, Rewards=$REWARDS_ADDR"

# Fund contracts
echo "Funding contracts..."
./kalefid tx bank send test-user $AMM_ADDR 1000ukale,1000uusdc --chain-id kale-test-1 --node tcp://localhost:26657 --gas auto --fees 5000ukale --keyring-backend test -y
./kalefid tx bank send test-user $REWARDS_ADDR 1000ukale,1000uusdc --chain-id kale-test-1 --node tcp://localhost:26657 --gas auto --fees 5000ukale --keyring-backend test -y

# Test interactions
echo "Testing kale-amm swap..."
./kalefid tx wasm execute $AMM_ADDR '{"kale_swap":{"amount":"100","token_in":"ukale","token_out":"uusdc"}}' --from test-user --amount 100ukale --chain-id kale-test-1 --node tcp://localhost:26657 --gas auto --fees 5000ukale --keyring-backend test -y
echo "Testing kale-social follow..."
./kalefid tx wasm execute $SOCIAL_ADDR '{"kale_follow":{"trader":"kale1trader","stake_amount":"100"}}' --from test-user --chain-id kale-test-1 --node tcp://localhost:26657 --gas auto --fees 5000ukale --keyring-backend test -y

echo "Test completed successfully!"
