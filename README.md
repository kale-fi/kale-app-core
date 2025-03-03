# KaleFi RollApp

KaleFi RollApp blockchain hosted on AWS EC2

## Overview

This repository implements the Dymension RollApp blockchain with CosmWasm smart contracts for trading, social features, and tokenomics. The platform combines AMM functionality with social trading features, allowing users to follow successful traders and copy their trades.

## Features

- Token swaps (AMM) with 0.2% fee
- Trader profiles and copy-trading
- $SOCIAL token rewards and $USDC yield distribution
- Chain ID: kalefi-1

## Development

```bash
# Build the application
make -f KaleMakefile build

# Run tests
make -f KaleMakefile test
```

## Directory Structure

- `/contracts` - CosmWasm smart contracts
- `/modules` - Cosmos SDK modules
- `/app` - Application wiring
- `/cmd` - Command-line interfaces
- `/proto` - Protocol buffer definitions