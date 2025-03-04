use cosmwasm_std::{
    to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128, BankMsg, Coin, CosmosMsg, Addr, StdError, attr, Decimal,
};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg, StakerInfoResponse};
use crate::state::{Config, StakerInfo, CONFIG, STAKERS, TOTAL_STAKED};
use crate::kale_state::{Pool, Staker, POOL};
use crate::kale_msg::{PoolResponse, TotalStakedResponse};

// Constants for APY calculation
const MIN_APY: u64 = 8;  // 8% minimum APY
const MAX_APY: u64 = 12; // 12% maximum APY
const APY_RANGE: u64 = MAX_APY - MIN_APY; // 4% range
const BASE_POINTS: u64 = 10000; // For percentage calculations
const SECONDS_PER_YEAR: u64 = 31536000; // 365 days in seconds
const KALE_RESERVE: u128 = 1_000_000_000_000; // 1M KALE in smallest denomination (assuming 6 decimals)
const FEE_YIELD_PERCENT: u64 = 50; // 50% of fees go to yield

// Execute functions
pub fn execute_stake(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    amount: Uint128,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Ensure the user sent the correct token
    if info.funds.len() != 1 || info.funds[0].denom != config.token_denom {
        return Err(ContractError::InvalidToken {});
    }
    
    if info.funds[0].amount != amount {
        return Err(ContractError::InvalidAmount {});
    }
    
    // Update staker info
    let staker_addr = info.sender.clone();
    let mut staker = STAKERS.may_load(deps.storage, &staker_addr)?.unwrap_or_else(|| StakerInfo {
        amount: Uint128::zero(),
        last_stake_time: env.block.time.seconds(),
        last_claim_time: env.block.time.seconds(),
        accumulated_rewards: Uint128::zero(),
    });
    
    // If there are unclaimed rewards, calculate and store them
    if staker.amount > Uint128::zero() {
        let unclaimed_rewards = calculate_yield(
            deps.as_ref(),
            &env,
            &staker,
            config.fee_pool,
        )?;
        staker.accumulated_rewards += unclaimed_rewards;
    }
    
    // Update staker info
    staker.amount += amount;
    staker.last_stake_time = env.block.time.seconds();
    STAKERS.save(deps.storage, &staker_addr, &staker)?;
    
    // Update total staked
    let mut total = TOTAL_STAKED.load(deps.storage)?;
    total += amount;
    TOTAL_STAKED.save(deps.storage, &total)?;
    
    // Also update the kale_state Staker if it exists
    let kale_staker = Staker {
        address: staker_addr.clone(),
        staked_amount: staker.amount,
        staked_since: staker.last_stake_time,
        last_claim_time: staker.last_claim_time,
        accumulated_rewards: staker.accumulated_rewards,
    };
    
    // Update the pool in kale_state
    let mut pool = POOL.may_load(deps.storage)?.unwrap_or_else(|| Pool {
        usdc: Uint128::zero(),
        kale: Uint128::zero(),
    });
    pool.kale += amount;
    POOL.save(deps.storage, &pool)?;
    
    Ok(Response::new()
        .add_attribute("method", "stake")
        .add_attribute("staker", info.sender)
        .add_attribute("amount", amount))
}

pub fn execute_unstake(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    amount: Uint128,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Get staker info
    let staker_addr = info.sender.clone();
    let mut staker = STAKERS.load(deps.storage, &staker_addr)?;
    
    // Check if lock period has passed
    let current_time = env.block.time.seconds();
    if current_time < staker.last_stake_time + config.lock_period {
        return Err(ContractError::StakeLocked {});
    }
    
    // Check if staker has enough staked
    if staker.amount < amount {
        return Err(ContractError::InsufficientFunds {});
    }
    
    // Calculate any unclaimed rewards
    let unclaimed_rewards = calculate_yield(
        deps.as_ref(),
        &env,
        &staker,
        config.fee_pool,
    )?;
    staker.accumulated_rewards += unclaimed_rewards;
    
    // Update staker info
    staker.amount -= amount;
    staker.last_stake_time = current_time;
    STAKERS.save(deps.storage, &staker_addr, &staker)?;
    
    // Update total staked
    let mut total = TOTAL_STAKED.load(deps.storage)?;
    total -= amount;
    TOTAL_STAKED.save(deps.storage, &total)?;
    
    // Update the pool in kale_state
    let mut pool = POOL.load(deps.storage)?;
    pool.kale -= amount;
    POOL.save(deps.storage, &pool)?;
    
    // Send tokens back to staker
    let msg = BankMsg::Send {
        to_address: info.sender.to_string(),
        amount: vec![Coin {
            denom: config.token_denom,
            amount,
        }],
    };
    
    Ok(Response::new()
        .add_message(msg)
        .add_attribute("method", "unstake")
        .add_attribute("staker", info.sender)
        .add_attribute("amount", amount))
}

pub fn execute_claim(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Get staker info
    let staker_addr = info.sender.clone();
    let mut staker = STAKERS.load(deps.storage, &staker_addr)?;
    
    // Calculate rewards
    let current_rewards = calculate_yield(
        deps.as_ref(),
        &env,
        &staker,
        config.fee_pool,
    )?;
    
    // Add current rewards to accumulated rewards
    let total_rewards = staker.accumulated_rewards + current_rewards;
    
    // Check if there are rewards to claim
    if total_rewards.is_zero() {
        return Err(ContractError::NoRewards {});
    }
    
    // Reset accumulated rewards
    staker.accumulated_rewards = Uint128::zero();
    staker.last_claim_time = env.block.time.seconds();
    STAKERS.save(deps.storage, &staker_addr, &staker)?;
    
    // Update the pool in kale_state
    let mut pool = POOL.load(deps.storage)?;
    pool.usdc = pool.usdc.checked_sub(total_rewards).unwrap_or_default();
    POOL.save(deps.storage, &pool)?;
    
    // Send rewards to staker
    let msg = BankMsg::Send {
        to_address: info.sender.to_string(),
        amount: vec![Coin {
            denom: "usdc".to_string(), // Rewards are paid in USDC
            amount: total_rewards,
        }],
    };
    
    // Update fee pool
    let mut config = CONFIG.load(deps.storage)?;
    config.fee_pool = config.fee_pool.checked_sub(total_rewards)?;
    CONFIG.save(deps.storage, &config)?;
    
    Ok(Response::new()
        .add_message(msg)
        .add_attribute("method", "claim_rewards")
        .add_attribute("staker", info.sender)
        .add_attribute("rewards", total_rewards))
}

// Calculate yield based on stake amount, time, and available fee pool
pub fn calculate_yield(
    deps: Deps,
    env: &Env,
    staker: &StakerInfo,
    fee_pool: Uint128,
) -> StdResult<Uint128> {
    // Get total staked
    let total_staked = TOTAL_STAKED.load(deps.storage)?;
    
    // If nothing is staked or this staker has no stake, return zero
    if total_staked.is_zero() || staker.amount.is_zero() {
        return Ok(Uint128::zero());
    }
    
    // Calculate the staker's share of the total stake
    let stake_ratio = Decimal::from_ratio(staker.amount, total_staked);
    
    // Calculate time since last claim
    let current_time = env.block.time.seconds();
    let time_since_claim = current_time.saturating_sub(staker.last_claim_time);
    
    // Calculate the base APY based on the total staked vs KALE reserve
    // As more is staked, APY decreases from MAX to MIN
    let utilization_ratio = Decimal::from_ratio(total_staked, Uint128::new(KALE_RESERVE));
    let utilization_factor = if utilization_ratio > Decimal::one() {
        Decimal::one() // Cap at 100% utilization
    } else {
        utilization_ratio
    };
    
    // Calculate APY: MAX_APY - (utilization_factor * APY_RANGE)
    let apy_decimal = Decimal::from_ratio(MAX_APY * BASE_POINTS - (utilization_factor * Decimal::from_ratio(APY_RANGE * BASE_POINTS, 1u128)).to_uint_floor(), BASE_POINTS);
    
    // Calculate the yield for the time period
    // yield = stake_amount * apy * (time_since_claim / seconds_per_year)
    let time_factor = Decimal::from_ratio(time_since_claim, SECONDS_PER_YEAR);
    let yield_amount = staker.amount * apy_decimal * time_factor;
    
    // Calculate the staker's share of the fee pool
    let fee_share = fee_pool * stake_ratio;
    
    // The actual yield is the minimum of the calculated yield and the staker's fee share
    // This ensures we don't pay out more than what's available in the fee pool
    let actual_yield = std::cmp::min(yield_amount, fee_share);
    
    Ok(actual_yield)
}

// Add fee to the yield pool (called by AMM contract)
pub fn add_fee_to_pool(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    amount: Uint128,
) -> Result<Response, ContractError> {
    // Only authorized contracts can add fees
    let config = CONFIG.load(deps.storage)?;
    if !config.authorized_contracts.contains(&info.sender) {
        return Err(ContractError::Unauthorized {});
    }
    
    // Update fee pool
    let mut config = CONFIG.load(deps.storage)?;
    config.fee_pool += amount;
    CONFIG.save(deps.storage, &config)?;
    
    // Update the pool in kale_state
    let mut pool = POOL.load(deps.storage)?;
    pool.usdc += amount;
    POOL.save(deps.storage, &pool)?;
    
    Ok(Response::new()
        .add_attribute("method", "add_fee_to_pool")
        .add_attribute("amount", amount))
}

// Helper function to update the config
pub fn update_config(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    lock_period: Option<u64>,
    authorized_contracts: Option<Vec<String>>,
) -> Result<Response, ContractError> {
    let mut config = CONFIG.load(deps.storage)?;
    
    // Only owner can update config
    if info.sender != config.owner {
        return Err(ContractError::Unauthorized {});
    }
    
    // Update lock period if provided
    if let Some(period) = lock_period {
        config.lock_period = period;
    }
    
    // Update authorized contracts if provided
    if let Some(contracts) = authorized_contracts {
        config.authorized_contracts = contracts
            .iter()
            .map(|addr| deps.api.addr_validate(addr))
            .collect::<StdResult<Vec<Addr>>>()?;
    }
    
    CONFIG.save(deps.storage, &config)?;
    
    Ok(Response::new().add_attribute("method", "update_config"))
}
