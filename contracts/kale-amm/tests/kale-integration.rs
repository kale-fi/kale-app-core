#[cfg(test)]
mod tests {
    use std::str::FromStr;
    use cosmwasm_std::{
        coins, Addr, Empty, Uint128,
    };
    use cw_multi_test::{App, Contract, ContractWrapper, Executor};
    
    use kale_amm::kale_contract::{execute, instantiate, query};
    use kale_amm::kale_msg::{ExecuteMsg, InstantiateMsg};

    // Helper function to create the contract
    fn amm_contract() -> Box<dyn Contract<Empty>> {
        let contract = ContractWrapper::new(execute, instantiate, query);
        Box::new(contract)
    }

    #[test]
    fn swap_works() {
        // Create a new app with initial funds
        let mut app = App::default();
        
        // Set up test accounts
        let owner = Addr::unchecked("owner");
        let user = Addr::unchecked("user");
        let yield_address = Addr::unchecked("kale1yield");
        let lp_address = Addr::unchecked("kale1lp");
        let treasury_address = Addr::unchecked("kale1treasury");
        
        // Fund accounts
        app.init_modules(|router, _, storage| {
            // Fund owner
            router.bank.init_balance(
                storage,
                &owner,
                coins(2000, "kale") // 1000 for pool, 1000 for other operations
            ).unwrap();
            
            router.bank.init_balance(
                storage,
                &owner,
                coins(2000, "usdc") // 1000 for pool, 1000 for other operations
            ).unwrap();
            
            // Fund user for swapping
            router.bank.init_balance(
                storage,
                &user,
                coins(1000, "kale") // For swap operations
            ).unwrap();
            
            router.bank.init_balance(
                storage,
                &user,
                coins(1000, "usdc") // For swap operations
            ).unwrap();
            
            // Fund fee recipients with empty balances
            router.bank.init_balance(
                storage,
                &yield_address,
                vec![],
            ).unwrap();
            
            router.bank.init_balance(
                storage,
                &lp_address,
                vec![],
            ).unwrap();
            
            router.bank.init_balance(
                storage,
                &treasury_address,
                vec![],
            ).unwrap();
        });
        
        // Upload and instantiate the contract
        let amm_id = app.store_code(amm_contract());
        
        // Instantiate the contract with initial pool reserves
        let contract_addr = app.instantiate_contract(
            amm_id,
            owner.clone(),
            &InstantiateMsg {
                owner: owner.to_string(),
                fee_percent: 2, // 0.2% fee (2/1000)
                yield_percent: 50, // 50% of fee goes to yield
                lp_percent: 30, // 30% of fee goes to LP
                treasury_percent: 20, // 20% of fee goes to Treasury
                token_a: "kale".to_string(),
                token_b: "usdc".to_string(),
                reserves_a: Uint128::new(1000),
                reserves_b: Uint128::new(1000),
            },
            &coins(2000, "kale"), // Send funds for reserves and operations
            "kale-amm",
            None,
        ).unwrap();
        
        // Send USDC to the contract for the pool
        app.send_tokens(
            owner.clone(),
            contract_addr.clone(),
            &coins(1000, "usdc"),
        ).unwrap();
        
        // Verify contract has received the tokens
        let contract_kale_balance = app.wrap().query_balance(contract_addr.clone(), "kale").unwrap().amount;
        let contract_usdc_balance = app.wrap().query_balance(contract_addr.clone(), "usdc").unwrap().amount;
        
        assert_eq!(contract_kale_balance, Uint128::new(2000)); // 1000 for reserves + 1000 extra
        assert_eq!(contract_usdc_balance, Uint128::new(1000));
        
        // Now execute a swap: 100 KALE for USDC
        let swap_msg = ExecuteMsg::Swap {
            amount_in: 100,
            token_in: "kale".to_string(),
            token_out: "usdc".to_string(),
        };
        
        let swap_response = app.execute_contract(
            user.clone(),
            contract_addr.clone(),
            &swap_msg,
            &coins(100, "kale"),
        ).unwrap();
        
        // Verify swap was successful by checking events
        let wasm_event = swap_response.events.iter()
            .find(|e| e.ty == "wasm")
            .expect("No wasm event found");
        
        // Extract amount_out from the event attributes
        let amount_out_attr = wasm_event.attributes.iter()
            .find(|attr| attr.key == "amount_out")
            .expect("No amount_out attribute found");
        
        let amount_out = Uint128::from_str(&amount_out_attr.value)
            .expect("Failed to parse amount_out");
        
        // Extract fee from the event attributes
        let fee_attr = wasm_event.attributes.iter()
            .find(|attr| attr.key == "fee")
            .expect("No fee attribute found");
        
        let fee = Uint128::from_str(&fee_attr.value)
            .expect("Failed to parse fee");
        
        // Calculate expected fee: 0.2% of 100 KALE = 0.2 KALE
        let expected_fee = Uint128::new(100).multiply_ratio(2u128, 1000u128);
        assert_eq!(fee, expected_fee);
        
        // Calculate expected amount out using XYK formula
        // amount_out = (reserve_out * amount_in_after_fee) / (reserve_in + amount_in_after_fee)
        // where amount_in_after_fee = amount_in - fee = 100 - 0.2 = 99.8 KALE
        let amount_in_after_fee = Uint128::new(100).checked_sub(expected_fee).unwrap();
        let reserve_in = Uint128::new(1000); // Initial KALE reserve
        let reserve_out = Uint128::new(1000); // Initial USDC reserve
        
        let numerator = reserve_out.checked_mul(amount_in_after_fee).unwrap();
        let denominator = reserve_in.checked_add(amount_in_after_fee).unwrap();
        let expected_amount_out = numerator.checked_div(denominator).unwrap();
        
        assert_eq!(amount_out, expected_amount_out);
        
        // Check balances after swap
        // User should have received the USDC
        let user_usdc_balance = app.wrap().query_balance(user.clone(), "usdc").unwrap().amount;
        assert_eq!(user_usdc_balance, Uint128::new(1000) + amount_out);
        
        // User should have 100 less KALE
        let user_kale_balance = app.wrap().query_balance(user.clone(), "kale").unwrap().amount;
        assert_eq!(user_kale_balance, Uint128::new(900));
        
        // Check fee distribution
        // 50% of fee should go to yield address (0.1 KALE)
        let yield_fee = fee.multiply_ratio(50u128, 100u128);
        let yield_kale_balance = app.wrap().query_balance(yield_address.clone(), "kale").unwrap().amount;
        assert_eq!(yield_kale_balance, yield_fee);
        
        // 30% of fee should go to LP address (0.06 KALE)
        let lp_fee = fee.multiply_ratio(30u128, 100u128);
        let lp_kale_balance = app.wrap().query_balance(lp_address.clone(), "kale").unwrap().amount;
        assert_eq!(lp_kale_balance, lp_fee);
        
        // 20% of fee should go to Treasury address (0.04 KALE)
        let treasury_fee = fee.multiply_ratio(20u128, 100u128);
        let treasury_kale_balance = app.wrap().query_balance(treasury_address.clone(), "kale").unwrap().amount;
        assert_eq!(treasury_kale_balance, treasury_fee);
        
        // Verify the sum of distributed fees equals the total fee
        assert_eq!(yield_fee + lp_fee + treasury_fee, fee);
    }
}
