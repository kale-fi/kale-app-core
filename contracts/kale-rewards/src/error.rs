use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("Insufficient funds")]
    InsufficientFunds {},

    #[error("Stake is still locked")]
    StakeLocked {},

    #[error("Invalid token")]
    InvalidToken {},

    #[error("Invalid amount")]
    InvalidAmount {},

    #[error("No rewards to claim")]
    NoRewards {},

    #[error("Invalid reward rate")]
    InvalidRewardRate {},

    #[error("Invalid lock period")]
    InvalidLockPeriod {},
    
    #[error("Invalid funds")]
    InvalidFunds {},
    
    #[error("Invalid APY configuration")]
    InvalidAPY {},
}
