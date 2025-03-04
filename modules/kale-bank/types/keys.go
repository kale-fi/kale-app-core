package types

const (
	// ModuleName defines the module name
	ModuleName = "kalebank"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// KaleDenom defines the native token denomination for KALE
	KaleDenom = "ukale"

	// TotalSupply defines the total supply of KALE tokens (100 million)
	TotalSupply = 100_000_000_000_000 // 100M with 6 decimal places
)

// Key prefixes
var (
	// InitializedKey is the key to track if the module has been initialized
	InitializedKey = []byte("initialized")
)
