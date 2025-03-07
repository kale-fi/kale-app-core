package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"

	"kale-app-core/modules/kale-bank/types"
)

// Register error codes
var (
	ErrUnknownRequest = errors.Register("kale_bank", 100, "unknown request")
	ErrJSONUnmarshal  = errors.Register("kale_bank", 101, "json unmarshal error")
	ErrJSONMarshal    = errors.Register("kale_bank", 102, "json marshal error")
)

// Querier creates a new querier for kale-bank module
func NewQuerier(k KaleBankKeeper, legacyQuerierCdc *codec.LegacyAmino) func(ctx context.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	return func(ctx context.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryKaleBalance:
			return queryKaleBalance(ctx, req, k, legacyQuerierCdc)
		case types.QueryTotalSupply:
			return queryTotalSupply(ctx, req, k, legacyQuerierCdc)
		default:
			return nil, errors.Wrapf(ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}

func queryKaleBalance(ctx context.Context, req abci.RequestQuery, k KaleBankKeeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryKaleBalanceParams
	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, errors.Wrap(ErrJSONUnmarshal, err.Error())
	}

	addr, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, err
	}

	balance := k.GetKaleBalance(ctx, addr)
	res := types.QueryKaleBalanceResponse{
		Balance: balance,
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, res)
	if err != nil {
		return nil, errors.Wrap(ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryTotalSupply(ctx context.Context, req abci.RequestQuery, k KaleBankKeeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	totalSupply := k.GetTotalKaleSupply(ctx)
	res := types.QueryTotalSupplyResponse{
		TotalSupply: totalSupply,
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, res)
	if err != nil {
		return nil, errors.Wrap(ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
