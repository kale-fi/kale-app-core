package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "kalefi"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// TradeEventPrefix is the prefix for trade event keys
	TradeEventPrefix = "trade_event"

	// TradeEventCounterKey is the key for the trade event counter
	TradeEventCounterKey = "trade_event_counter"

	// EventTypeTradeCreated is the event type for trade creation
	EventTypeTradeCreated = "trade_created"

	// EventTypeTradeExecuted is the event type for trade execution
	EventTypeTradeExecuted = "trade_executed"

	// EventTypeTradeCancelled is the event type for trade cancellation
	EventTypeTradeCancelled = "trade_cancelled"

	// AttributeKeyTradeId is the attribute key for trade ID
	AttributeKeyTradeId = "trade_id"

	// AttributeKeyTrader is the attribute key for trader address
	AttributeKeyTrader = "trader"
)

// Key prefixes for store keys
var (
	// KeyPrefixTradeEvent is the prefix for storing trade events
	KeyPrefixTradeEvent = []byte{0x01}

	// KeyPrefixTraderEvents is the prefix for storing trader events
	KeyPrefixTraderEvents = []byte{0x02}

	// KeyTradeEventCounter is the key for the trade event counter
	KeyTradeEventCounter = []byte{0x03}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x04}

	// TradeEventKeyPrefix is the prefix for storing trade events (alias for KeyPrefixTradeEvent)
	TradeEventKeyPrefix = KeyPrefixTradeEvent

	// TraderIndexPrefix is the prefix for indexing trades by trader (alias for KeyPrefixTraderEvents)
	TraderIndexPrefix = KeyPrefixTraderEvents
)

// GetTradeEventKey returns the store key for a trade event
func GetTradeEventKey(id uint64) []byte {
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, id)
	return append(KeyPrefixTradeEvent, idBytes...)
}

// GetTradeByTraderPrefix returns the prefix for a trader's trades
func GetTradeByTraderPrefix(traderAddr sdk.AccAddress) []byte {
	return append(KeyPrefixTraderEvents, traderAddr...)
}

// GetTradeByTraderIndexKey returns the key for a specific trade by a trader
func GetTradeByTraderIndexKey(traderAddr sdk.AccAddress, tradeID uint64) []byte {
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, tradeID)
	return append(GetTradeByTraderPrefix(traderAddr), idBytes...)
}

// ParseTradeByTraderIndexKey parses a trade-by-trader index key
func ParseTradeByTraderIndexKey(key []byte) (sdk.AccAddress, uint64) {
	// Format: prefix + traderAddr + tradeID
	prefixLen := len(KeyPrefixTraderEvents)
	addrLen := 20 // standard address length

	// Extract trader address
	traderAddr := key[prefixLen : prefixLen+addrLen]

	// Extract trade ID (last 8 bytes)
	tradeIDBytes := key[prefixLen+addrLen:]
	tradeID := binary.BigEndian.Uint64(tradeIDBytes)

	return traderAddr, tradeID
}
