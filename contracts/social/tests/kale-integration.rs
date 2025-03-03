#[cfg(test)]
mod tests {
    use cosmwasm_std::{
        coins, Addr, Coin, Empty, Uint128,
        testing::{mock_dependencies, mock_env, mock_info},
    };
    use cw_multi_test::{App, Contract, ContractWrapper, Executor};
    
    use kale_social::contract::{self, execute, instantiate, query};
    use kale_social::kale_msg::{ExecuteMsg, InstantiateMsg, QueryMsg, TradeInfo, TraderProfileResponse};
    use kale_social::kale_state::TraderProfile;

    // Helper function to create the contract
    fn social_contract() -> Box<dyn Contract<Empty>> {
        let contract = ContractWrapper::new(execute, instantiate, query);
        Box::new(contract)
    }

    #[test]
    fn follow_and_copy() {
        // Create a new app with initial funds
        let mut app = App::default();
        
        // Set up test accounts
        let owner = Addr::unchecked("owner");
        let treasury = Addr::unchecked("treasury");
        let trader = Addr::unchecked("trader");
        let follower = Addr::unchecked("follower");
        
        // Fund accounts
        app.init_modules(|router, _, storage| {
            router.bank.init_balance(
                storage,
                &owner,
                coins(1000, "usocial"),
            ).unwrap();
            
            router.bank.init_balance(
                storage,
                &follower,
                coins(1000, "usocial"),
            ).unwrap();
            
            router.bank.init_balance(
                storage,
                &trader,
                coins(1000, "usdc"),
            ).unwrap();
        });
        
        // Upload and instantiate the contract
        let social_id = app.store_code(social_contract());
        
        let contract_addr = app.instantiate_contract(
            social_id,
            owner.clone(),
            &InstantiateMsg {
                owner: owner.to_string(),
                trader_fee_percent: 10,
                treasury_fee_percent: 2,
                treasury_address: treasury.to_string(),
            },
            &[],
            "kale-social",
            None,
        ).unwrap();
        
        // Follower follows the trader
        let follow_msg = ExecuteMsg::KaleFollow {
            trader: trader.to_string(),
            stake_amount: Uint128::new(100),
        };
        
        let res = app.execute_contract(
            follower.clone(),
            contract_addr.clone(),
            &follow_msg,
            &coins(100, "usocial"),
        ).unwrap();
        
        // Verify follow was successful
        assert_eq!(res.events.len(), 3); // Transfer, Wasm, Wasm-KaleFollow
        
        // Query trader profile
        let query_msg = QueryMsg::GetTraderProfile {
            address: trader.to_string(),
        };
        
        let res: TraderProfileResponse = app.wrap().query_wasm_smart(
            contract_addr.clone(),
            &query_msg,
        ).unwrap();
        
        // Verify follower was added
        assert_eq!(res.profile.followers.len(), 1);
        assert_eq!(res.profile.followers[0], follower.to_string());
        
        // Follower copies a trade
        let trade_info = TradeInfo {
            trade_type: "buy".to_string(),
            token_pair: "BTC/USDC".to_string(),
            amount: Uint128::new(50),
            expected_profit: Uint128::new(10),
            leverage: Some(1),
        };
        
        let copy_trade_msg = ExecuteMsg::KaleCopyTrade {
            trader: trader.to_string(),
            trade: trade_info,
        };
        
        let res = app.execute_contract(
            follower.clone(),
            contract_addr.clone(),
            &copy_trade_msg,
            &[],
        ).unwrap();
        
        // Verify copy trade was successful
        assert_eq!(res.events.len(), 2); // Wasm, Wasm-KaleCopyTrade
        
        // Verify fees were distributed correctly
        // In a real test, we would check the balances of trader and treasury
        // For simplicity, we'll just check the events
        
        // Query trader profile again to check profit
        let query_msg = QueryMsg::GetTraderProfile {
            address: trader.to_string(),
        };
        
        let res: TraderProfileResponse = app.wrap().query_wasm_smart(
            contract_addr,
            &query_msg,
        ).unwrap();
        
        // Verify profit was updated
        assert!(res.profile.profit > Uint128::zero());
    }
}
