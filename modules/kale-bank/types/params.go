package types

import (
	"fmt"
)

// Module parameter keys
const (
	// EnableMinting parameter key
	EnableMintingKey = "EnableMinting"

	// MintingCap parameter key
	MintingCapKey = "MintingCap"
)

// Key prefixes for the KV store
var (
	// ParamsKey is the key used to store module parameters
	ParamsKey = []byte("Params")
)

// Params defines the parameters for the kale-bank module
type Params struct {
	EnableMinting bool   `protobuf:"varint,1,opt,name=enable_minting,json=enableMinting,proto3" json:"enable_minting,omitempty" yaml:"enable_minting"`
	MintingCap    string `protobuf:"bytes,2,opt,name=minting_cap,json=mintingCap,proto3" json:"minting_cap,omitempty" yaml:"minting_cap"`
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		EnableMinting: true,
		MintingCap:    "1000000000",
	}
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.MintingCap == "" {
		return fmt.Errorf("minting cap cannot be empty")
	}
	return nil
}

// String implements the Stringer interface
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  EnableMinting: %t
  MintingCap:    %s`,
		p.EnableMinting, p.MintingCap)
}

// ProtoMessage implements the proto.Message interface
func (*Params) ProtoMessage() {}

// Reset implements the proto.Message interface
func (p *Params) Reset() {
	*p = Params{}
}
