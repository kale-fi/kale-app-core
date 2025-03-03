use cosmwasm_std::{
    to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128, BankMsg, Coin, CosmosMsg, Addr, StdError, attr,
};

use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{Config, CONFIG, POOLS, Pool};

pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> StdResult<Response> {
    let config = Config {
        owner: deps.api.addr_validate(&msg.owner)?,
        fee_percent: msg.fee_percent,
        yield_percent: msg.yield_percent,
        lp_percent: msg.lp_percent,
        treasury_percent: msg.treasury_percent,
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
        ExecuteMsg::Swap {
            amount_in,
            token_in,
            token_out,
        } => execute_swap(deps, env, info, amount_in, token_in, token_out),
        // Add other execute functions as needed
    }
}

// Helper function to calculate the amount out based on XYK formula
fn calculate_xyk(reserve_in: Uint128, reserve_out: Uint128, amount_in: Uint128) -> StdResult<Uint128> {
    // Formula: amount_out = (reserve_out * amount_in) / (reserve_in + amount_in)
    // To avoid overflow, we use checked operations
    let numerator = reserve_out.checked_mul(amount_in)?;
    let denominator = reserve_in.checked_add(amount_in)?;
    
    // Avoid division by zero
    if denominator.is_zero() {
        return Err(StdError::generic_err("Denominator cannot be zero"));
    }
    
    Ok(numerator.checked_div(denominator)?)
}

// Helper function to split the fee according to the specified percentages
fn split_fee(
    deps: DepsMut,
    fee: Uint128,
    yield_address: &Addr,
    lp_address: &Addr,
    treasury_address: &Addr,
    denom: &str,
) -> StdResult<Vec<CosmosMsg>> {
    let config = CONFIG.load(deps.storage)?;
    
    // Calculate the fee splits
    let total_percent = config.yield_percent + config.lp_percent + config.treasury_percent;
    
    let yield_amount = fee.multiply_ratio(config.yield_percent, total_percent);
    let lp_amount = fee.multiply_ratio(config.lp_percent, total_percent);
    let treasury_amount = fee.multiply_ratio(config.treasury_percent, total_percent);
    
    let mut messages: Vec<CosmosMsg> = vec![];
    
    // Add transfer messages for each recipient
    if !yield_amount.is_zero() {
        messages.push(CosmosMsg::Bank(BankMsg::Send {
            to_address: yield_address.to_string(),
            amount: vec![Coin {
                denom: denom.to_string(),
                amount: yield_amount,
            }],
        }));
    }
    
    if !lp_amount.is_zero() {
        messages.push(CosmosMsg::Bank(BankMsg::Send {
            to_address: lp_address.to_string(),
            amount: vec![Coin {
                denom: denom.to_string(),
                amount: lp_amount,
            }],
        }));
    }
    
    if !treasury_amount.is_zero() {
        messages.push(CosmosMsg::Bank(BankMsg::Send {
            to_address: treasury_address.to_string(),
            amount: vec![Coin {
                denom: denom.to_string(),
                amount: treasury_amount,
            }],
        }));
    }
    
    Ok(messages)
}

pub fn execute_swap(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    amount_in: u128,
    token_in: String,
    token_out: String,
) -> StdResult<Response> {
    // Convert amount_in to Uint128 for safer arithmetic
    let amount_in = Uint128::new(amount_in);
    
    // Get the pool for the token pair
    let pool_key = if token_in < token_out {
        (&token_in, &token_out)
    } else {
        (&token_out, &token_in)
    };
    
    let mut pool = POOLS.load(deps.storage, pool_key)?;
    
    // Get the reserves for the tokens
    let (reserve_in, reserve_out) = if token_in == pool.token_a {
        (pool.reserve_a, pool.reserve_b)
    } else {
        (pool.reserve_b, pool.reserve_a)
    };
    
    // Calculate the swap fee (0.2% of amount_in)
    let fee_rate = Uint128::new(2); // 0.2% = 2/1000
    let fee = amount_in.multiply_ratio(fee_rate, Uint128::new(1000));
    
    // Calculate the amount after fee
    let amount_in_after_fee = amount_in.checked_sub(fee)?;
    
    // Calculate the amount out using the XYK formula
    let amount_out = calculate_xyk(reserve_in, reserve_out, amount_in_after_fee)?;
    
    // Split the fee (50% yield, 30% LP, 20% Treasury)
    // For this example, we'll use placeholder addresses
    let yield_address = deps.api.addr_validate("kale1yield")?;
    let lp_address = deps.api.addr_validate("kale1lp")?;
    let treasury_address = deps.api.addr_validate("kale1treasury")?;
    
    let fee_messages = split_fee(
        deps,
        fee,
        &yield_address,
        &lp_address,
        &treasury_address,
        &token_in,
    )?;
    
    // Update the pool reserves
    if token_in == pool.token_a {
        pool.reserve_a = pool.reserve_a.checked_add(amount_in_after_fee)?;
        pool.reserve_b = pool.reserve_b.checked_sub(amount_out)?;
    } else {
        pool.reserve_b = pool.reserve_b.checked_add(amount_in_after_fee)?;
        pool.reserve_a = pool.reserve_a.checked_sub(amount_out)?;
    }
    
    // Save the updated pool
    POOLS.save(deps.storage, pool_key, &pool)?;
    
    // Transfer the tokens to the user
    let transfer_msg = CosmosMsg::Bank(BankMsg::Send {
        to_address: info.sender.to_string(),
        amount: vec![Coin {
            denom: token_out,
            amount: amount_out,
        }],
    });
    
    // Combine all messages
    let mut messages = fee_messages;
    messages.push(transfer_msg);
    
    // Return response with messages and attributes
    Ok(Response::new()
        .add_messages(messages)
        .add_attributes(vec![
            attr("method", "swap"),
            attr("sender", info.sender),
            attr("token_in", token_in),
            attr("token_out", token_out),
            attr("amount_in", amount_in.to_string()),
            attr("amount_out", amount_out.to_string()),
            attr("fee", fee.to_string()),
        ]))
}

pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        // Add query functions as needed
    }
}
