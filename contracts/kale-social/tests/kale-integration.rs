#[cfg(test)]
mod tests {
    use cosmwasm_std::{
        Addr, Empty,
    };
    use cw_multi_test::{App, Contract, ContractWrapper, Executor};
    
    use kale_social::contract::{execute, instantiate, query};
    use kale_social::msg::InstantiateMsg;

    // Helper function to create the contract
    fn social_contract() -> Box<dyn Contract<Empty>> {
        let contract = ContractWrapper::new(execute, instantiate, query);
        Box::new(contract)
    }

    #[test]
    fn instantiate_contract() {
        // Create a new app with initial funds
        let mut app = App::default();
        
        // Set up test accounts
        let owner = Addr::unchecked("owner");
        let treasury = Addr::unchecked("treasury");
        
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
        
        // Verify contract was instantiated successfully
        assert!(contract_addr.as_str().len() > 0);
    }
}
