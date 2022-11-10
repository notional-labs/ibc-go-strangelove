package keeper

import (
	"bytes"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v5/modules/core/03-connection/types"
	"github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
)

// validateChanOpenInit performs the appropriate validation logic for ChanOpenInit.
func (k Keeper) validateChanOpenInit(ctx sdk.Context, connectionID string, order types.Order) error {
	if connectionID != connectiontypes.LocalhostID {
		// connection hop length checked on msg.ValidateBasic()
		connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionID)
		if !found {
			return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, connectionID)
		}

		getVersions := connectionEnd.GetVersions()
		if len(getVersions) != 1 {
			return sdkerrors.Wrapf(
				connectiontypes.ErrInvalidVersion,
				"single version must be negotiated on connection before opening channel, got: %v",
				getVersions,
			)
		}

		if !connectiontypes.VerifySupportedFeature(getVersions[0], order.String()) {
			return sdkerrors.Wrapf(
				connectiontypes.ErrInvalidVersion,
				"connection version %s does not support channel ordering: %s",
				getVersions[0], order.String(),
			)
		}
	}

	return nil
}

// verifyChannelState performs the appropriate channel state verification logic for either a localhost connection
// or a counterparty chain.
func (k Keeper) verifyChannelState(
	ctx sdk.Context,
	expectedChannel types.Channel,
	portID string,
	channelID string,
	proofHeight exported.Height,
	proof []byte,
) error {
	connectionID := expectedChannel.ConnectionHops[0]

	if connectionID == connectiontypes.LocalhostID {
		// get the counterparty channel directly from this chain's channelKeeper store
		actualChannel, ok := k.GetChannel(ctx, portID, channelID)

		// check that the counterparty channel actually exists in this chain's channelKeeper store and is in the expected state
		if !ok {
			return sdkerrors.Wrap(types.ErrChannelNotFound, "failed localhost channel state verification, counterparty channel does not exist")
		}

		if actualChannel.State != expectedChannel.State {
			return sdkerrors.Wrapf(types.ErrInvalidChannelState, "failed localhost channel state verification, channel state is not %s (got %s)", expectedChannel.State, actualChannel.State)
		}
	} else {
		connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionID)
		if !found {
			return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, connectionID)
		}

		if connectionEnd.GetState() != int32(connectiontypes.OPEN) {
			return sdkerrors.Wrapf(
				connectiontypes.ErrInvalidConnectionState,
				"connection state is not OPEN (got %s)", connectiontypes.State(connectionEnd.GetState()).String(),
			)
		}

		// connection version & feature verification only need to happen on ChanOpenTry
		if expectedChannel.State == types.INIT {
			getVersions := connectionEnd.GetVersions()
			if len(getVersions) != 1 {
				return sdkerrors.Wrapf(
					connectiontypes.ErrInvalidVersion,
					"single version must be negotiated on connection before opening channel, got: %v",
					getVersions,
				)
			}

			if !connectiontypes.VerifySupportedFeature(getVersions[0], expectedChannel.Ordering.String()) {
				return sdkerrors.Wrapf(
					connectiontypes.ErrInvalidVersion,
					"connection version %s does not support channel ordering: %s", getVersions[0], expectedChannel.Ordering.String(),
				)
			}
		}

		counterpartyHops := []string{connectionEnd.GetCounterparty().GetConnectionID()}

		expectedChannel.ConnectionHops = counterpartyHops

		if err := k.connectionKeeper.VerifyChannelState(
			ctx, connectionEnd, proofHeight, proof,
			portID, channelID, expectedChannel,
		); err != nil {
			return err
		}
	}

	return nil
}

// validatePacketSend performs the appropriate validation logic for either a packet being sent on a localhost
// connection or a packet being sent to a counterparty chain.
func (k Keeper) validatePacketSend(ctx sdk.Context, connectionID string, packet exported.PacketI) error {
	if connectionID == connectiontypes.LocalhostID {
		// check if packet is timed out
		latestHeight := uint64(ctx.BlockHeight())
		timeoutHeight := packet.GetTimeoutHeight()
		if !timeoutHeight.IsZero() && latestHeight >= timeoutHeight.GetRevisionHeight() {
			return sdkerrors.Wrapf(
				types.ErrPacketTimeout,
				"receiving chain block height >= packet timeout height (%d >= %s)", latestHeight, timeoutHeight,
			)
		}

		latestTimestamp := uint64(ctx.BlockTime().UnixNano())
		if packet.GetTimeoutTimestamp() != 0 && latestTimestamp >= packet.GetTimeoutTimestamp() {
			return sdkerrors.Wrapf(
				types.ErrPacketTimeout,
				"receiving chain block timestamp >= packet timeout timestamp (%s >= %s)", time.Unix(0, int64(latestTimestamp)), time.Unix(0, int64(packet.GetTimeoutTimestamp())),
			)
		}
	} else {
		connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionID)
		if !found {
			return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, connectionID)
		}

		clientState, found := k.clientKeeper.GetClientState(ctx, connectionEnd.GetClientID())
		if !found {
			return clienttypes.ErrConsensusStateNotFound
		}

		// prevent accidental sends with clients that cannot be updated
		clientStore := k.clientKeeper.ClientStore(ctx, connectionEnd.GetClientID())
		if status := clientState.Status(ctx, clientStore, k.cdc); status != exported.Active {
			return sdkerrors.Wrapf(clienttypes.ErrClientNotActive, "cannot send packet using client (%s) with status %s", connectionEnd.GetClientID(), status)
		}

		// check if packet is timed out on the receiving chain
		latestHeight := clientState.GetLatestHeight()
		timeoutHeight := packet.GetTimeoutHeight()
		if !timeoutHeight.IsZero() && latestHeight.GTE(timeoutHeight) {
			return sdkerrors.Wrapf(
				types.ErrPacketTimeout,
				"receiving chain block height >= packet timeout height (%s >= %s)", latestHeight, timeoutHeight,
			)
		}

		latestTimestamp, err := k.connectionKeeper.GetTimestampAtHeight(ctx, connectionEnd, latestHeight)
		if err != nil {
			return err
		}

		if packet.GetTimeoutTimestamp() != 0 && latestTimestamp >= packet.GetTimeoutTimestamp() {
			return sdkerrors.Wrapf(
				types.ErrPacketTimeout,
				"receiving chain block timestamp >= packet timeout timestamp (%s >= %s)", time.Unix(0, int64(latestTimestamp)), time.Unix(0, int64(packet.GetTimeoutTimestamp())),
			)
		}
	}

	return nil
}

// verifyPacketCommitment performs the appropriate packet verification logic for either a packet commitment on a localhost
// connection or a packet commitment on a counterparty chain.
func (k Keeper) verifyPacketCommitment(
	ctx sdk.Context,
	connectionID string,
	packet exported.PacketI,
	expectedState []byte,
	proof []byte,
	proofHeight exported.Height,
) error {
	if connectionID == connectiontypes.LocalhostID {
		// get the desired packet commitment hash directly from this chain's channelKeeper
		actualState := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

		// check that the packet commitment actually exists in this chain's channelKeeper store and matches the expected state
		if len(actualState) == 0 {
			return sdkerrors.Wrap(types.ErrPacketCommitmentNotFound, "couldn't verify localhost packet commitment, commitment does not exist")
		}

		if !bytes.Equal(actualState, expectedState) {
			return sdkerrors.Wrapf(types.ErrInvalidPacket, "couldn't verify localhost packet commitment (%s ≠ %s)", actualState, expectedState)
		}

	} else {
		// Connection must be OPEN to receive a packet. It is possible for connection to not yet be open if packet was
		// sent optimistically before connection and channel handshake completed. However, to receive a packet,
		// connection and channel must both be open
		connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionID)
		if !found {
			return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, connectionID)
		}

		if connectionEnd.GetState() != int32(connectiontypes.OPEN) {
			return sdkerrors.Wrapf(
				connectiontypes.ErrInvalidConnectionState,
				"connection state is not OPEN (got %s)", connectiontypes.State(connectionEnd.GetState()).String(),
			)
		}

		// verify that the counterparty did commit to sending this packet
		if err := k.connectionKeeper.VerifyPacketCommitment(
			ctx, connectionEnd, proofHeight, proof,
			packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(),
			expectedState,
		); err != nil {
			return sdkerrors.Wrap(err, "couldn't verify counterparty packet commitment")
		}
	}

	return nil
}

// verifyPacketAcknowledgement performs the appropriate verification logic for either an acknowledgement on a localhost
// connection or an acknowledgement on a counterparty chain.
func (k Keeper) verifyPacketAcknowledgement(
	ctx sdk.Context,
	connectionID string,
	packet exported.PacketI,
	expectedState []byte,
	proof []byte,
	proofHeight exported.Height,
) error {
	if connectionID == connectiontypes.LocalhostID {
		// get the desired ack hash directly from this chain's channelKeeper
		storedAck, ok := k.GetPacketAcknowledgement(ctx, packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())

		// check that the packet acknowledgement actually exists in this chain's channelKeeper store and matches the expected state
		if !ok {
			return sdkerrors.Wrap(types.ErrInvalidAcknowledgement, "couldn't verify localhost acknowledgement, acknowledgement does not exist")
		}

		expectedAck := types.CommitAcknowledgement(expectedState)
		if !bytes.Equal(expectedAck, storedAck) {
			return sdkerrors.Wrapf(types.ErrInvalidAcknowledgement, "couldn't verify localhost acknowledgement (%s ≠ %s)", expectedAck, storedAck)
		}
	} else {
		connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionID)
		if !found {
			return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, connectionID)
		}

		if connectionEnd.GetState() != int32(connectiontypes.OPEN) {
			return sdkerrors.Wrapf(
				connectiontypes.ErrInvalidConnectionState,
				"connection state is not OPEN (got %s)", connectiontypes.State(connectionEnd.GetState()).String(),
			)
		}

		if err := k.connectionKeeper.VerifyPacketAcknowledgement(
			ctx, connectionEnd, proofHeight, proof, packet.GetDestPort(), packet.GetDestChannel(),
			packet.GetSequence(), expectedState,
		); err != nil {
			return err
		}
	}

	return nil
}

// verifyTimeoutPacket performs the appropriate verification logic for either a timeout on a localhost connection or
// a timeout on a counterparty chain.
func (k Keeper) verifyTimeoutPacket(
	ctx sdk.Context,
	packet exported.PacketI,
	channel types.Channel,
	expectedNextSeqRecv uint64,
	proofHeight exported.Height,
	proof []byte,
) error {
	if channel.ConnectionHops[0] == connectiontypes.LocalhostID {
		// check that timeout height or timeout timestamp has passed on the other end
		timeoutHeight := packet.GetTimeoutHeight()
		if (timeoutHeight.IsZero() || uint64(ctx.BlockHeight()) < timeoutHeight.GetRevisionHeight()) &&
			(packet.GetTimeoutTimestamp() == 0 || uint64(ctx.BlockTime().UnixNano()) < packet.GetTimeoutTimestamp()) {
			return sdkerrors.Wrap(types.ErrPacketTimeout, "packet timeout has not been reached for height or timestamp")
		}

		commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

		if len(commitment) == 0 {
			EmitTimeoutPacketEvent(ctx, packet, channel)
			// This error indicates that the timeout has already been relayed
			// or there is a misconfigured relayer attempting to prove a timeout
			// for a packet never sent. Core IBC will treat this error as a no-op in order to
			// prevent an entire relay transaction from failing and consuming unnecessary fees.
			return types.ErrNoOpMsg
		}

		if channel.State != types.OPEN {
			return sdkerrors.Wrapf(
				types.ErrInvalidChannelState,
				"channel state is not OPEN (got %s)", channel.State.String(),
			)
		}

		packetCommitment := types.CommitPacket(k.cdc, packet)

		// verify we sent the packet and haven't cleared it out yet
		if !bytes.Equal(commitment, packetCommitment) {
			return sdkerrors.Wrapf(types.ErrInvalidPacket, "packet commitment bytes are not equal: got (%v), expected (%v)", commitment, packetCommitment)
		}

		switch channel.Ordering {
		case types.ORDERED:
			// check that packet has not been received
			if expectedNextSeqRecv > packet.GetSequence() {
				return sdkerrors.Wrapf(
					types.ErrPacketReceived,
					"packet already received, next sequence receive > packet sequence (%d > %d)", expectedNextSeqRecv, packet.GetSequence(),
				)
			}

			actualNextSeqRecv, ok := k.GetNextSequenceRecv(ctx, packet.GetDestPort(), packet.GetDestChannel())
			if !ok {
				return sdkerrors.Wrap(types.ErrSequenceReceiveNotFound, "failed localhost next sequence receive verification, next sequence receive does not exist in store")
			}

			// check that the recv sequence is as claimed
			if actualNextSeqRecv != expectedNextSeqRecv {
				return sdkerrors.Wrapf(types.ErrPacketSequenceOutOfOrder, "failed localhost next sequence receive verification, (%d ≠ %d)", actualNextSeqRecv, expectedNextSeqRecv)
			}
		case types.UNORDERED:
			_, ok := k.GetPacketReceipt(ctx, packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			if ok {
				return sdkerrors.Wrap(types.ErrAcknowledgementExists, "failed localhost packet receipt absence verification")
			}
		default:
			panic(sdkerrors.Wrapf(types.ErrInvalidChannelOrdering, channel.Ordering.String()))
		}

		return nil
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(
			connectiontypes.ErrConnectionNotFound,
			channel.ConnectionHops[0],
		)
	}

	// check that timeout height or timeout timestamp has passed on the other end
	proofTimestamp, err := k.connectionKeeper.GetTimestampAtHeight(ctx, connectionEnd, proofHeight)
	if err != nil {
		return err
	}

	timeoutHeight := packet.GetTimeoutHeight()
	if (timeoutHeight.IsZero() || proofHeight.LT(timeoutHeight)) &&
		(packet.GetTimeoutTimestamp() == 0 || proofTimestamp < packet.GetTimeoutTimestamp()) {
		return sdkerrors.Wrap(types.ErrPacketTimeout, "packet timeout has not been reached for height or timestamp")
	}

	commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

	if len(commitment) == 0 {
		EmitTimeoutPacketEvent(ctx, packet, channel)
		// This error indicates that the timeout has already been relayed
		// or there is a misconfigured relayer attempting to prove a timeout
		// for a packet never sent. Core IBC will treat this error as a no-op in order to
		// prevent an entire relay transaction from failing and consuming unnecessary fees.
		return types.ErrNoOpMsg
	}

	if channel.State != types.OPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not OPEN (got %s)", channel.State.String(),
		)
	}

	packetCommitment := types.CommitPacket(k.cdc, packet)

	// verify we sent the packet and haven't cleared it out yet
	if !bytes.Equal(commitment, packetCommitment) {
		return sdkerrors.Wrapf(types.ErrInvalidPacket, "packet commitment bytes are not equal: got (%v), expected (%v)", commitment, packetCommitment)
	}

	switch channel.Ordering {
	case types.ORDERED:
		// check that packet has not been received
		if expectedNextSeqRecv > packet.GetSequence() {
			return sdkerrors.Wrapf(
				types.ErrPacketReceived,
				"packet already received, next sequence receive > packet sequence (%d > %d)", expectedNextSeqRecv, packet.GetSequence(),
			)
		}

		// check that the recv sequence is as claimed
		err = k.connectionKeeper.VerifyNextSequenceRecv(
			ctx, connectionEnd, proofHeight, proof,
			packet.GetDestPort(), packet.GetDestChannel(), expectedNextSeqRecv,
		)
	case types.UNORDERED:
		err = k.connectionKeeper.VerifyPacketReceiptAbsence(
			ctx, connectionEnd, proofHeight, proof,
			packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
		)
	default:
		panic(sdkerrors.Wrapf(types.ErrInvalidChannelOrdering, channel.Ordering.String()))
	}

	if err != nil {
		return err
	}

	return nil
}

// verifyTimeoutOnClose performs the appropriate verification logic for either a timeout on a localhost channel closing or
// a timeout on a counterparty chain's channel closing.
func (k Keeper) verifyTimeoutOnClose(
	ctx sdk.Context,
	packet exported.PacketI,
	channel types.Channel,
	actualNextSeqRecv uint64,
	proofHeight exported.Height,
	proof []byte,
	proofClosed []byte,
) error {
	if channel.ConnectionHops[0] == connectiontypes.LocalhostID {
		commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

		if len(commitment) == 0 {
			EmitTimeoutPacketEvent(ctx, packet, channel)
			// This error indicates that the timeout has already been relayed
			// or there is a misconfigured relayer attempting to prove a timeout
			// for a packet never sent. Core IBC will treat this error as a no-op in order to
			// prevent an entire relay transaction from failing and consuming unnecessary fees.
			return types.ErrNoOpMsg
		}

		packetCommitment := types.CommitPacket(k.cdc, packet)

		// verify we sent the packet and haven't cleared it out yet
		if !bytes.Equal(commitment, packetCommitment) {
			return sdkerrors.Wrapf(types.ErrInvalidPacket, "packet commitment bytes are not equal: got (%v), expected (%v)", commitment, packetCommitment)
		}

		counterpartyChan, found := k.GetChannel(ctx, packet.GetDestPort(), packet.GetDestChannel())
		if !found {
			return sdkerrors.Wrapf(types.ErrChannelNotFound, "port ID (%s) channel ID (%s)", packet.GetDestPort(), packet.GetDestChannel())
		}

		if counterpartyChan.State != types.CLOSED {
			return sdkerrors.Wrapf(types.ErrInvalidChannelState, "failed localhost timeout verification, counterparty channel state is not CLOSED (got %s)", counterpartyChan.State)
		}

		switch channel.Ordering {
		case types.ORDERED:
			// check that packet has not been received
			if actualNextSeqRecv > packet.GetSequence() {
				return sdkerrors.Wrapf(
					types.ErrPacketReceived,
					"packet already received, next sequence receive > packet sequence (%d > %d)", actualNextSeqRecv, packet.GetSequence(),
				)
			}

			actualNextSeq, ok := k.GetNextSequenceRecv(ctx, packet.GetDestPort(), packet.GetDestChannel())
			if !ok {
				return sdkerrors.Wrap(types.ErrSequenceReceiveNotFound, "failed localhost next sequence receive verification, next sequence receive does not exist in store")
			}
			// check that the recv sequence is as claimed
			if actualNextSeqRecv != actualNextSeq {
				return sdkerrors.Wrapf(types.ErrPacketSequenceOutOfOrder, "failed localhost next sequence receive verification, (%d ≠ %d)", actualNextSeq, actualNextSeqRecv)
			}
		case types.UNORDERED:
			_, ok := k.GetPacketReceipt(ctx, packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
			if ok {
				return sdkerrors.Wrapf(types.ErrAcknowledgementExists, "failed localhost packet receipt absence verification")
			}
		default:
			panic(sdkerrors.Wrapf(types.ErrInvalidChannelOrdering, channel.Ordering.String()))
		}

		return nil
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	commitment := k.GetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())

	if len(commitment) == 0 {
		EmitTimeoutPacketEvent(ctx, packet, channel)
		// This error indicates that the timeout has already been relayed
		// or there is a misconfigured relayer attempting to prove a timeout
		// for a packet never sent. Core IBC will treat this error as a no-op in order to
		// prevent an entire relay transaction from failing and consuming unnecessary fees.
		return types.ErrNoOpMsg
	}

	packetCommitment := types.CommitPacket(k.cdc, packet)

	// verify we sent the packet and haven't cleared it out yet
	if !bytes.Equal(commitment, packetCommitment) {
		return sdkerrors.Wrapf(types.ErrInvalidPacket, "packet commitment bytes are not equal: got (%v), expected (%v)", commitment, packetCommitment)
	}

	counterpartyHops := []string{connectionEnd.GetCounterparty().GetConnectionID()}

	counterparty := types.NewCounterparty(packet.GetSourcePort(), packet.GetSourceChannel())
	expectedChannel := types.NewChannel(
		types.CLOSED, channel.Ordering, counterparty, counterpartyHops, channel.Version,
	)

	// check that the opposing channel end has closed
	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofClosed,
		channel.Counterparty.PortId, channel.Counterparty.ChannelId,
		expectedChannel,
	); err != nil {
		return err
	}

	var err error
	switch channel.Ordering {
	case types.ORDERED:
		// check that packet has not been received
		if actualNextSeqRecv > packet.GetSequence() {
			return sdkerrors.Wrapf(types.ErrInvalidPacket, "packet already received, next sequence receive > packet sequence (%d > %d", actualNextSeqRecv, packet.GetSequence())
		}

		// check that the recv sequence is as claimed
		err = k.connectionKeeper.VerifyNextSequenceRecv(
			ctx, connectionEnd, proofHeight, proof,
			packet.GetDestPort(), packet.GetDestChannel(), actualNextSeqRecv,
		)
	case types.UNORDERED:
		err = k.connectionKeeper.VerifyPacketReceiptAbsence(
			ctx, connectionEnd, proofHeight, proof,
			packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence(),
		)
	default:
		panic(sdkerrors.Wrapf(types.ErrInvalidChannelOrdering, channel.Ordering.String()))
	}

	if err != nil {
		return err
	}
	return nil
}
