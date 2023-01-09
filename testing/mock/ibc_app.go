package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
)

// IBCApp contains IBC application module callbacks as defined in 05-port.
type IBCApp struct {
	PortID       string
	ScopedKeeper capabilitykeeper.ScopedKeeper

	OnChanOpenInit func(
		ctx sdk.Context,
		order channeltypes.Order,
		connectionHops []string,
		portID string,
		channelID string,
		channelCap *capabilitytypes.Capability,
		counterparty channeltypes.Counterparty,
		version string,
		middlewareData exported.MiddlewareData,
	) (string, error)

	OnChanOpenTry func(
		ctx sdk.Context,
		order channeltypes.Order,
		connectionHops []string,
		portID,
		channelID string,
		channelCap *capabilitytypes.Capability,
		counterparty channeltypes.Counterparty,
		counterpartyVersion string,
		middlewareData exported.MiddlewareData,
	) (version string, err error)

	OnChanOpenAck func(
		ctx sdk.Context,
		portID,
		channelID string,
		counterpartyChannelID string,
		counterpartyVersion string,
		middlewareData exported.MiddlewareData,
	) error

	OnChanOpenConfirm func(
		ctx sdk.Context,
		portID,
		channelID string,
		middlewareData exported.MiddlewareData,
	) error

	OnChanCloseInit func(
		ctx sdk.Context,
		portID,
		channelID string,
		middlewareData exported.MiddlewareData,
	) error

	OnChanCloseConfirm func(
		ctx sdk.Context,
		portID,
		channelID string,
		middlewareData exported.MiddlewareData,
	) error

	// OnRecvPacket must return an acknowledgement that implements the Acknowledgement interface.
	// In the case of an asynchronous acknowledgement, nil should be returned.
	// If the acknowledgement returned is successful, the state changes on callback are written,
	// otherwise the application state changes are discarded. In either case the packet is received
	// and the acknowledgement is written (in synchronous cases).
	OnRecvPacket func(
		ctx sdk.Context,
		packet channeltypes.Packet,
		relayer sdk.AccAddress,
		middlewareData exported.MiddlewareData,
	) exported.Acknowledgement

	OnAcknowledgementPacket func(
		ctx sdk.Context,
		packet channeltypes.Packet,
		acknowledgement []byte,
		relayer sdk.AccAddress,
		middlewareData exported.MiddlewareData,
	) error

	OnTimeoutPacket func(
		ctx sdk.Context,
		packet channeltypes.Packet,
		relayer sdk.AccAddress,
		middlewareData exported.MiddlewareData,
	) error
}

// NewIBCApp returns a IBCApp. An empty PortID indicates the mock app doesn't bind/claim ports.
func NewIBCApp(portID string, scopedKeeper capabilitykeeper.ScopedKeeper) *IBCApp {
	return &IBCApp{
		PortID:       portID,
		ScopedKeeper: scopedKeeper,
	}
}
