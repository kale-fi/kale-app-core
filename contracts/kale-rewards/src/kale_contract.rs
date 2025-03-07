use cosmwasm_std::{
    to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128, BankMsg, Coin, Decimal,
};

use crate::error::ContractError;
use crate::kale_msg::{PoolResponse, StakerResponse, ConfigResponse, APYResponse, TotalStakedResponse};
use crate::kale_state::{Config, Pool, Staker, CONFIG, POOL, STAKERS, TOTAL_STAKED};

// Constants for APY calculation
const MIN_APY: u64 = 8;  // 8% minimum APY
const MAX_APY: u64 = 12; // 12% maximum APY
const APY_RANGE: u64 = MAX_APY - MIN_APY; // 4% range
const BASE_POINTS: u64 = 10000; // For percentage calculations
const SECONDS_PER_YEAR: u64 = 31536000; // 365 days in seconds
const KALE_RESERVE: u128 = 1_000_000_000_000; // 1M KALE in smallest denomination (assuming 6 decimals)
const FEE_YIELD_PERCENT: u64 = 50; // 50% of fees go to yield
const LOCK_PERIOD: u64 = 86400; // 1 day in seconds

// Execute functions
pub fn execute_stake(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    amount: Uint128,
) -> Result<Response, ContractError> {
    // Validate that the user sent KALE tokens
    let sent_funds = info.funds.iter().find(|coin| coin.denom == "kale");
    if sent_funds.is_none() || sent_funds.unwrap().amount != amount {
        return Err(ContractError::InvalidFunds {});
    }
    
    // Get or create staker
    let staker_addr = info.sender.clone();
    let mut staker = STAKERS
        .may_load(deps.storage, &staker_addr)?
        .unwrap_or_else(|| Staker {
            address: staker_addr.clone(),
            staked_amount: Uint128::zero(),
            staked_since: env.block.time.seconds(),
            last_claim_time: env.block.time.seconds(),
            accumulated_rewards: Uint128::zero(),
            locked_until: env.block.time.seconds() + LOCK_PERIOD,
        });
    
    // Update staker info
    staker.staked_amount += amount;
    if staker.staked_amount == amount {
        // First time staking, update the staked_since timestamp
        staker.staked_since = env.block.time.seconds();
        staker.last_claim_time = env.block.time.seconds();
    }
    
    // Set the lock period for the stake - 1 day (86400 seconds)
    staker.locked_until = env.block.time.seconds() + LOCK_PERIOD;
    
    // Save staker info
    STAKERS.save(deps.storage, &staker_addr, &staker)?;
    
    // Update total staked
    let mut total_staked = TOTAL_STAKED.load(deps.storage)?;
    total_staked += amount;
    TOTAL_STAKED.save(deps.storage, &total_staked)?;
    
    // Update pool
    let mut pool = POOL.load(deps.storage)?;
    pool.kale += amount;
    POOL.save(deps.storage, &pool)?;
    
    Ok(Response::new()
        .add_attribute("action", "stake")
        .add_attribute("staker", info.sender)
        .add_attribute("amount", amount.to_string())
        .add_attribute("locked_until", staker.locked_until.to_string()))
}

pub fn execute_unstake(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    amount: Uint128,
) -> Result<Response, ContractError> {
    let staker_addr = info.sender.clone();
    
    // Load staker info
    let mut staker = STAKERS.load(deps.storage, &staker_addr)?;
    
    // Check if staker has enough staked
    if staker.staked_amount < amount {
        return Err(ContractError::InsufficientFunds {});
    }
    
    // Check if lock period has passed
    let current_time = env.block.time.seconds();
    if current_time < staker.locked_until {
        return Err(ContractError::StakeLocked {});
    }
    
    // Update staker info
    staker.staked_amount -= amount;
    STAKERS.save(deps.storage, &staker_addr, &staker)?;
    
    // Update total staked
    let mut total_staked = TOTAL_STAKED.load(deps.storage)?;
    total_staked -= amount;
    TOTAL_STAKED.save(deps.storage, &total_staked)?;
    
    // Update pool
    let mut pool = POOL.load(deps.storage)?;
    pool.kale -= amount;
    POOL.save(deps.storage, &pool)?;
    
    // Send tokens back to staker
    let msg = BankMsg::Send {
        to_address: staker_addr.to_string(),
        amount: vec![Coin {
            denom: "kale".to_string(),
            amount,
        }],
    };
    
    Ok(Response::new()
        .add_message(msg)
        .add_attribute("action", "unstake")
        .add_attribute("staker", staker_addr)
        .add_attribute("amount", amount.to_string()))
}

pub fn execute_claim(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
) -> Result<Response, ContractError> {
    let staker_addr = info.sender.clone();
    
    // Load staker info
    let mut staker = STAKERS.load(deps.storage, &staker_addr)?;
    
    // Calculate yield
    let rewards = calculate_yield(deps.as_ref(), &env, &staker)?;
    
    if rewards.is_zero() {
        return Err(ContractError::NoRewards {});
    }
    
    // Update pool
    let mut pool = POOL.load(deps.storage)?;
    if pool.usdc < rewards {
        return Err(ContractError::InsufficientFunds {});
    }
    pool.usdc -= rewards;
    POOL.save(deps.storage, &pool)?;
    
    // Update staker
    staker.accumulated_rewards = Uint128::zero();
    staker.last_claim_time = env.block.time.seconds();
    STAKERS.save(deps.storage, &staker_addr, &staker)?;
    
    // Send rewards to staker
    let msg = BankMsg::Send {
        to_address: staker_addr.to_string(),
        amount: vec![Coin {
            denom: "usdc".to_string(),
            amount: rewards,
        }],
    };
    
    Ok(Response::new()
        .add_message(msg)
        .add_attribute("action", "claim")
        .add_attribute("staker", staker_addr)
        .add_attribute("rewards", rewards.to_string()))
}

// The problem is in the calculate_yield function where integer division
// results in the yield_per_second being rounded down to zero.
// Here's the fixed function:

pub fn calculate_yield(
    deps: Deps,
    env: &Env,
    staker: &Staker,
) -> StdResult<Uint128> {
    // Get config and pool
    let config = CONFIG.load(deps.storage)?;
    let pool = POOL.load(deps.storage)?;
    let total_staked = TOTAL_STAKED.load(deps.storage)?;
    
    // Log inputs for debugging
    deps.api.debug(&format!(
        "Calculating yield - Staked amount: {}, Total staked: {}, USDC reserve: {}, Last claim time: {}",
        staker.staked_amount, total_staked, pool.usdc, staker.last_claim_time
    ));
    
    // If no USDC in pool or no tokens staked, return zero
    if pool.usdc.is_zero() || total_staked.is_zero() || staker.staked_amount.is_zero() {
        deps.api.debug("Zero USDC in pool or zero tokens staked, returning zero yield");
        return Ok(Uint128::zero());
    }
    
    // Calculate time since last claim
    let current_time = env.block.time.seconds();
    let time_since_last_claim = current_time.saturating_sub(staker.last_claim_time);
    
    deps.api.debug(&format!(
        "Current time: {}, Time since last claim: {} seconds",
        current_time, time_since_last_claim
    ));
    
    // If no time has passed, return zero
    if time_since_last_claim == 0 {
        deps.api.debug("No time has passed since last claim, returning zero yield");
        return Ok(Uint128::zero());
    }
    
    // Fixed 8% APY as per spec
    let apy = Decimal::from_ratio(8u64, 100u64);
    
    // Calculate the staker's share of the total staked amount
    let staker_share = Decimal::from_ratio(staker.staked_amount, total_staked);
    deps.api.debug(&format!("Staker's share of total staked: {}", staker_share));
    
    // Calculate the annual yield in USDC for the staker
    let annual_yield_usdc = pool.usdc * staker_share * apy;
    deps.api.debug(&format!("Annual yield in USDC: {}", annual_yield_usdc));

    // Instead of dividing the annual yield by seconds per year and then multiplying by time passed,
    // calculate the proportion of a year that has passed and multiply by the annual yield
    let seconds_per_year = 31_536_000u64;
    let time_ratio = Decimal::from_ratio(time_since_last_claim, seconds_per_year);
    
    // Calculate the actual yield for the time period
    let yield_amount = annual_yield_usdc * time_ratio;
    deps.api.debug(&format!(
        "Time ratio: {}, Final yield amount for {} seconds: {}",
        time_ratio, time_since_last_claim, yield_amount
    ));
    
    Ok(yield_amount)
}

// Add fee to the pool (from AMM contract)
pub fn add_fee_to_pool(
    deps: DepsMut,
    _info: MessageInfo,
    amount: Uint128,
    token: String,
) -> Result<Response, ContractError> {
    // Only allow USDC to be added to the pool
    if token != "usdc" {
        return Err(ContractError::InvalidToken {});
    }
    
    // Update pool
    let mut pool = POOL.load(deps.storage)?;
    pool.usdc += amount;
    POOL.save(deps.storage, &pool)?;
    
    Ok(Response::new()
        .add_attribute("action", "add_fee_to_pool")
        .add_attribute("amount", amount.to_string())
        .add_attribute("token", token))
}

// Update config
pub fn update_config(
    deps: DepsMut,
    info: MessageInfo,
    min_apy: Option<u64>,
    max_apy: Option<u64>,
    lock_period: Option<u64>,
    kale_reserve: Option<Uint128>,
    fee_yield_percent: Option<u64>,
) -> Result<Response, ContractError> {
    // Only owner can update config
    let mut config = CONFIG.load(deps.storage)?;
    if info.sender != config.owner {
        return Err(ContractError::Unauthorized {});
    }
    
    // Update config fields
    if let Some(min_apy) = min_apy {
        config.min_apy = min_apy;
    }
    
    if let Some(max_apy) = max_apy {
        config.max_apy = max_apy;
    }
    
    if let Some(lock_period) = lock_period {
        config.lock_period = lock_period;
    }
    
    if let Some(kale_reserve) = kale_reserve {
        config.kale_reserve = kale_reserve;
    }
    
    if let Some(fee_yield_percent) = fee_yield_percent {
        config.fee_yield_percent = fee_yield_percent;
    }
    
    // Save updated config
    CONFIG.save(deps.storage, &config)?;
    
    Ok(Response::new().add_attribute("action", "update_config"))
}

// Query functions
pub fn query_config(deps: Deps) -> StdResult<Binary> {
    let config = CONFIG.load(deps.storage)?;
    
    let response = ConfigResponse {
        owner: config.owner.to_string(),
        min_apy: config.min_apy,
        max_apy: config.max_apy,
        lock_period: config.lock_period,
        kale_reserve: config.kale_reserve,
        fee_yield_percent: config.fee_yield_percent,
    };
    
    to_json_binary(&response)
}

pub fn query_pool(deps: Deps) -> StdResult<Binary> {
    let pool = POOL.load(deps.storage)?;
    
    let response = PoolResponse {
        usdc: pool.usdc,
        kale: pool.kale,
    };
    
    to_json_binary(&response)
}

pub fn query_staker(deps: Deps, env: Env, address: String) -> StdResult<Binary> {
    let addr = deps.api.addr_validate(&address)?;
    
    let staker = STAKERS.may_load(deps.storage, &addr)?.unwrap_or_else(|| Staker {
        address: addr.clone(),
        staked_amount: Uint128::zero(),
        staked_since: env.block.time.seconds(),
        last_claim_time: env.block.time.seconds(),
        accumulated_rewards: Uint128::zero(),
        locked_until: env.block.time.seconds() + LOCK_PERIOD,
    });
    
    // Calculate estimated rewards
    let estimated_rewards = calculate_yield(deps, &env, &staker)?;
    
    let response = StakerResponse {
        address: staker.address.to_string(),
        staked_amount: staker.staked_amount,
        staked_since: staker.staked_since,
        last_claim_time: staker.last_claim_time,
        accumulated_rewards: staker.accumulated_rewards,
        estimated_rewards,
        locked_until: staker.locked_until,
    };
    
    to_json_binary(&response)
}

pub fn query_total_staked(deps: Deps) -> StdResult<Binary> {
    let total = TOTAL_STAKED.load(deps.storage)?;
    
    let response = TotalStakedResponse {
        amount: total,
    };
    
    to_json_binary(&response)
}

pub fn query_current_apy(deps: Deps) -> StdResult<Binary> {
    let config = CONFIG.load(deps.storage)?;
    let pool = POOL.load(deps.storage)?;
    
    // Calculate APY based on pool utilization
    // If kale_reserve is zero, use the minimum APY to avoid division by zero
    let current_apy = if config.kale_reserve.is_zero() {
        Decimal::from_ratio(config.min_apy, 100u64)
    } else {
        let utilization_ratio = Decimal::from_ratio(pool.kale, config.kale_reserve);
        let apy_range = config.max_apy - config.min_apy;
        let apy_adjustment = Decimal::from_ratio(apy_range, 100u64) * utilization_ratio;
        Decimal::from_ratio(config.min_apy, 100u64) + apy_adjustment
    };
    
    let response = APYResponse {
        current_apy,
        min_apy: config.min_apy,
        max_apy: config.max_apy,
    };
    
    to_json_binary(&response)
}
