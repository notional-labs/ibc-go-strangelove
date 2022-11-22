package wasm_test

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	wasm "github.com/cosmos/ibc-go/v5/modules/light-clients/10-wasm"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	"github.com/cosmos/ibc-go/v5/testing/simapp"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type WasmTestSuite struct {
	suite.Suite
	coordinator *ibctesting.Coordinator
	chainA      *ibctesting.TestChain
	ctx         sdk.Context
	cdc         codec.Codec
	now         time.Time
	store       sdk.KVStore
}

var (
	RepositoryInitial = "go_side_added"
	RepositoryFinal = RepositoryInitial + "_grandpa_contract_added"
)

func (suite *WasmTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	// commit some blocks so that QueryProof returns valid proof (cannot return valid query if height <= 1)
	suite.coordinator.CommitNBlocks(suite.chainA, 2)

	// TODO: deprecate usage in favor of testing package
	checkTx := false
	app := simapp.Setup(checkTx)
	suite.cdc = app.AppCodec()
	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, Time: suite.now})

	wasmConfig := wasm.VMConfig{
		DataDir:           "tmp",
		SupportedFeatures: []string{"storage", "iterator"},
		MemoryLimitMb:     uint32(math.Pow(2, 12)),
		PrintDebug:        true,
		CacheSizeMb:       uint32(math.Pow(2, 8)),
	}
	validationConfig := wasm.ValidationConfig{
		MaxSizeAllowed: int(math.Pow(2, 26)),
	}
	suite.store = suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), exported.Wasm)
	os.MkdirAll("tmp", 0o755)
	wasm.CreateVM(&wasmConfig, &validationConfig)
	data, err := os.ReadFile("ics10_grandpa_cw.wasm")
	suite.Require().NoError(err)
	
	// Currently pushing to client-specific store, this will change to a keeper, but okay for now/testing (single wasm client)
	codeId, err := wasm.PushNewWasmCode(suite.store, data)
	suite.Require().NoError(err)

	clientState := wasm.ClientState{
		Data: []byte("ClientStateData"),
		CodeId: codeId,
		LatestHeight: types.NewHeight(1,1),
		ProofSpecs: nil,
		Repository: RepositoryInitial,
	}

	consensusState := wasm.ConsensusState{
		Data: []byte("ConsensusStateData"),
		CodeId: codeId,
		Timestamp: 1,
	}

	clientTestItem := []byte("client_test_item")
	originalNumber := []byte("12783490")
	finalNumber := []byte("12783491")
	suite.store.Set(clientTestItem, originalNumber)

	// Initialize only instantiates the contract
	err = clientState.Initialize(suite.ctx, suite.cdc, suite.store, &consensusState)
	suite.Require().NoError(err)

	// "client_test_item" is initialized in contract instantiation (wasm-side) and read here (go-side)
	value := suite.store.Get(clientTestItem)
	suite.Require().NotNil(value)
	suite.Require().Equal(finalNumber, value)
	fmt.Println("Value: ", string(value[:]))

	// Replicate the 02-client CreateClient code
	suite.store.Set(host.ClientStateKey(), types.MustMarshalClientState(suite.cdc, &clientState))
	suite.store.Set(host.ConsensusStateKey(clientState.GetLatestHeight()), types.MustMarshalConsensusState(suite.cdc, &consensusState))

}

func (suite *WasmTestSuite) TestWasmExecute() {
	clientState, err := types.UnmarshalClientState(suite.cdc, suite.store.Get(host.ClientStateKey()))
	suite.Require().NoError(err)

	err = clientState.(*wasm.ClientState).TestSharedKVStore(suite.ctx, suite.store)
	suite.Require().NoError(err)

	clientState2, err := types.UnmarshalClientState(suite.cdc, suite.store.Get(host.ClientStateKey()))
	suite.Require().NoError(err)
	suite.Require().Equal(RepositoryFinal, clientState2.(*wasm.ClientState).Repository)
}

func TestWasmTestSuite(t *testing.T) {
	suite.Run(t, new(WasmTestSuite))
}
