use cosmwasm_std::{
    to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128, BankMsg, Coin, CosmosMsg, StdError,
};

use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg, TradeInfo, TraderResponse};
use crate::state::{Config, CONFIG, TRADER_PROFILES, TraderProfile, Follower};

pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> StdResult<Response> {
    let config = Config {
        owner: deps.api.addr_validate(&msg.owner)?,
        trader_fee_percent: msg.trader_fee_percent,
        treasury_fee_percent: msg.treasury_fee_percent,
        treasury_address: deps.api.addr_validate(&msg.treasury_address)?,
    };
    CONFIG.save(deps.storage, &config)?;

    Ok(Response::new().add_attribute("method", "instantiate"))
}

pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> StdResult<Response> {
    match msg {
        ExecuteMsg::Follow { trader, stake_amount } => 
            execute_follow(deps, env, info, trader, stake_amount),
        ExecuteMsg::CopyTrade { trader, trade } => 
            execute_copy_trade(deps, env, info, trader, trade),
        ExecuteMsg::KaleFollow { trader, stake_amount } => 
            execute_follow(deps, env, info, trader, stake_amount),
        ExecuteMsg::KaleCopyTrade { trader, trade } => {
            // Convert TradeMsg to TradeInfo for compatibility
            let trade_info = TradeInfo {
                trade_type: "swap".to_string(),
                token_pair: format!("{}/{}", trade.token_in, trade.token_out),
                amount: trade.amount,
                expected_profit: Uint128::zero(), // Will be calculated in execute_copy_trade
            };
            execute_copy_trade(deps, env, info, trader, trade_info)
        },
    }
}

pub fn execute_follow(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    trader: String,
    stake_amount: Uint128,
) -> StdResult<Response> {
    // Validate trader address
    let trader_addr = deps.api.addr_validate(&trader)?;
    
    // Get or create trader profile
    let mut profile = TRADER_PROFILES
        .may_load(deps.storage, &trader_addr)?
        .unwrap_or_else(|| TraderProfile {
            address: trader_addr.clone(),
            followers: vec![],
            total_stake: Uint128::zero(),
            performance: Uint128::zero(),
        });
    
    // Check if follower is already following this trader
    if !profile.followers.iter().any(|f| f.address == info.sender) {
        // Add follower to trader's profile
        profile.followers.push(Follower {
            address: info.sender.clone(),
            stake_amount,
        });
        
        // Update total stake
        profile.total_stake = profile.total_stake.checked_add(stake_amount)?;
        
        // Save updated profile
        TRADER_PROFILES.save(deps.storage, &trader_addr, &profile)?;
    } else {
        return Err(StdError::generic_err("Already following this trader"));
    }
    
    // Create staking message (in a real implementation, this would interact with a staking module)
    let stake_msg = CosmosMsg::Bank(BankMsg::Send {
        to_address: trader.clone(),
        amount: vec![Coin {
            denom: "usocial".to_string(),
            amount: stake_amount,
        }],
    });
    
    // Return response with messages and attributes
    Ok(Response::new()
        .add_message(stake_msg)
        .add_attribute("method", "follow")
        .add_attribute("sender", info.sender.to_string())
        .add_attribute("trader", trader)
        .add_attribute("stake_amount", stake_amount.to_string()))
}

pub fn execute_copy_trade(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    trader: String,
    trade: TradeInfo,
) -> StdResult<Response> {
    // Validate trader address
    let trader_addr = deps.api.addr_validate(&trader)?;
    
    // Get trader profile
    let trader_profile = TRADER_PROFILES.load(deps.storage, &trader_addr)?;
    
    // Check if the sender is authorized to copy trades for this trader
    // In a real implementation, this would have more complex authorization logic
    if info.sender != trader_addr && !trader_profile.followers.iter().any(|f| f.address == info.sender) {
        return Err(StdError::generic_err("Not authorized to copy trades for this trader"));
    }
    
    // Execute the trade (in a real implementation, this would call the AMM contract)
    // For this example, we'll simulate a successful trade with a profit
    let profit = trade.expected_profit;
    
    // Calculate fees
    let config = CONFIG.load(deps.storage)?;
    let trader_fee = profit.multiply_ratio(config.trader_fee_percent, 100u64);
    let treasury_fee = profit.multiply_ratio(config.treasury_fee_percent, 100u64);
    
    // Create fee distribution messages
    let mut messages: Vec<CosmosMsg> = vec![];
    
    // Send trader fee
    if !trader_fee.is_zero() {
        messages.push(CosmosMsg::Bank(BankMsg::Send {
            to_address: trader.clone(),
            amount: vec![Coin {
                denom: "usdc".to_string(),
                amount: trader_fee,
            }],
        }));
    }
    
    // Send treasury fee
    if !treasury_fee.is_zero() {
        messages.push(CosmosMsg::Bank(BankMsg::Send {
            to_address: config.treasury_address.to_string(),
            amount: vec![Coin {
                denom: "usdc".to_string(),
                amount: treasury_fee,
            }],
        }));
    }
    
    // Return response with messages and attributes
    Ok(Response::new()
        .add_messages(messages)
        .add_attribute("method", "copy_trade")
        .add_attribute("sender", info.sender.to_string())
        .add_attribute("trader", trader)
        .add_attribute("trade_type", trade.trade_type)
        .add_attribute("token_pair", trade.token_pair)
        .add_attribute("amount", trade.amount.to_string())
        .add_attribute("profit", profit.to_string())
        .add_attribute("trader_fee", trader_fee.to_string())
        .add_attribute("treasury_fee", treasury_fee.to_string()))
}

pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetTraderProfile { address } => query_trader_profile(deps, address),
        QueryMsg::GetTrader { address } => query_trader(deps, address),
    }
}

fn query_trader_profile(deps: Deps, address: String) -> StdResult<Binary> {
    let addr = deps.api.addr_validate(&address)?;
    let profile = TRADER_PROFILES
        .may_load(deps.storage, &addr)?
        .unwrap_or_else(|| TraderProfile {
            address: addr.clone(),
            followers: vec![],
            total_stake: Uint128::zero(),
            performance: Uint128::zero(),
        });
    
    to_json_binary(&profile)
}

fn query_trader(deps: Deps, address: String) -> StdResult<Binary> {
    let addr = deps.api.addr_validate(&address)?;
    let profile = TRADER_PROFILES
        .may_load(deps.storage, &addr)?
        .unwrap_or_else(|| TraderProfile {
            address: addr.clone(),
            followers: vec![],
            total_stake: Uint128::zero(),
            performance: Uint128::zero(),
        });
    
    // Convert to TraderResponse
    let response = TraderResponse {
        address: addr.to_string(),
        total_stake: profile.total_stake,
        performance: profile.performance,
        followers_count: profile.followers.len() as u64,
    };
    
    to_json_binary(&response)
}
