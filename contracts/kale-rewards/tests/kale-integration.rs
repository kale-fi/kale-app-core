use cosmwasm_std::{
    Addr, Coin, Decimal, Empty, Uint128,
    testing::{mock_dependencies, mock_env, mock_info},
};
use cw_multi_test::{App, Contract, ContractWrapper, Executor};

use kale_rewards::kale_contract::{execute_stake, execute_claim, calculate_yield};
use kale_rewards::kale_msg::{ExecuteMsg, InstantiateMsg, QueryMsg, StakerResponse, PoolResponse, ConfigResponse, APYResponse};
use kale_rewards::kale_state::{Config, Pool, Staker};

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
    let user1 = Addr::unchecked("user1");
    let user2 = Addr::unchecked("user2");
    
    // Fund accounts with KALE and USDC
    app.init_bank_balance(
        &owner,
        vec![
            Coin {
                denom: "kale".to_string(),
                amount: Uint128::new(1_000_000_000_000), // 1M KALE
            },
            Coin {
                denom: "usdc".to_string(),
                amount: Uint128::new(1_000_000_000_000), // 1M USDC
            },
        ],
    )
    .unwrap();
    
    app.init_bank_balance(
        &user1,
        vec![
            Coin {
                denom: "kale".to_string(),
                amount: Uint128::new(10_000_000_000), // 10K KALE
            },
        ],
    )
    .unwrap();
    
    app.init_bank_balance(
        &user2,
        vec![
            Coin {
                denom: "kale".to_string(),
                amount: Uint128::new(5_000_000_000), // 5K KALE
            },
        ],
    )
    .unwrap();
    
    // Store the contract code
    let rewards_code_id = app.store_code(contract_rewards());
    
    // Instantiate the contract
    let rewards_contract = app
        .instantiate_contract(
            rewards_code_id,
            owner.clone(),
            &InstantiateMsg {
                min_apy: 8,
                max_apy: 12,
                lock_period: 86400, // 1 day in seconds
                kale_reserve: Uint128::new(1_000_000_000_000), // 1M KALE
                fee_yield_percent: 50,
            },
            &[],
            "kale-rewards",
            None,
        )
        .unwrap();
    
    // Add some USDC to the fee pool (simulating AMM fees)
    app.execute_contract(
        owner.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::AddFeeToPool {
            amount: Uint128::new(10_000_000_000), // 10K USDC
            token: "usdc".to_string(),
        },
        &[Coin {
            denom: "usdc".to_string(),
            amount: Uint128::new(10_000_000_000),
        }],
    )
    .unwrap();
    
    // Query the pool to verify the USDC was added
    let pool: PoolResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetPool {})
        .unwrap();
    
    assert_eq!(pool.usdc, Uint128::new(10_000_000_000));
    assert_eq!(pool.kale, Uint128::zero());
    
    // User1 stakes 10K KALE
    app.execute_contract(
        user1.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleStake {
            amount: Uint128::new(10_000_000_000),
        },
        &[Coin {
            denom: "kale".to_string(),
            amount: Uint128::new(10_000_000_000),
        }],
    )
    .unwrap();
    
    // Query user1's staking info
    let staker1: StakerResponse = app
        .wrap()
        .query_wasm_smart(
            &rewards_contract,
            &QueryMsg::GetStaker {
                address: user1.to_string(),
            },
        )
        .unwrap();
    
    assert_eq!(staker1.staked_amount, Uint128::new(10_000_000_000));
    
    // User2 stakes 5K KALE
    app.execute_contract(
        user2.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleStake {
            amount: Uint128::new(5_000_000_000),
        },
        &[Coin {
            denom: "kale".to_string(),
            amount: Uint128::new(5_000_000_000),
        }],
    )
    .unwrap();
    
    // Query total staked
    let total_staked: TotalStakedResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetTotalStaked {})
        .unwrap();
    
    assert_eq!(total_staked.amount, Uint128::new(15_000_000_000)); // 15K KALE
    
    // Query current APY
    let apy: APYResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetCurrentAPY {})
        .unwrap();
    
    // Verify APY is within expected range (8-12%)
    assert!(apy.current_apy >= Decimal::percent(8));
    assert!(apy.current_apy <= Decimal::percent(12));
    
    // Advance time by 30 days to accumulate rewards
    app.update_block(|block| {
        block.time = block.time.plus_seconds(30 * 86400);
    });
    
    // Query user1's staking info to see estimated rewards
    let staker1_after: StakerResponse = app
        .wrap()
        .query_wasm_smart(
            &rewards_contract,
            &QueryMsg::GetStaker {
                address: user1.to_string(),
            },
        )
        .unwrap();
    
    // Verify that rewards have accumulated
    assert!(staker1_after.estimated_rewards > Uint128::zero());
    
    // User1 claims rewards
    app.execute_contract(
        user1.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleClaim {},
        &[],
    )
    .unwrap();
    
    // Check user1's USDC balance to verify they received rewards
    let user1_balance = app.wrap().query_balance(user1.to_string(), "usdc").unwrap();
    
    // Verify user received rewards
    assert!(user1_balance.amount > Uint128::zero());
    
    // Query the pool again to verify USDC was deducted
    let pool_after: PoolResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetPool {})
        .unwrap();
    
    // Verify pool USDC decreased by the claimed amount
    assert!(pool_after.usdc < pool.usdc);
    
    // User1 tries to unstake before lock period (should fail)
    let unstake_result = app.execute_contract(
        user1.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleUnstake {
            amount: Uint128::new(1_000_000_000), // 1K KALE
        },
        &[],
    );
    
    // Verify unstake failed due to lock period
    assert!(unstake_result.is_err());
    
    // Advance time past lock period
    app.update_block(|block| {
        block.time = block.time.plus_seconds(86400);
    });
    
    // User1 unstakes after lock period
    app.execute_contract(
        user1.clone(),
        rewards_contract.clone(),
        &ExecuteMsg::KaleUnstake {
            amount: Uint128::new(1_000_000_000), // 1K KALE
        },
        &[],
    )
    .unwrap();
    
    // Query user1's staking info after unstake
    let staker1_final: StakerResponse = app
        .wrap()
        .query_wasm_smart(
            &rewards_contract,
            &QueryMsg::GetStaker {
                address: user1.to_string(),
            },
        )
        .unwrap();
    
    // Verify staked amount decreased
    assert_eq!(staker1_final.staked_amount, Uint128::new(9_000_000_000)); // 9K KALE
    
    // Query total staked after unstake
    let total_staked_final: TotalStakedResponse = app
        .wrap()
        .query_wasm_smart(&rewards_contract, &QueryMsg::GetTotalStaked {})
        .unwrap();
    
    assert_eq!(total_staked_final.amount, Uint128::new(14_000_000_000)); // 14K KALE
}

// Additional helper struct for total staked response
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TotalStakedResponse {
    pub amount: Uint128,
}

use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
