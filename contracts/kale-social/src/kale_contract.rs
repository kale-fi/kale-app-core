use cosmwasm_std::{
    to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128, BankMsg, Coin, CosmosMsg, StdError, Addr, attr,
};

use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg, TradeMsg, TraderResponse};
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

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", msg.owner)
        .add_attribute("trader_fee_percent", msg.trader_fee_percent.to_string())
        .add_attribute("treasury_fee_percent", msg.treasury_fee_percent.to_string())
        .add_attribute("treasury_address", msg.treasury_address))
}

pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> StdResult<Response> {
    match msg {
        ExecuteMsg::KaleFollow { trader, stake_amount } => {
            execute_follow(deps, env, info, trader, stake_amount)
        },
        ExecuteMsg::KaleCopyTrade { trader, trade } => {
            execute_copy_trade(deps, env, info, trader, trade)
        },
        ExecuteMsg::Follow { trader, stake_amount } => {
            execute_follow(deps, env, info, trader, stake_amount)
        },
        ExecuteMsg::CopyTrade { trader, trade: trade_info } => {
            // Convert TradeInfo to TradeMsg for compatibility
            let trade = TradeMsg {
                amount: trade_info.amount,
                token_in: trade_info.token_pair.split('/').next().unwrap_or("").to_string(),
                token_out: trade_info.token_pair.split('/').last().unwrap_or("").to_string(),
            };
            execute_copy_trade(deps, env, info, trader, trade)
        },
        ExecuteMsg::UpdateProfile { bio, avatar_url } => {
            // Implementation for updating profile
            Ok(Response::new().add_attribute("method", "update_profile"))
        },
        ExecuteMsg::Unfollow { trader } => {
            // Implementation for unfollowing
            Ok(Response::new().add_attribute("method", "unfollow"))
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
    
    // Check if the follower has sent enough tokens for staking
    let sent_funds = info.funds.iter().find(|coin| coin.denom == "usocial");
    if sent_funds.is_none() || sent_funds.unwrap().amount < stake_amount {
        return Err(StdError::generic_err("Not enough usocial tokens sent for staking"));
    }
    
    // Get or create trader profile
    let mut trader_profile = TRADER_PROFILES
        .may_load(deps.storage, &trader_addr)?
        .unwrap_or_else(|| TraderProfile {
            address: trader_addr.clone(),
            total_stake: Uint128::zero(),
            performance: Uint128::zero(),
            followers: vec![],
        });
    
    // Create new follower
    let follower = Follower {
        address: info.sender.clone(),
        stake_amount,
    };
    
    // Add follower if not already following
    if !trader_profile.followers.iter().any(|f| f.address == info.sender) {
        trader_profile.followers.push(follower);
        trader_profile.total_stake += stake_amount;
    }
    
    // Save updated trader profile
    TRADER_PROFILES.save(deps.storage, &trader_addr, &trader_profile)?;
    
    // Return success response
    Ok(Response::new()
        .add_attribute("method", "follow")
        .add_attribute("follower", info.sender.to_string())
        .add_attribute("trader", trader)
        .add_attribute("stake_amount", stake_amount.to_string()))
}

pub fn execute_copy_trade(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    trader: String,
    trade: TradeMsg,
) -> StdResult<Response> {
    // Validate trader address
    let trader_addr = deps.api.addr_validate(&trader)?;
    
    // Get trader profile
    let mut trader_profile = TRADER_PROFILES.load(deps.storage, &trader_addr)?;
    
    // Check if follower is following the trader
    if !trader_profile.followers.iter().any(|f| f.address == info.sender) {
        return Err(StdError::generic_err("Not following this trader"));
    }
    
    // Execute the trade (in a real implementation, this would interact with an AMM contract)
    // For this example, we'll simulate a successful trade with a profit of 5% of the trade amount
    let profit = trade.amount.multiply_ratio(5u128, 100u128);
    
    // Get config for fee calculations
    let config = CONFIG.load(deps.storage)?;
    
    // Calculate fees (10% trader fee, 2% Treasury)
    let trader_fee = profit.multiply_ratio(config.trader_fee_percent, 100u64);
    let treasury_fee = profit.multiply_ratio(config.treasury_fee_percent, 100u64);
    
    // Update trader's performance
    trader_profile.performance = trader_profile.performance.checked_add(trader_fee)?;
    TRADER_PROFILES.save(deps.storage, &trader_addr, &trader_profile)?;
    
    // Create transfer messages for fees
    let mut messages: Vec<CosmosMsg> = vec![];
    
    // Transfer trader fee
    if !trader_fee.is_zero() {
        messages.push(CosmosMsg::Bank(BankMsg::Send {
            to_address: trader_addr.to_string(),
            amount: vec![Coin {
                denom: "usdc".to_string(),
                amount: trader_fee,
            }],
        }));
    }
    
    // Transfer treasury fee
    if !treasury_fee.is_zero() {
        messages.push(CosmosMsg::Bank(BankMsg::Send {
            to_address: config.treasury_address.to_string(),
            amount: vec![Coin {
                denom: "usdc".to_string(),
                amount: treasury_fee,
            }],
        }));
    }
    
    // Return success response with fee distribution
    Ok(Response::new()
        .add_messages(messages)
        .add_attribute("method", "copy_trade")
        .add_attribute("follower", info.sender.to_string())
        .add_attribute("trader", trader)
        .add_attribute("token_in", trade.token_in)
        .add_attribute("token_out", trade.token_out)
        .add_attribute("amount", trade.amount.to_string())
        .add_attribute("profit", profit.to_string())
        .add_attribute("trader_fee", trader_fee.to_string())
        .add_attribute("treasury_fee", treasury_fee.to_string()))
}

pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetTrader { address } => query_trader_profile(deps, address),
        QueryMsg::GetTraderProfile { address } => query_trader_profile(deps, address),
    }
}

fn query_trader_profile(deps: Deps, address: String) -> StdResult<Binary> {
    let addr = deps.api.addr_validate(&address)?;
    let profile = TRADER_PROFILES
        .may_load(deps.storage, &addr)?
        .unwrap_or_else(|| TraderProfile {
            address: addr.clone(),
            total_stake: Uint128::zero(),
            performance: Uint128::zero(),
            followers: vec![],
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
