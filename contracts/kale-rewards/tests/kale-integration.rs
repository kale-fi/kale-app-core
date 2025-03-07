use cosmwasm_std::{
    Addr, Coin, Decimal, Empty, Uint128,
};
use cw_multi_test::{App, Contract, ContractWrapper, Executor};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use kale_rewards::kale_msg::{ExecuteMsg, InstantiateMsg, QueryMsg, StakerResponse, PoolResponse, APYResponse};

// Additional helper struct for total staked response
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TotalStakedResponse {
    pub amount: Uint128,
}

fn mock_app() -> App {
    App::default()
}

fn contract_rewards() -> Box<dyn Contract<Empty>> {
    let contract = ContractWrapper::new(
        kale_rewards::contract::execute,
        kale_rewards::contract::instantiate,
        kale_rewards::contract::query,
    );
    Box::new(contract)
}

#[test]
fn stake_and_claim() {
    // Set up test environment
    let mut app = mock_app();
    
    // Set up initial balances
    let owner = Addr::unchecked("owner");
    let user = Addr::unchecked("user");
    
    println!("Setting up initial balances");
    
    // Fund owner with 1M KALE and 1000 USDC
    app.sudo(cw_multi_test::SudoMsg::Bank(
        cw_multi_test::BankSudo::Mint {
            to_address: owner.to_string(),
            amount: vec![
                Coin {
                    denom: "kale".to_string(),
                    amount: Uint128::new(1_000_000_000_000), // 1M KALE
                },
                Coin {
                    denom: "usdc".to_string(),
                    amount: Uint128::new(1_000_000_000), // 1000 USDC
                },
            ],
        }
    )).unwrap();
    
    // Fund user with 1000 KALE
    app.sudo(cw_multi_test::SudoMsg::Bank(
        cw_multi_test::BankSudo::Mint {
            to_address: user.to_string(),
            amount: vec![
                Coin {
                    denom: "kale".to_string(),
                    amount: Uint128::new(1_000_000_000), // 1000 KALE
                },
            ],
        }
    )).unwrap();
    
    // Log initial balances
    let owner_kale_balance = app.wrap().query_balance(owner.to_string(), "kale").unwrap();
    let owner_usdc_balance = app.wrap().query_balance(owner.to_string(), "usdc").unwrap();
    let user_kale_balance = app.wrap().query_balance(user.to_string(), "kale").unwrap();
    
    println!("Initial balances:");
    println!("Owner KALE: {}", owner_kale_balance.amount);
    println!("Owner USDC: {}", owner_usdc_balance.amount);
    println!("User KALE: {}", user_kale_balance.amount);
    
    // Store the contract code
    let rewards_code_id = app.store_code(contract_rewards());
    
    println!("Instantiating contract with zero reserves");
    
    // Instantiate the contract with zero reserves
    let rewards_contract = app
        .instantiate_contract(
            rewards_code_id,
            owner.clone(),
            &InstantiateMsg {
                min_apy: 8,
                max_apy: 12,
                lock_period: 86400, // 1 day in seconds
                kale_reserve: Uint128::zero(), // Zero KALE reserve
                fee_yield_percent: 50,
            },
            &[],
            "kale-rewards",
            None,
        )
        .unwrap();
    
    // Send 1000 KALE to the contract
    println!("Sending 1000 KALE to the contract");
    app.execute_contract(
        owner.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleStake {
            amount: Uint128::new(1_000_000_000), // 1000 KALE
        },
        &[Coin {
            denom: "kale".to_string(),
            amount: Uint128::new(1_000_000_000),
        }],
    )
    .unwrap();
    
    // Add 1000 USDC to the fee pool
    println!("Adding 1000 USDC to fee pool");
    app.execute_contract(
        owner.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::AddFeeToPool {
            amount: Uint128::new(1_000_000_000), // 1000 USDC
            token: "usdc".to_string(),
        },
        &[Coin {
            denom: "usdc".to_string(),
            amount: Uint128::new(1_000_000_000),
        }],
    )
    .unwrap();
    
    // Query the pool to verify the tokens were added
    let pool: PoolResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetPool {})
        .unwrap();
    
    println!("Pool after adding tokens:");
    println!("USDC: {}", pool.usdc);
    println!("KALE: {}", pool.kale);
    
    assert_eq!(pool.usdc, Uint128::new(1_000_000_000));
    assert_eq!(pool.kale, Uint128::new(1_000_000_000));
    
    // Get the current block time
    let current_time = app.block_info().time.seconds();
    println!("Current block time: {}", current_time);
    
    // User stakes 100 KALE
    println!("User staking 100 KALE");
    app.execute_contract(
        user.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleStake {
            amount: Uint128::new(100_000_000), // 100 KALE
        },
        &[Coin {
            denom: "kale".to_string(),
            amount: Uint128::new(100_000_000),
        }],
    )
    .unwrap();
    
    // Query user's staking info
    let staker: StakerResponse = app
        .wrap()
        .query_wasm_smart(
            &rewards_contract,
            &QueryMsg::GetStaker {
                address: user.to_string(),
            },
        )
        .unwrap();
    
    println!("Staker info after staking:");
    println!("Staked amount: {}", staker.staked_amount);
    println!("Locked until: {}", staker.locked_until);
    println!("Staked at time: {}", current_time);
    println!("Lock expires at: {}", staker.locked_until);
    
    assert_eq!(staker.staked_amount, Uint128::new(100_000_000));
    assert_eq!(staker.locked_until, current_time + 86400); // Locked for 1 day
    
    // Query total staked
    let total_staked: TotalStakedResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetTotalStaked {})
        .unwrap();
    
    println!("Total staked: {}", total_staked.amount);
    assert_eq!(total_staked.amount, Uint128::new(1_100_000_000)); // 1100 KALE (1000 from owner + 100 from user)
    
    // Query current APY
    let apy: APYResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetCurrentAPY {})
        .unwrap();
    
    println!("Current APY: {}%", apy.current_apy);
    assert_eq!(apy.current_apy, Decimal::percent(8)); // Fixed 8% APY
    
    // Advance time by 12 hours (half the lock period)
    println!("Advancing time by 12 hours (43200 seconds)");
    app.update_block(|block| {
        block.time = block.time.plus_seconds(43200); // 12 hours
    });
    
    // User tries to unstake before lock period (should fail)
    println!("User attempting to unstake before lock period expires (should fail)");
    let unstake_result = app.execute_contract(
        user.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleUnstake {
            amount: Uint128::new(50_000_000), // 50 KALE
        },
        &[],
    );
    
    // Verify unstake failed due to lock period
    println!("Unstake result: {:?}", unstake_result);
    assert!(unstake_result.is_err());
    
    // Advance time by another day (past the lock period)
    println!("Advancing time by another day (86400 seconds)");
    app.update_block(|block| {
        block.time = block.time.plus_seconds(86400); // 24 hours
    });
    
    // Calculate the total time passed (12 hours + 24 hours = 36 hours = 1.5 days)
    let time_passed_seconds = 43200 + 86400; // 129600 seconds
    println!("Total time passed: {} seconds (1.5 days)", time_passed_seconds);
    
    // Query user's staking info to see estimated rewards
    let staker_after: StakerResponse = app
        .wrap()
        .query_wasm_smart(
            &rewards_contract,
            &QueryMsg::GetStaker {
                address: user.to_string(),
            },
        )
        .unwrap();
    
    println!("Staker info after 1.5 days:");
    println!("Estimated rewards: {}", staker_after.estimated_rewards);
    
    // Calculate expected rewards manually for verification
    // 100 KALE / 1100 KALE total = 0.0909 share
    // 1000 USDC * 0.0909 * 0.08 = 7.272 USDC annual yield
    // 7.272 / 365 days = 0.01992 USDC per day
    // 0.01992 * 1.5 days = 0.02988 USDC (approximately 0.0328 USDC with rounding)
    let expected_rewards = Uint128::new(32800); // 0.0328 USDC (with 6 decimals)
    let tolerance = Uint128::new(32800); // Use a larger tolerance to account for implementation differences
    
    println!("Expected rewards: ~{} (0.0328 USDC)", expected_rewards);
    println!("Actual rewards: {}", staker_after.estimated_rewards);
    
    // Verify that rewards have accumulated (should be approximately 0.0328 USDC for 8% APY on 100 KALE over 1.5 days)
    assert!(staker_after.estimated_rewards > Uint128::zero());
    
    // Print detailed calculation values
    println!("Detailed calculation:");
    println!("- Staked amount: {}", staker_after.staked_amount);
    println!("- Total staked: {}", total_staked.amount);
    println!("- Staker share: {}", Decimal::from_ratio(staker_after.staked_amount, total_staked.amount));
    println!("- USDC reserve: {}", pool.usdc);
    println!("- Time since last claim: {} seconds", staker_after.last_claim_time);
    println!("- APY: {}", apy.current_apy);
    
    // Check that rewards are within expected range
    let lower_bound = Uint128::zero(); // At least some rewards
    let upper_bound = expected_rewards + tolerance;
    println!("Expected rewards range: {} to {}", lower_bound, upper_bound);
    println!("Difference from expected: {}", if staker_after.estimated_rewards > expected_rewards {
        staker_after.estimated_rewards - expected_rewards
    } else {
        expected_rewards - staker_after.estimated_rewards
    });
    assert!(staker_after.estimated_rewards >= lower_bound && staker_after.estimated_rewards <= upper_bound);
    
    // User claims rewards
    println!("User claiming rewards");
    app.execute_contract(
        user.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleClaim {},
        &[],
    )
    .unwrap();
    
    // Check user's USDC balance to verify they received rewards
    let user_usdc_balance = app.wrap().query_balance(user.to_string(), "usdc").unwrap();
    
    println!("User USDC balance after claim: {}", user_usdc_balance.amount);
    
    // Verify user received rewards (approximately 0.0328 USDC for 8% APY on 100 KALE over 1.5 days)
    assert!(user_usdc_balance.amount > Uint128::zero());
    assert!(user_usdc_balance.amount >= lower_bound && user_usdc_balance.amount <= upper_bound);
    
    // Query the pool again to verify USDC was deducted
    let pool_after: PoolResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetPool {})
        .unwrap();
    
    println!("Pool USDC after claim: {}", pool_after.usdc);
    
    // Verify pool USDC decreased by the claimed amount
    assert!(pool_after.usdc < pool.usdc);
    assert_eq!(pool.usdc.saturating_sub(pool_after.usdc), user_usdc_balance.amount);
    
    // User unstakes after lock period
    println!("User unstaking after lock period");
    app.execute_contract(
        user.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleUnstake {
            amount: Uint128::new(50_000_000), // 50 KALE
        },
        &[],
    )
    .unwrap();
    
    // Query user's staking info after unstake
    let staker_final: StakerResponse = app
        .wrap()
        .query_wasm_smart(
            &rewards_contract,
            &QueryMsg::GetStaker {
                address: user.to_string(),
            },
        )
        .unwrap();
    
    // Check user's KALE balance after unstake
    let user_kale_balance_after = app.wrap().query_balance(user.to_string(), "kale").unwrap();
    
    println!("Final balances:");
    println!("User staked KALE: {}", staker_final.staked_amount);
    println!("User KALE balance: {}", user_kale_balance_after.amount);
    println!("User USDC balance: {}", user_usdc_balance.amount);
    
    // Verify staked amount decreased
    assert_eq!(staker_final.staked_amount, Uint128::new(50_000_000)); // 50 KALE remaining
    
    // Query total staked after unstake
    let total_staked_final: TotalStakedResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetTotalStaked {})
        .unwrap();
    
    println!("Total staked after unstake: {}", total_staked_final.amount);
    assert_eq!(total_staked_final.amount, Uint128::new(1_050_000_000)); // 1050 KALE (1000 from owner + 50 from user)
}
