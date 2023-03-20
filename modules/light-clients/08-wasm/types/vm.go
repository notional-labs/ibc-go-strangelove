package types

import (
	cosmwasm "github.com/CosmWasm/wasmvm"
	"github.com/CosmWasm/wasmvm/types"
	ics23 "github.com/confio/ics23/go"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	committypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

const GasMultiplier uint64 = 100
const maxGasLimit = uint64(0x7FFFFFFFFFFFFFFF)

var WasmVM *cosmwasm.VM

var _ exported.ClientState = (*ClientState)(nil)

type queryResponse struct {
	ProofSpecs      []*ics23.ProofSpec            `json:"proof_specs,omitempty"`
	Height          clienttypes.Height            `json:"height,omitempty"`
	GenesisMetadata []clienttypes.GenesisMetadata `json:"genesis_metadata,omitempty"`
	Result          contractResult                `json:"result,omitempty"`
	Root            committypes.MerkleRoot        `json:"root,omitempty"`
	Timestamp       uint64                        `json:"timestamp,omitempty"`
	Status          exported.Status               `json:"status,omitempty"`
}

type ClientCreateRequest struct {
	ClientCreateRequest ClientState `json:"client_create_request,omitempty"`
}

type ContractResult interface {
	Validate() bool
	Error() string
}

type contractResult struct {
	IsValid  bool   `json:"is_valid,omitempty"`
	ErrorMsg string `json:"err_msg,omitempty"`
	Data     []byte `json:"data,omitempty"`
}

func (r contractResult) Validate() bool {
	return r.IsValid
}

func (r contractResult) Error() string {
	return r.ErrorMsg
}

type clientStateCallResponse struct {
	Me                *ClientState    `json:"me,omitempty"`
	NewConsensusState *ConsensusState `json:"new_consensus_state,omitempty"`
	NewClientState    *ClientState    `json:"new_client_state,omitempty"`
	Result            contractResult  `json:"result,omitempty"`
}

func (r *clientStateCallResponse) resetImmutables(c *ClientState) {
	if r.Me != nil {
		r.Me.CodeId = c.CodeId
	}

	if r.NewConsensusState != nil {
		r.NewConsensusState.CodeId = c.CodeId
	}

	if r.NewClientState != nil {
		r.NewClientState.CodeId = c.CodeId
	}
}

func (r clientStateCallResponse) Validate() bool {
	return r.Result.Validate()
}

func (r clientStateCallResponse) Error() string {
	return r.Result.Error()
}

// Calls vm.Init with appropriate arguments
func initContract(codeID []byte, ctx sdk.Context, store sdk.KVStore) (*types.Response, error) {
	vmStore := NewStoreAdapter(store)
	gasMeter := ctx.GasMeter()
	chainID := ctx.BlockHeader().ChainID
	height := ctx.BlockHeader().Height
	// safety checks before casting below
	if height < 0 {
		panic("Block height must never be negative")
	}
	sec := ctx.BlockTime().Unix()
	if sec < 0 {
		panic("Block (unix) time must never be negative ")
	}
	env := types.Env{
		Block: types.BlockInfo{
			Height:  uint64(height),
			Time:    uint64(sec),
			ChainID: chainID,
		},
		Contract: types.ContractInfo{
			Address: "",
		},
	}

	msgInfo := types.MessageInfo{
		Sender: "",
		Funds:  nil,
	}
	// mockFailureAPI := *api.NewMockFailureAPI()
	// mockQuerier := api.MockQuerier{}

	desercost := types.UFraction{Numerator: 0, Denominator: 1}
	response, _, err := WasmVM.Instantiate(codeID, env, msgInfo, []byte("{}"), vmStore, cosmwasm.GoAPI{}, nil, gasMeter, gasMeter.Limit(), desercost)
	return response, err
}

// Calls vm.Execute with internally constructed Gas meter and environment
func callContract(codeID []byte, ctx sdk.Context, store sdk.KVStore, msg []byte) (*types.Response, error) {
	gasMeter := ctx.GasMeter()
	chainID := ctx.BlockHeader().ChainID
	height := ctx.BlockHeader().Height
	// safety checks before casting below
	if height < 0 {
		panic("Block height must never be negative")
	}
	sec := ctx.BlockTime().Unix()
	if sec < 0 {
		panic("Block (unix) time must never be negative ")
	}
	env := types.Env{
		Block: types.BlockInfo{
			Height:  uint64(height),
			Time:    uint64(sec),
			ChainID: chainID,
		},
		Contract: types.ContractInfo{
			Address: "",
		},
	}

	return callContractWithEnvAndMeter(codeID, ctx, store, env, gasMeter, msg)
}

// Calls vm.Execute with supplied environment and gas meter
func callContractWithEnvAndMeter(codeID cosmwasm.Checksum, ctx sdk.Context, store sdk.KVStore, env types.Env, gasMeter sdk.GasMeter, msg []byte) (*types.Response, error) {
	vmStore := NewStoreAdapter(store)
	msgInfo := types.MessageInfo{
		Sender: "",
		Funds:  nil,
	}
	// TODO: fix this
	// mockFailureAPI := *api.NewMockFailureAPI()
	// mockQuerier := api.MockQuerier{}
	desercost := types.UFraction{Numerator: 1, Denominator: 1}
	resp, gasUsed, err := WasmVM.Execute(codeID, env, msgInfo, msg, vmStore, cosmwasm.GoAPI{}, nil, gasMeter, gasMeter.Limit(), desercost)
	if &ctx != nil {
		consumeGas(ctx, gasUsed)
	}
	return resp, err
}

// TODO: Move this into a public method on the 28-wasm keeper
func queryContractWithStore(codeID cosmwasm.Checksum, store sdk.KVStore, msg []byte) ([]byte, error) {
	// TODO: fix this
	// mockEnv := api.MockEnv()
	// mockGasMeter := api.NewMockGasMeter(1)
	// mockFailureAPI := *api.NewMockFailureAPI()
	// mockQuerier := api.MockQuerier{}
	// TODO: figure out what this is for
	vmStore := NewStoreAdapter(store)
	desercost := types.UFraction{Numerator: 1, Denominator: 1}
	resp, _, err := WasmVM.Query(codeID, types.Env{}, msg, vmStore, cosmwasm.GoAPI{}, nil, nil, maxGasLimit, desercost)
	return resp, err
}

func consumeGas(ctx sdk.Context, gas uint64) {
	consumed := gas / GasMultiplier
	ctx.GasMeter().ConsumeGas(consumed, "wasm contract")
	// throw OutOfGas error if we ran out (got exactly to zero due to better limit enforcing)
	if ctx.GasMeter().IsOutOfGas() {
		panic(sdk.ErrorOutOfGas{Descriptor: "Wasmer function execution"})
	}
}
