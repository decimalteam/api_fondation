package types

import "time"

const (
	RequestTimeout    = 16 * time.Second
	RequestRetryDelay = 32 * time.Millisecond

	Bech32Prefix        = "d0"
	Bech32PrefixAccAddr = Bech32Prefix

	EvmtypesModuleName        = "evm"
	CointypesModuleName       = "coin"
	NfttypesReservedPool      = "reserved_pool"
	LegacytypesLegacyCoinPool = "legacy_coin_pool"
	SwaptypesSwapPool         = "atomic_swap_pool"
	ValidatortypesModuleName  = "validator"
	FeetypesBurningPool       = "burning_pool"
)
