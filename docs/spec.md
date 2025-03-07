# RollApp-Core Specification for Windsurf (Social Trading DEX/SocialFi Platform)

## Overview
This repository implements the Dymension RollApp blockchain with CosmWasm smart contracts for trading, social features, and tokenomics. Windsurf’s Cascade feature will handle multi-file integration and Rust complexity.

## Use Case
- Execute token swaps (AMM), manage trader profiles/copy-trading, and distribute `$SOCIAL` rewards and `$USDC` yield.

## Directory Structure & File Details

### /contracts/
#### /amm/
- **src/lib.rs**: Entry point for AMM contract.
- **src/contract.rs**: Swap and liquidity logic.
  - Use: Executes swaps with 0.2% fee (50% yield, 25% LP, 25% Treasury).
  - Pseudocode:
    ```
    execute_swap(sender, amount_in, token_in, token_out) {
      reserves = get_pool_reserves(token_in, token_out);
      amount_out = calculate_xyk(reserves, amount_in);
      fee = amount_in * 0.002;
      split_fee(fee, 0.5, 0.3, 0.2); // yield, LP, Treasury
      update_reserves(reserves, amount_in, amount_out);
      emit_event("Swap", sender, amount_out);
    }
    ```
- **src/state.rs**: Pool state (reserves, LP tokens).
- **src/msg.rs**: Messages (e.g., `Swap { amount, token_in, token_out }`).
- **tests/integration.rs**: Tests swap logic.

#### /social/
- **src/contract.rs**: Trader profiles and copy-trading.
  - Use: Registers traders, executes mirrored trades (10% profit to trader, 2% Treasury).
  - Pseudocode:
    ```
    execute_follow(sender, trader, stake_amount) {
      profile = get_trader_profile(trader);
      profile.followers.push(sender);
      stake(sender, stake_amount);
      emit_event("Follow", sender, trader);
    }
    execute_copy_trade(trader, trade) {
      profit = execute_trade(trade);
      trader_fee = profit * 0.1;
      treasury_fee = profit * 0.02;
      distribute(trader_fee, treasury_fee);
    }
    ```

#### /rewards/
- **src/contract.rs**: Staking and yield distribution.
  - Use: Stakes `$SOCIAL`/`$USDC`, distributes 8–12% `$USDC` APY.
  - Pseudocode: (See Tokenomics spec below)

### /modules/
#### /bank/
- **keeper.rs**: Manages `$SOCIAL` minting (100M supply).
- **types.rs**: Token metadata.

#### /socialdex/
- **keeper.rs**: App-specific state (e.g., trade events).
- **types.rs**: Structs (e.g., `TradeEvent`).

### /app/
- **app.go**: Wires modules and contracts.
- **genesis.go**: Initial state (100M `$SOCIAL`).

### /cmd/
- **socialdexd/main.go**: RollApp daemon.
- **socialdexcli/main.go**: CLI.

### /proto/
- **socialdex.proto**: Protobuf for gRPC.

## Prompt Guidance for Windsurf
- "Fork `dymensionxyz/rollapp-template` and add an AMM contract in `contracts/amm/` with 0.2% fee split per this spec."
- "Implement `social/contract.rs` for trader profiles and copy-trading with 10% trader fee and 2% Treasury."
- "Extend `modules/bank/keeper.rs` to mint 100M `$SOCIAL` tokens."

## Notes
- Windsurf’s Cascade can wire `app.go` with contracts; ensure Rust compiles (`cargo build`).
- Fork `cw-plus` and `osmosis` for contract bases.