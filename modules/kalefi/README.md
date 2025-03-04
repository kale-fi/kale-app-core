# KaleFi Module

The KaleFi module is responsible for tracking trade events within the KaleFi platform. It provides functionality for storing, retrieving, and querying trade events.

## Overview

The module is designed to track trading activity in the KaleFi ecosystem. The primary features include:

- Recording trade events with trader information and amounts
- Retrieving trade events by ID or trader
- Configurable parameters for maximum trade amounts and enabling/disabling trading

## Module Components

### Types

- `KaleTradeEvent`: Represents a trading event with trader information and amount
- `Params`: Module parameters, including maximum trade amount and trading status

### Keeper

- `KalefiKeeper`: Manages the module's state and provides methods for trade event operations
- Key methods:
  - `StoreTradeEvent`: Records a new trade event in the module's state
  - `GetTradeEvent`: Retrieves a trade event by its ID
  - `GetAllTradeEvents`: Returns all recorded trade events
  - `GetTradeEventsByTrader`: Returns all trade events for a specific trader

### Genesis

The module supports genesis initialization with configurable parameters.

## Usage

The KaleFi module is integrated into the KaleFi RollApp and interacts with other modules like the AMM module for tracking trading activity.

## Integration

To integrate this module into your application:

1. Import the module in your app.go file
2. Add the module to your app's module manager
3. Configure the module's keeper with the necessary dependencies

## License

This module is part of the KaleFi platform and is subject to the same licensing terms as the rest of the codebase.
