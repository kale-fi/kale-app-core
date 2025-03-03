# KaleFi Staking and Yield Contract

This contract manages staking and yield distribution for the KaleFi platform. It is forked from the CosmWasm cw-staking contract and customized for KaleFi's specific requirements.

## Features

- Token staking with configurable lock periods
- Yield distribution based on stake amount and time
- Treasury fee collection
- Configurable reward rates

## Contract Functions

### Stake

Users can stake their tokens to earn rewards. Tokens are locked for a configurable period.

### Unstake

Users can unstake their tokens after the lock period has passed.

### Claim Rewards

Users can claim rewards based on their staked amount and the time since their last claim.

### Update Config

The contract owner can update configuration parameters such as reward rates and lock periods.

## Integration

This contract integrates with the KaleFi AMM and Social contracts to provide a complete DeFi ecosystem.
