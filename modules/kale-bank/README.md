# Kale Bank Module

The Kale Bank module is responsible for handling the native KALE token in the KaleFi ecosystem. It provides functionality for minting, transferring, and managing the KALE token supply.

## Overview

The module is a fork of the Cosmos SDK bank module, customized to support the specific requirements of the KaleFi platform. The primary features include:

- Initialization of the KALE token supply (100M tokens)
- Token minting capabilities
- Balance tracking and management
- Integration with other KaleFi modules

## Token Details

- **Name**: Kale
- **Symbol**: KALE
- **Denomination**: ukale (micro-kale, 1 KALE = 1,000,000 ukale)
- **Total Supply**: 100,000,000 KALE (100,000,000,000,000 ukale)

## Module Components

### Types

- `KaleToken`: Defines the metadata for the KALE token
- `Params`: Module parameters, including minting controls

### Keeper

- `KaleBankKeeper`: Manages the module's state and provides methods for token operations
- Key methods:
  - `MintKale`: Mints KALE tokens to a specified address
  - `InitializeKaleSupply`: Initializes the total supply of KALE tokens
  - `GetKaleBalance`: Retrieves the KALE balance of an address
  - `GetTotalKaleSupply`: Gets the total supply of KALE tokens

### Genesis

The module supports genesis initialization, allowing the initial supply to be allocated to a specific address during chain initialization.

## Usage

The Kale Bank module is integrated into the KaleFi RollApp and interacts with other modules like the Kale Rewards module for staking and reward distribution.

## Integration

To integrate this module into your application:

1. Import the module in your app.go file
2. Add the module to your app's module manager
3. Configure the module's keeper with the necessary dependencies

## License

This module is part of the KaleFi platform and is subject to the same licensing terms as the rest of the codebase.
