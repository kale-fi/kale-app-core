#[cfg(test)]
mod tests {
    use cosmwasm_std::{Addr, Uint128, coins};
    use cw_multi_test::{App, ContractWrapper, Executor};
    
    use kale_amm::{execute, instantiate, query, ExecuteMsg, InstantiateMsg, QueryMsg};

    #[test]
    fn swap_works() {
        // Create addresses for testing
        let owner = Addr::unchecked("owner");
        let user1 = Addr::unchecked("user1");
        let yield_address = Addr::unchecked("kale1yield");
        let lp_address = Addr::unchecked("kale1lp");
        let treasury_address = Addr::unchecked("kale1treasury");
        
        // Create a completely default App
        let mut app = App::default();
        
        // First check no balances exist
        let owner_initial = app.wrap().query_all_balances(&owner).unwrap_or_default();
        println!("Owner initial balance (should be empty): {:?}", owner_initial);
        
        // Explicitly mint coins to accounts (this is more reliable than init_balance)
        app.sudo(cw_multi_test::SudoMsg::Bank(
            cw_multi_test::BankSudo::Mint {
                to_address: owner.to_string(),
                amount: coins(10_000, "ukale"),
            }
        )).unwrap();
        
        app.sudo(cw_multi_test::SudoMsg::Bank(
            cw_multi_test::BankSudo::Mint {
                to_address: owner.to_string(),
                amount: coins(10_000, "uusdc"),
            }
        )).unwrap();
        
        app.sudo(cw_multi_test::SudoMsg::Bank(
            cw_multi_test::BankSudo::Mint {
                to_address: user1.to_string(),
                amount: coins(1100, "ukale"),
            }
        )).unwrap();
        
        app.sudo(cw_multi_test::SudoMsg::Bank(
            cw_multi_test::BankSudo::Mint {
                to_address: user1.to_string(),
                amount: coins(1000, "uusdc"),
            }
        )).unwrap();
        
        // Verify owner balances before contract creation - this is crucial
        let owner_balance = app.wrap().query_all_balances(&owner).unwrap();
        println!("Owner balances after minting: {:?}", owner_balance);
        
        // Store contract code
        let contract = ContractWrapper::new(execute, instantiate, query);
        let code_id = app.store_code(Box::new(contract));
        
        // Set a fee threshold of 5 tokens - fees will accumulate until they reach this amount
        let fee_threshold = Uint128::from(5u128);
        
        // Instantiate contract with initial reserves
        let contract_addr = app.instantiate_contract(
            code_id,
            owner.clone(),
            &InstantiateMsg {
                owner: owner.to_string(),
                fee_percent: 2, // 0.2% fee
                yield_percent: 50, // 50% to yield
                lp_percent: 25,    // 25% to LP
                treasury_percent: 25, // 25% to treasury
                token_a: "ukale".to_string(),
                token_b: "uusdc".to_string(),
                reserves_a: Uint128::new(1000), // Start with 1000 ukale
                reserves_b: Uint128::new(1000), // Start with 1000 uusdc
                fee_threshold: fee_threshold,
            },
            &coins(1000, "ukale"), // Send 1000 ukale for reserves
            "kale-amm",
            None,
        ).unwrap();
        
        // Send additional uusdc to the contract for initial liquidity
        app.send_tokens(
            owner.clone(),
            contract_addr.clone(),
            &coins(1000, "uusdc"),
        ).unwrap();
        
        // Verify contract balances after initialization
        let contract_balance = app.wrap().query_all_balances(&contract_addr).unwrap();
        println!("Contract balances after initialization: {:?}", contract_balance);
        
        // Verify user balances before swap
        let user_balance = app.wrap().query_all_balances(&user1).unwrap();
        println!("User balances before swap: {:?}", user_balance);

        // Execute first swap - This will accumulate fees but not distribute them yet
        let swap_amount = Uint128::from(100u128);
        let swap_response = app.execute_contract(
            user1.clone(),
            contract_addr.clone(),
            &ExecuteMsg::KaleSwap {
                amount: swap_amount,
                token_in: "ukale".to_string(),
                token_out: "uusdc".to_string(),
            },
            &coins(100, "ukale"),
        ).unwrap();

        // Print events for inspection
        println!("\n--- DEBUG: First Swap Events ---");
        for event in &swap_response.events {
            println!("Event Type: {}", event.ty);
            for attr in &event.attributes {
                println!("  {} = {}", attr.key, attr.value);
            }
        }
        
        // Extract swap results from events
        let events = swap_response.events;
        let wasm_event = events.iter().find(|e| e.ty == "wasm").unwrap();
        
        // Find relevant attributes
        let amount_out_attr = wasm_event.attributes.iter()
            .find(|a| a.key == "amount_out")
            .expect("amount_out attribute not found");
        let amount_out = amount_out_attr.value.parse::<u128>().unwrap();
        
        let fee_attr = wasm_event.attributes.iter()
            .find(|a| a.key == "fee")
            .expect("fee attribute not found");
        let fee = fee_attr.value.parse::<u128>().unwrap();
        
        let accumulated_fee_attr = wasm_event.attributes.iter()
            .find(|a| a.key == "accumulated_fee")
            .expect("accumulated_fee attribute not found");
        let accumulated_fee = accumulated_fee_attr.value.parse::<u128>().unwrap();
        
        let fees_distributed_attr = wasm_event.attributes.iter()
            .find(|a| a.key == "fees_distributed")
            .expect("fees_distributed attribute not found");
        let fees_distributed = fees_distributed_attr.value == "true";
        
        println!("First swap result: amount_out={}, fee={}, accumulated_fee={}, fees_distributed={}", 
                 amount_out, fee, accumulated_fee, fees_distributed);
        
        // Check balances after first swap
        let user_balance_after = app.wrap().query_all_balances(&user1).unwrap();
        println!("User balances after first swap: {:?}", user_balance_after);
        
        let yield_balance = app.wrap().query_all_balances(&yield_address).unwrap();
        let lp_balance = app.wrap().query_all_balances(&lp_address).unwrap();
        let treasury_balance = app.wrap().query_all_balances(&treasury_address).unwrap();
        
        println!("Yield balance after first swap: {:?}", yield_balance);
        println!("LP balance after first swap: {:?}", lp_balance);
        println!("Treasury balance after first swap: {:?}", treasury_balance);
        
        // Verify the output amount is approximately 89 uusdc
        assert!(amount_out >= 89 && amount_out <= 90, "Expected ~89, got {}", amount_out);
        
        // Verify fee is 0.2% of 100 = 0.2 ukale (which is 2 in micro units)
        assert_eq!(fee, 2, "Expected fee 0.2, got {}", fee);
        
        // Verify fees are being accumulated but not distributed yet
        assert_eq!(accumulated_fee, 2, "Expected accumulated fee 2, got {}", accumulated_fee);
        assert!(!fees_distributed, "Fees should not be distributed yet");
        assert!(yield_balance.is_empty(), "Yield should not have received fees yet");
        assert!(lp_balance.is_empty(), "LP should not have received fees yet");
        assert!(treasury_balance.is_empty(), "Treasury should not have received fees yet");
        
        // Execute more swaps to accumulate fees up to the threshold
        for i in 0..2 {
            // Mint more tokens for user1
            app.sudo(cw_multi_test::SudoMsg::Bank(
                cw_multi_test::BankSudo::Mint {
                    to_address: user1.to_string(),
                    amount: coins(100, "ukale"),
                }
            )).unwrap();
            
            // Execute swap
            app.execute_contract(
                user1.clone(),
                contract_addr.clone(),
                &ExecuteMsg::KaleSwap {
                    amount: swap_amount,
                    token_in: "ukale".to_string(),
                    token_out: "uusdc".to_string(),
                },
                &coins(100, "ukale"),
            ).unwrap();
            
            println!("Executed swap {}", i + 2);
        }
        
        // Query accumulated fees
        let accumulated_fees: Uint128 = app.wrap()
            .query_wasm_smart(
                contract_addr.clone(),
                &QueryMsg::GetAccumulatedFees {
                    denom: "ukale".to_string(),
                },
            )
            .unwrap();
        
        println!("Accumulated fees after multiple swaps: {}", accumulated_fees);
        
        // Execute one more swap to trigger fee distribution (total of 4 swaps * 2 fee per swap = 8 > threshold of 5)
        app.sudo(cw_multi_test::SudoMsg::Bank(
            cw_multi_test::BankSudo::Mint {
                to_address: user1.to_string(),
                amount: coins(100, "ukale"),
            }
        )).unwrap();
        
        let final_swap_response = app.execute_contract(
            user1.clone(),
            contract_addr.clone(),
            &ExecuteMsg::KaleSwap {
                amount: swap_amount,
                token_in: "ukale".to_string(),
                token_out: "uusdc".to_string(),
            },
            &coins(100, "ukale"),
        ).unwrap();
        
        // Print events for inspection
        println!("\n--- DEBUG: Final Swap Events ---");
        for event in &final_swap_response.events {
            println!("Event Type: {}", event.ty);
            for attr in &event.attributes {
                println!("  {} = {}", attr.key, attr.value);
            }
        }
        
        // Check balances after final swap that should trigger fee distribution
        let yield_balance_after = app.wrap().query_all_balances(&yield_address).unwrap();
        let lp_balance_after = app.wrap().query_all_balances(&lp_address).unwrap();
        let treasury_balance_after = app.wrap().query_all_balances(&treasury_address).unwrap();
        
        println!("Yield balance after final swap: {:?}", yield_balance_after);
        println!("LP balance after final swap: {:?}", lp_balance_after);
        println!("Treasury balance after final swap: {:?}", treasury_balance_after);
        
        // Verify fee distribution (50% yield, 25% LP, 25% treasury)
        // With accumulated fees of 8, the distribution should be:
        // - Yield: 4 ukale (50%)
        // - LP: 2 ukale (25%) 
        // - Treasury: 2 ukale (25%)
        
        let yield_ukale = yield_balance_after.iter().find(|c| c.denom == "ukale").map(|c| c.amount.u128()).unwrap_or(0);
        let lp_ukale = lp_balance_after.iter().find(|c| c.denom == "ukale").map(|c| c.amount.u128()).unwrap_or(0);
        let treasury_ukale = treasury_balance_after.iter().find(|c| c.denom == "ukale").map(|c| c.amount.u128()).unwrap_or(0);
        
        // Check that fees are distributed correctly
        assert_eq!(yield_ukale, 4, "Yield should get exactly 4 ukale (50%), got {}", yield_ukale);
        assert_eq!(lp_ukale, 2, "LP should get exactly 2 ukale (25%), got {}", lp_ukale);
        assert_eq!(treasury_ukale, 2, "Treasury should get exactly 2 ukale (25%), got {}", treasury_ukale);
        
        // Check that all accumulated fees were distributed
        assert_eq!(yield_ukale + lp_ukale + treasury_ukale, 8, 
                   "Total distributed should be 8, got {}", yield_ukale + lp_ukale + treasury_ukale);
    }
}