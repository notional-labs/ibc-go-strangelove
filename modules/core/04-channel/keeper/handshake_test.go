package keeper_test

import (
	"errors"
	"fmt"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"

	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v5/modules/core/03-connection/types"
	"github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
)

type testCase = struct {
	msg      string
	malleate func()
	expPass  bool
}

// TestChanOpenInit tests the OpenInit handshake call for channels. It uses message passing
// to enter into the appropriate state and then calls ChanOpenInit directly. The channel is
// being created on chainA. The port capability must be created on chainA before ChanOpenInit
// can succeed.
func (suite *KeeperTestSuite) TestChanOpenInit() {
	var (
		path     *ibctesting.Path
		features []string
		portCap  *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success", func() {
			suite.coordinator.SetupConnections(path)
			features = []string{"ORDER_ORDERED", "ORDER_UNORDERED"}
			suite.chainA.CreatePortCapability(suite.chainA.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainA.GetPortCapability(ibctesting.MockPort)
		}, true},
		{"channel already exists", func() {
			suite.coordinator.Setup(path)
		}, false},
		{"connection doesn't exist", func() {
			// any non-empty values
			path.EndpointA.ConnectionID = "connection-0"
			path.EndpointB.ConnectionID = "connection-0"
		}, false},
		{"capability is incorrect", func() {
			suite.coordinator.SetupConnections(path)
			features = []string{"ORDER_ORDERED", "ORDER_UNORDERED"}
			portCap = capabilitytypes.NewCapability(3)
		}, false},
		{"connection version not negotiated", func() {
			suite.coordinator.SetupConnections(path)

			// modify connA versions
			conn := path.EndpointA.GetConnection()

			version := connectiontypes.NewVersion("2", []string{"ORDER_ORDERED", "ORDER_UNORDERED"})
			conn.Versions = append(conn.Versions, version)

			suite.chainA.App.GetIBCKeeper().ConnectionKeeper.SetConnection(
				suite.chainA.GetContext(),
				path.EndpointA.ConnectionID, conn,
			)
			features = []string{"ORDER_ORDERED", "ORDER_UNORDERED"}
			suite.chainA.CreatePortCapability(suite.chainA.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainA.GetPortCapability(ibctesting.MockPort)
		}, false},
		{"connection does not support ORDERED channels", func() {
			suite.coordinator.SetupConnections(path)

			// modify connA versions to only support UNORDERED channels
			conn := path.EndpointA.GetConnection()

			version := connectiontypes.NewVersion("1", []string{"ORDER_UNORDERED"})
			conn.Versions = []*connectiontypes.Version{version}

			suite.chainA.App.GetIBCKeeper().ConnectionKeeper.SetConnection(
				suite.chainA.GetContext(),
				path.EndpointA.ConnectionID, conn,
			)
			// NOTE: Opening UNORDERED channels is still expected to pass but ORDERED channels should fail
			features = []string{"ORDER_UNORDERED"}
			suite.chainA.CreatePortCapability(suite.chainA.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainA.GetPortCapability(ibctesting.MockPort)
		}, true},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			// run test for all types of ordering
			for _, order := range []types.Order{types.UNORDERED, types.ORDERED} {
				suite.SetupTest() // reset
				path = ibctesting.NewPath(suite.chainA, suite.chainB)
				path.EndpointA.ChannelConfig.Order = order
				path.EndpointB.ChannelConfig.Order = order

				tc.malleate()

				counterparty := types.NewCounterparty(ibctesting.MockPort, ibctesting.FirstChannelID)

				channelID, cap, err := suite.chainA.App.GetIBCKeeper().ChannelKeeper.ChanOpenInit(
					suite.chainA.GetContext(), path.EndpointA.ChannelConfig.Order, []string{path.EndpointA.ConnectionID},
					path.EndpointA.ChannelConfig.PortID, portCap, counterparty, path.EndpointA.ChannelConfig.Version,
				)

				// check if order is supported by channel to determine expected behaviour
				orderSupported := false
				for _, f := range features {
					if f == order.String() {
						orderSupported = true
					}
				}

				// Testcase must have expectedPass = true AND channel order supported before
				// asserting the channel handshake initiation succeeded
				if tc.expPass && orderSupported {
					suite.Require().NoError(err)
					suite.Require().NotNil(cap)
					suite.Require().Equal(types.FormatChannelIdentifier(0), channelID)

					chanCap, ok := suite.chainA.App.GetScopedIBCKeeper().GetCapability(
						suite.chainA.GetContext(),
						host.ChannelCapabilityPath(path.EndpointA.ChannelConfig.PortID, channelID),
					)
					suite.Require().True(ok, "could not retrieve channel capability after successful ChanOpenInit")
					suite.Require().Equal(chanCap.String(), cap.String(), "channel capability is not correct")
				} else {
					suite.Require().Error(err)
					suite.Require().Nil(cap)
					suite.Require().Equal("", channelID)
				}
			}
		})
	}
}

// TestChanOpenTry tests the OpenTry handshake call for channels. It uses message passing
// to enter into the appropriate state and then calls ChanOpenTry directly. The channel
// is being created on chainB. The port capability must be created on chainB before
// ChanOpenTry can succeed.
func (suite *KeeperTestSuite) TestChanOpenTry() {
	var (
		path       *ibctesting.Path
		portCap    *capabilitytypes.Capability
		heightDiff uint64
	)

	testCases := []testCase{
		{"success", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()
			path.EndpointA.ChanOpenInit()

			suite.chainB.CreatePortCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, true},
		{"connection doesn't exist", func() {
			path.EndpointA.ConnectionID = ibctesting.FirstConnectionID
			path.EndpointB.ConnectionID = ibctesting.FirstConnectionID

			// pass capability check
			suite.chainB.CreatePortCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, false},
		{"connection is not OPEN", func() {
			suite.coordinator.SetupClients(path)
			// pass capability check
			suite.chainB.CreatePortCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)

			err := path.EndpointB.ConnOpenInit()
			suite.Require().NoError(err)
		}, false},
		{"consensus state not found", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()
			path.EndpointA.ChanOpenInit()

			suite.chainB.CreatePortCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)

			heightDiff = 3 // consensus state doesn't exist at this height
		}, false},
		{"channel verification failed", func() {
			// not creating a channel on chainA will result in an invalid proof of existence
			suite.coordinator.SetupConnections(path)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, false},
		{"port capability not found", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()
			path.EndpointA.ChanOpenInit()

			portCap = capabilitytypes.NewCapability(3)
		}, false},
		{"connection version not negotiated", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()
			path.EndpointA.ChanOpenInit()

			// modify connB versions
			conn := path.EndpointB.GetConnection()

			version := connectiontypes.NewVersion("2", []string{"ORDER_ORDERED", "ORDER_UNORDERED"})
			conn.Versions = append(conn.Versions, version)

			suite.chainB.App.GetIBCKeeper().ConnectionKeeper.SetConnection(
				suite.chainB.GetContext(),
				path.EndpointB.ConnectionID, conn,
			)
			suite.chainB.CreatePortCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, false},
		{"connection does not support ORDERED channels", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()
			path.EndpointA.ChanOpenInit()

			// modify connA versions to only support UNORDERED channels
			conn := path.EndpointA.GetConnection()

			version := connectiontypes.NewVersion("1", []string{"ORDER_UNORDERED"})
			conn.Versions = []*connectiontypes.Version{version}

			suite.chainA.App.GetIBCKeeper().ConnectionKeeper.SetConnection(
				suite.chainA.GetContext(),
				path.EndpointA.ConnectionID, conn,
			)
			suite.chainA.CreatePortCapability(suite.chainA.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainA.GetPortCapability(ibctesting.MockPort)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			heightDiff = 0    // must be explicitly changed in malleate
			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			if path.EndpointB.ClientID != "" {
				// ensure client is up to date
				err := path.EndpointB.UpdateClient()
				suite.Require().NoError(err)
			}

			counterparty := types.NewCounterparty(path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID)

			channelKey := host.ChannelKey(counterparty.PortId, counterparty.ChannelId)
			proof, proofHeight := suite.chainA.QueryProof(channelKey)

			channelID, cap, err := suite.chainB.App.GetIBCKeeper().ChannelKeeper.ChanOpenTry(
				suite.chainB.GetContext(), types.ORDERED, []string{path.EndpointB.ConnectionID},
				path.EndpointB.ChannelConfig.PortID, portCap, counterparty, path.EndpointA.ChannelConfig.Version,
				proof, malleateHeight(proofHeight, heightDiff),
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(cap)

				chanCap, ok := suite.chainB.App.GetScopedIBCKeeper().GetCapability(
					suite.chainB.GetContext(),
					host.ChannelCapabilityPath(path.EndpointB.ChannelConfig.PortID, channelID),
				)
				suite.Require().True(ok, "could not retrieve channel capapbility after successful ChanOpenTry")
				suite.Require().Equal(chanCap.String(), cap.String(), "channel capability is not correct")
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestChanOpenAck tests the OpenAck handshake call for channels. It uses message passing
// to enter into the appropriate state and then calls ChanOpenAck directly. The handshake
// call is occurring on chainA.
func (suite *KeeperTestSuite) TestChanOpenAck() {
	var (
		path                  *ibctesting.Path
		counterpartyChannelID string
		channelCap            *capabilitytypes.Capability
		heightDiff            uint64
	)

	testCases := []testCase{
		{"success", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()
			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, true},
		{"success with empty stored counterparty channel ID", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			// set the channel's counterparty channel identifier to empty string
			channel := path.EndpointA.GetChannel()
			channel.Counterparty.ChannelId = ""

			// use a different channel identifier
			counterpartyChannelID = path.EndpointB.ChannelID

			suite.chainA.App.GetIBCKeeper().ChannelKeeper.SetChannel(suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, channel)

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not INIT", func() {
			// create fully open channels on both chains
			suite.coordinator.Setup(path)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"connection not found", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()
			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)

			// set the channel's connection hops to wrong connection ID
			channel := path.EndpointA.GetChannel()
			channel.ConnectionHops[0] = "doesnotexist"
			suite.chainA.App.GetIBCKeeper().ChannelKeeper.SetChannel(suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, channel)
		}, false},
		{"connection is not OPEN", func() {
			suite.coordinator.SetupClients(path)

			err := path.EndpointA.ConnOpenInit()
			suite.Require().NoError(err)

			// create channel in init
			path.SetChannelOrdered()

			err = path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			suite.chainA.CreateChannelCapability(suite.chainA.GetSimApp().ScopedIBCMockKeeper, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"consensus state not found", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)

			heightDiff = 3 // consensus state doesn't exist at this height
		}, false},
		{"invalid counterparty channel identifier", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			counterpartyChannelID = "otheridentifier"

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"channel verification failed", func() {
			// chainB is INIT, chainA in TRYOPEN
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointB.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointA.ChanOpenTry()
			suite.Require().NoError(err)

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"channel capability not found", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()
			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			path.EndpointB.ChanOpenTry()

			channelCap = capabilitytypes.NewCapability(6)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()          // reset
			counterpartyChannelID = "" // must be explicitly changed in malleate
			heightDiff = 0             // must be explicitly changed
			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			if counterpartyChannelID == "" {
				counterpartyChannelID = ibctesting.FirstChannelID
			}

			if path.EndpointA.ClientID != "" {
				// ensure client is up to date
				err := path.EndpointA.UpdateClient()
				suite.Require().NoError(err)
			}

			channelKey := host.ChannelKey(path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID)
			proof, proofHeight := suite.chainB.QueryProof(channelKey)

			err := suite.chainA.App.GetIBCKeeper().ChannelKeeper.ChanOpenAck(
				suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, channelCap, path.EndpointB.ChannelConfig.Version, counterpartyChannelID,
				proof, malleateHeight(proofHeight, heightDiff),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestChanOpenConfirm tests the OpenAck handshake call for channels. It uses message passing
// to enter into the appropriate state and then calls ChanOpenConfirm directly. The handshake
// call is occurring on chainB.
func (suite *KeeperTestSuite) TestChanOpenConfirm() {
	var (
		path       *ibctesting.Path
		channelCap *capabilitytypes.Capability
		heightDiff uint64
	)
	testCases := []testCase{
		{"success", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			err = path.EndpointA.ChanOpenAck()
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
		}, true},
		{"channel doesn't exist", func() {}, false},
		{"channel state is not TRYOPEN", func() {
			// create fully open channels on both cahins
			suite.coordinator.Setup(path)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
		}, false},
		{"connection not found", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			err = path.EndpointA.ChanOpenAck()
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			// set the channel's connection hops to wrong connection ID
			channel := path.EndpointB.GetChannel()
			channel.ConnectionHops[0] = "doesnotexist"
			suite.chainB.App.GetIBCKeeper().ChannelKeeper.SetChannel(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, channel)
		}, false},
		{"connection is not OPEN", func() {
			suite.coordinator.SetupClients(path)

			err := path.EndpointB.ConnOpenInit()
			suite.Require().NoError(err)

			suite.chainB.CreateChannelCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID)
		}, false},
		{"consensus state not found", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			err = path.EndpointA.ChanOpenAck()
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			heightDiff = 3
		}, false},
		{"channel verification failed", func() {
			// chainA is INIT, chainB in TRYOPEN
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
		}, false},
		{"channel capability not found", func() {
			suite.coordinator.SetupConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.ChanOpenTry()
			suite.Require().NoError(err)

			err = path.EndpointA.ChanOpenAck()
			suite.Require().NoError(err)

			channelCap = capabilitytypes.NewCapability(6)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			heightDiff = 0    // must be explicitly changed
			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			if path.EndpointB.ClientID != "" {
				// ensure client is up to date
				err := path.EndpointB.UpdateClient()
				suite.Require().NoError(err)

			}

			channelKey := host.ChannelKey(path.EndpointA.ChannelConfig.PortID, ibctesting.FirstChannelID)
			proof, proofHeight := suite.chainA.QueryProof(channelKey)

			err := suite.chainB.App.GetIBCKeeper().ChannelKeeper.ChanOpenConfirm(
				suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID,
				channelCap, proof, malleateHeight(proofHeight, heightDiff),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestChanCloseInit tests the initial closing of a handshake on chainA by calling
// ChanCloseInit. Both chains will use message passing to setup OPEN channels.
func (suite *KeeperTestSuite) TestChanCloseInit() {
	var (
		path       *ibctesting.Path
		channelCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success", func() {
			suite.coordinator.Setup(path)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, true},
		{"channel doesn't exist", func() {
			// any non-nil values work for connections
			path.EndpointA.ConnectionID = ibctesting.FirstConnectionID
			path.EndpointB.ConnectionID = ibctesting.FirstConnectionID

			path.EndpointA.ChannelID = ibctesting.FirstChannelID
			path.EndpointB.ChannelID = ibctesting.FirstChannelID

			// ensure channel capability check passes
			suite.chainA.CreateChannelCapability(suite.chainA.GetSimApp().ScopedIBCMockKeeper, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"channel state is CLOSED", func() {
			suite.coordinator.Setup(path)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)

			// close channel
			err := path.EndpointA.SetChannelClosed()
			suite.Require().NoError(err)
		}, false},
		{"connection not found", func() {
			suite.coordinator.Setup(path)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)

			// set the channel's connection hops to wrong connection ID
			channel := path.EndpointA.GetChannel()
			channel.ConnectionHops[0] = "doesnotexist"
			suite.chainA.App.GetIBCKeeper().ChannelKeeper.SetChannel(suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, channel)
		}, false},
		{"connection is not OPEN", func() {
			suite.coordinator.SetupClients(path)

			err := path.EndpointA.ConnOpenInit()
			suite.Require().NoError(err)

			// create channel in init
			path.SetChannelOrdered()
			err = path.EndpointA.ChanOpenInit()

			// ensure channel capability check passes
			suite.chainA.CreateChannelCapability(suite.chainA.GetSimApp().ScopedIBCMockKeeper, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"channel capability not found", func() {
			suite.coordinator.Setup(path)
			channelCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			err := suite.chainA.App.GetIBCKeeper().ChannelKeeper.ChanCloseInit(
				suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, ibctesting.FirstChannelID, channelCap,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestChanCloseConfirm tests the confirming closing channel ends by calling ChanCloseConfirm
// on chainB. Both chains will use message passing to setup OPEN channels. ChanCloseInit is
// bypassed on chainA by setting the channel state in the ChannelKeeper.
func (suite *KeeperTestSuite) TestChanCloseConfirm() {
	var (
		path       *ibctesting.Path
		channelCap *capabilitytypes.Capability
		heightDiff uint64
	)

	testCases := []testCase{
		{"success", func() {
			suite.coordinator.Setup(path)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			err := path.EndpointA.SetChannelClosed()
			suite.Require().NoError(err)
		}, true},
		{"channel doesn't exist", func() {
			// any non-nil values work for connections
			path.EndpointA.ChannelID = ibctesting.FirstChannelID
			path.EndpointB.ChannelID = ibctesting.FirstChannelID

			// ensure channel capability check passes
			suite.chainB.CreateChannelCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID)
		}, false},
		{"channel state is CLOSED", func() {
			suite.coordinator.Setup(path)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			err := path.EndpointB.SetChannelClosed()
			suite.Require().NoError(err)
		}, false},
		{"connection not found", func() {
			suite.coordinator.Setup(path)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			// set the channel's connection hops to wrong connection ID
			channel := path.EndpointB.GetChannel()
			channel.ConnectionHops[0] = "doesnotexist"
			suite.chainB.App.GetIBCKeeper().ChannelKeeper.SetChannel(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, channel)
		}, false},
		{"connection is not OPEN", func() {
			suite.coordinator.SetupClients(path)

			err := path.EndpointB.ConnOpenInit()
			suite.Require().NoError(err)

			// create channel in init
			path.SetChannelOrdered()
			err = path.EndpointB.ChanOpenInit()
			suite.Require().NoError(err)

			// ensure channel capability check passes
			suite.chainB.CreateChannelCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
		}, false},
		{"consensus state not found", func() {
			suite.coordinator.Setup(path)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			err := path.EndpointA.SetChannelClosed()
			suite.Require().NoError(err)

			heightDiff = 3
		}, false},
		{"channel verification failed", func() {
			// channel not closed
			suite.coordinator.Setup(path)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
		}, false},
		{"channel capability not found", func() {
			suite.coordinator.Setup(path)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			err := path.EndpointA.SetChannelClosed()
			suite.Require().NoError(err)

			channelCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			heightDiff = 0    // must explicitly be changed
			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			channelKey := host.ChannelKey(path.EndpointA.ChannelConfig.PortID, ibctesting.FirstChannelID)
			proof, proofHeight := suite.chainA.QueryProof(channelKey)

			err := suite.chainB.App.GetIBCKeeper().ChannelKeeper.ChanCloseConfirm(
				suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID, channelCap,
				proof, malleateHeight(proofHeight, heightDiff),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestLocalhostChanOpenInit tests the OpenInit handshake call for channels where both ends are on the same chain.
// It uses message passing to enter into the appropriate state and then calls ChanOpenInit directly. The channel is
// being created on chainA. The port capability must be created on chainA before ChanOpenInit can succeed.
func (suite *KeeperTestSuite) TestLocalhostChanOpenInit() {
	var (
		path    *ibctesting.Path
		portCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success", func() {
			suite.coordinator.SetupLocalhostConnections(path)
			suite.chainA.CreatePortCapability(suite.chainA.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainA.GetPortCapability(ibctesting.MockPort)
		}, true},
		{"channel already exists", func() {
			suite.coordinator.SetupLocalhost(path)
		}, false},
		{"capability is incorrect", func() {
			suite.coordinator.SetupLocalhostConnections(path)

			portCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			// run test for all types of ordering
			for _, order := range []types.Order{types.UNORDERED, types.ORDERED} {
				suite.SetupLocalhostTest() // reset test
				path = ibctesting.NewPath(suite.chainA, suite.chainB)
				path.EndpointA.ChannelConfig.Order = order
				path.EndpointB.ChannelConfig.Order = order

				tc.malleate()

				counterparty := types.NewCounterparty(ibctesting.MockPort, "")

				channelID, cap, err := suite.chainA.App.GetIBCKeeper().ChannelKeeper.ChanOpenInit(
					suite.chainA.GetContext(), path.EndpointA.ChannelConfig.Order, []string{path.EndpointA.ConnectionID},
					path.EndpointA.ChannelConfig.PortID, portCap, counterparty, path.EndpointA.ChannelConfig.Version,
				)

				// Testcase must have expectedPass = true AND channel order supported before
				// asserting the channel handshake initiation succeeded
				if tc.expPass {
					suite.Require().NoError(err)
					suite.Require().NotNil(cap)
					suite.Require().Equal(types.FormatChannelIdentifier(0), channelID)

					chanCap, ok := suite.chainA.App.GetScopedIBCKeeper().GetCapability(
						suite.chainA.GetContext(),
						host.ChannelCapabilityPath(path.EndpointA.ChannelConfig.PortID, channelID),
					)
					suite.Require().True(ok, "could not retrieve channel capability after successful ChanOpenInit")
					suite.Require().Equal(chanCap.String(), cap.String(), "channel capability is not correct")
				} else {
					suite.Require().Error(err)
					suite.Require().Nil(cap)
					suite.Require().Equal("", channelID)
				}
			}
		})
	}
}

// TestLocalhostChanOpenTry tests the OpenTry handshake call for channels where both ends are on the same chain.
// It uses message passing to enter into the appropriate state and then calls ChanOpenTry directly. The channel
// is being created on chainB. The port capability must be created on chainB before ChanOpenTry can succeed.
func (suite *KeeperTestSuite) TestLocalhostChanOpenTry() {
	var (
		path     *ibctesting.Path
		portCap  *capabilitytypes.Capability
		expError error
	)

	testCases := []testCase{
		{"success: ORDERED channel", func() {
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()
			path.EndpointA.ChanOpenInit()

			suite.chainB.CreatePortCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, true},
		{"success: UNORDERED channel", func() {
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelUnordered()
			path.EndpointA.ChanOpenInit()

			suite.chainB.CreatePortCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, true},
		{"localhost channel verification failed", func() {
			expError = types.ErrChannelNotFound
			// the channel for EndpointA does not exist in state
			suite.coordinator.SetupLocalhostConnections(path)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, false},
		{"localhost channel verification failed", func() {
			expError = types.ErrChannelNotFound
			// the channel for EndpointA does not exist in state
			suite.coordinator.SetupLocalhostConnections(path)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, false},
		{"port capability not found", func() {
			expError = porttypes.ErrInvalidPort
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()
			path.EndpointA.ChanOpenInit()

			portCap = capabilitytypes.NewCapability(3)
		}, false},
		{"counterparty channel not in INIT state", func() {
			expError = types.ErrInvalidChannel
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()

			// move channel state for EndpointA to the OPEN state then call ChanOpenTry again
			suite.Require().NoError(path.EndpointA.ChanOpenInit())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenTry())
			suite.Require().NoError(path.EndpointA.LocalhostChanOpenAck())

			suite.chainB.CreatePortCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, false},
		{"verify channel end invalid channel", func() {
			expError = types.ErrInvalidChannel
			suite.coordinator.SetupLocalhostConnections(path)
			path.EndpointA.ChanOpenInit()

			path.EndpointA.ChannelConfig.Version = "invalidversion"

			suite.chainB.CreatePortCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, ibctesting.MockPort)
			portCap = suite.chainB.GetPortCapability(ibctesting.MockPort)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupLocalhostTest() // reset test
			expError = nil             // must be explicitly changed by failed cases

			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			counterparty := types.NewCounterparty(path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID)

			channelID, cap, err := suite.chainB.App.GetIBCKeeper().ChannelKeeper.ChanOpenTry(
				suite.chainB.GetContext(), path.EndpointA.ChannelConfig.Order, []string{path.EndpointB.ConnectionID},
				path.EndpointB.ChannelConfig.PortID, portCap, counterparty, path.EndpointA.ChannelConfig.Version,
				nil, nil,
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(cap)

				suite.Require().NotEqual(path.EndpointA.ChannelID, path.EndpointB.ChannelID)

				chanCap, ok := suite.chainB.App.GetScopedIBCKeeper().GetCapability(
					suite.chainB.GetContext(),
					host.ChannelCapabilityPath(path.EndpointB.ChannelConfig.PortID, channelID),
				)
				suite.Require().True(ok, "could not retrieve channel capapbility after successful ChanOpenTry")
				suite.Require().Equal(chanCap.String(), cap.String(), "channel capability is not correct")
			} else {
				suite.Require().Error(err)
				// only check if expError is set, since not all error codes can be known
				if expError != nil {
					suite.Require().True(errors.Is(err, expError))
				}
			}
		})
	}
}

// TestLocalhostChanOpenAck tests the OpenAck handshake call for channels where both ends are on the same chain.
// It uses message passing to enter into the appropriate state and then calls ChanOpenAck directly. The handshake
// call is occurring on chainA.
func (suite *KeeperTestSuite) TestLocalhostChanOpenAck() {
	var (
		path                  *ibctesting.Path
		counterpartyChannelID string
		channelCap            *capabilitytypes.Capability
		expError              error
	)

	testCases := []testCase{
		{"success: ORDERED channel", func() {
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.LocalhostChanOpenTry()
			suite.Require().NoError(err)

			counterpartyChannelID = path.EndpointB.ChannelID

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, true},
		{"success: UNORDERED channel", func() {
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelUnordered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.LocalhostChanOpenTry()
			suite.Require().NoError(err)

			counterpartyChannelID = path.EndpointB.ChannelID

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, true},
		{"channel doesn't exist", func() {
			expError = types.ErrChannelNotFound
		}, false},
		{"channel state is not INIT", func() {
			expError = types.ErrInvalidChannelState
			suite.coordinator.SetupLocalhostConnections(path)

			// create fully open channels on both ends
			suite.Require().NoError(path.EndpointA.ChanOpenInit())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenTry())
			suite.Require().NoError(path.EndpointA.LocalhostChanOpenAck())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenConfirm())

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"invalid counterparty channel identifier", func() {
			expError = types.ErrChannelNotFound
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.LocalhostChanOpenTry()
			suite.Require().NoError(err)

			counterpartyChannelID = "otheridentifier"

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"counterparty channel verification failed", func() {
			expError = types.ErrInvalidChannelState

			// chainB is INIT, chainA in TRYOPEN
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointB.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointA.LocalhostChanOpenTry()
			suite.Require().NoError(err)

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"channel capability not found", func() {
			expError = types.ErrChannelCapabilityNotFound
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()
			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			path.EndpointB.LocalhostChanOpenTry()

			channelCap = capabilitytypes.NewCapability(6)
		}, false},
		{"verify channel end invalid channel", func() {
			expError = types.ErrInvalidChannel
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelUnordered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.LocalhostChanOpenTry()
			suite.Require().NoError(err)

			counterpartyChannelID = path.EndpointB.ChannelID

			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)

			path.EndpointB.ChannelConfig.Version = "invalidversion"
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupLocalhostTest() // reset
			expError = nil             // must be explicitly changed by failed cases
			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			err := suite.chainA.App.GetIBCKeeper().ChannelKeeper.ChanOpenAck(
				suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, channelCap, path.EndpointB.ChannelConfig.Version, counterpartyChannelID,
				nil, nil,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				// only check if expError is set, since not all error codes can be known
				if expError != nil {
					suite.Require().True(errors.Is(err, expError))
				}
			}
		})
	}
}

// TestLocalhostChanOpenConfirm tests the OpenAck handshake call for channels where both ends are on the same chain.
// It uses message passing to enter into the appropriate state and then calls ChanOpenConfirm directly. The handshake
// call is occurring on chainB.
func (suite *KeeperTestSuite) TestLocalhostChanOpenConfirm() {
	var (
		path       *ibctesting.Path
		channelCap *capabilitytypes.Capability
		expError   error
	)

	testCases := []testCase{
		{"success", func() {
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.LocalhostChanOpenTry()
			suite.Require().NoError(err)

			err = path.EndpointA.LocalhostChanOpenAck()
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
		}, true},
		{"channel doesn't exist", func() {
			expError = types.ErrChannelNotFound
		}, false},
		{"channel state is not TRYOPEN", func() {
			expError = types.ErrInvalidChannelState

			// create fully open channels on both ends
			suite.coordinator.SetupLocalhostConnections(path)

			suite.Require().NoError(path.EndpointA.ChanOpenInit())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenTry())
			suite.Require().NoError(path.EndpointA.LocalhostChanOpenAck())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenConfirm())

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
		}, false},
		{"counterparty channel verification failed", func() {
			expError = types.ErrInvalidChannel

			// chainA is INIT, chainB in TRYOPEN
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.LocalhostChanOpenTry()
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
		}, false},
		{"channel capability not found", func() {
			expError = types.ErrChannelCapabilityNotFound
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.LocalhostChanOpenTry()
			suite.Require().NoError(err)

			err = path.EndpointA.LocalhostChanOpenAck()
			suite.Require().NoError(err)

			channelCap = capabilitytypes.NewCapability(6)
		}, false},
		{"verify channel end invalid channel", func() {
			expError = types.ErrInvalidChannel
			suite.coordinator.SetupLocalhostConnections(path)
			path.SetChannelOrdered()

			err := path.EndpointA.ChanOpenInit()
			suite.Require().NoError(err)

			err = path.EndpointB.LocalhostChanOpenTry()
			suite.Require().NoError(err)

			err = path.EndpointA.LocalhostChanOpenAck()
			suite.Require().NoError(err)

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			// get the channel and override the version so that the verifyChannelEnd call fails
			channel := path.EndpointB.GetChannel()
			channel.Version = "invalidversion"
			suite.chainB.App.GetIBCKeeper().ChannelKeeper.SetChannel(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, channel)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupLocalhostTest() // reset
			expError = nil             // must be explicitly changed by failed cases
			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			err := suite.chainB.App.GetIBCKeeper().ChannelKeeper.ChanOpenConfirm(
				suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID,
				channelCap, nil, nil,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				// only check if expError is set, since not all error codes can be known
				if expError != nil {
					suite.Require().True(errors.Is(err, expError))
				}
			}
		})
	}
}

// TestLocalhostChanCloseInit tests the initial closing of a handshake on chainA by calling
// ChanCloseInit. Both chains will use message passing to setup OPEN channels.
func (suite *KeeperTestSuite) TestLocalhostChanCloseInit() {
	var (
		path       *ibctesting.Path
		channelCap *capabilitytypes.Capability
	)

	testCases := []testCase{
		{"success", func() {
			suite.coordinator.SetupLocalhost(path)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, true},
		{"channel doesn't exist", func() {
			// any non-nil values work for connections
			path.EndpointA.ConnectionID = ibctesting.FirstConnectionID
			path.EndpointB.ConnectionID = ibctesting.FirstConnectionID

			path.EndpointA.ChannelID = ibctesting.FirstChannelID
			path.EndpointB.ChannelID = ibctesting.FirstChannelID

			// ensure channel capability check passes
			suite.chainA.CreateChannelCapability(suite.chainA.GetSimApp().ScopedIBCMockKeeper, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		}, false},
		{"channel state is CLOSED", func() {
			suite.coordinator.SetupLocalhost(path)
			channelCap = suite.chainA.GetChannelCapability(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)

			// close channel
			path.EndpointA.SetLocalhostChannelClosed()
		}, false},
		{"channel capability not found", func() {
			suite.coordinator.SetupLocalhost(path)
			channelCap = capabilitytypes.NewCapability(3)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupLocalhostTest() // reset
			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			err := suite.chainA.App.GetIBCKeeper().ChannelKeeper.ChanCloseInit(
				suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, ibctesting.FirstChannelID, channelCap,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestLocalhostChanCloseConfirm tests the confirming closing channel ends by calling ChanCloseConfirm
// on chainB. Both chains will use message passing to setup OPEN channels. ChanCloseInit is
// bypassed on chainA by setting the channel state in the ChannelKeeper.
func (suite *KeeperTestSuite) TestLocalhostChanCloseConfirm() {
	var (
		path       *ibctesting.Path
		channelCap *capabilitytypes.Capability
		expError   error
	)

	testCases := []testCase{
		{"success", func() {
			suite.coordinator.SetupLocalhostConnections(path)

			suite.Require().NoError(path.EndpointA.ChanOpenInit())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenTry())
			suite.Require().NoError(path.EndpointA.LocalhostChanOpenAck())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenConfirm())

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			channel := path.EndpointA.GetChannel()

			channel.State = types.CLOSED
			path.EndpointA.Chain.App.GetIBCKeeper().ChannelKeeper.SetChannel(path.EndpointA.Chain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, channel)
		}, true},
		{"channel doesn't exist", func() {
			expError = types.ErrChannelNotFound

			path.EndpointA.ChannelID = ibctesting.FirstChannelID
			path.EndpointB.ChannelID = ibctesting.FirstChannelID

			// ensure channel capability check passes
			suite.chainB.CreateChannelCapability(suite.chainB.GetSimApp().ScopedIBCMockKeeper, path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID)
			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, ibctesting.FirstChannelID)
		}, false},
		{"channel state is CLOSED", func() {
			expError = types.ErrInvalidChannelState
			suite.coordinator.SetupLocalhostConnections(path)

			suite.Require().NoError(path.EndpointA.ChanOpenInit())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenTry())
			suite.Require().NoError(path.EndpointA.LocalhostChanOpenAck())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenConfirm())

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			channel := path.EndpointB.GetChannel()

			channel.State = types.CLOSED
			path.EndpointB.Chain.App.GetIBCKeeper().ChannelKeeper.SetChannel(path.EndpointB.Chain.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, channel)
		}, false},
		{"counterparty channel verification failed", func() {
			expError = types.ErrInvalidChannel

			// channel not closed
			suite.coordinator.SetupLocalhostConnections(path)

			suite.Require().NoError(path.EndpointA.ChanOpenInit())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenTry())
			suite.Require().NoError(path.EndpointA.LocalhostChanOpenAck())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenConfirm())

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
		}, false},
		{"channel capability not found", func() {
			expError = types.ErrChannelCapabilityNotFound
			suite.coordinator.SetupLocalhostConnections(path)

			suite.Require().NoError(path.EndpointA.ChanOpenInit())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenTry())
			suite.Require().NoError(path.EndpointA.LocalhostChanOpenAck())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenConfirm())

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			channel := path.EndpointA.GetChannel()

			channel.State = types.CLOSED
			path.EndpointA.Chain.App.GetIBCKeeper().ChannelKeeper.SetChannel(path.EndpointA.Chain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, channel)

			channelCap = capabilitytypes.NewCapability(3)
		}, false},
		{"verify channel end invalid channel", func() {
			expError = types.ErrInvalidChannel
			suite.coordinator.SetupLocalhostConnections(path)

			suite.Require().NoError(path.EndpointA.ChanOpenInit())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenTry())
			suite.Require().NoError(path.EndpointA.LocalhostChanOpenAck())
			suite.Require().NoError(path.EndpointB.LocalhostChanOpenConfirm())

			channelCap = suite.chainB.GetChannelCapability(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

			channel := path.EndpointA.GetChannel()

			channel.State = types.CLOSED
			path.EndpointA.Chain.App.GetIBCKeeper().ChannelKeeper.SetChannel(path.EndpointA.Chain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, channel)

			// get the channel and override the version so that the verifyChannelEnd call fails
			channel = path.EndpointB.GetChannel()
			channel.Version = "invalidversion"
			suite.chainB.App.GetIBCKeeper().ChannelKeeper.SetChannel(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, channel)
		}, false},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupLocalhostTest() // reset
			expError = nil             // must be explicitly changed by failed cases
			path = ibctesting.NewPath(suite.chainA, suite.chainB)

			tc.malleate()

			err := suite.chainB.App.GetIBCKeeper().ChannelKeeper.ChanCloseConfirm(
				suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, channelCap,
				nil, nil,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				// only check if expError is set, since not all error codes can be known
				if expError != nil {
					suite.Require().True(errors.Is(err, expError))
				}
			}
		})
	}
}

func malleateHeight(height exported.Height, diff uint64) exported.Height {
	return clienttypes.NewHeight(height.GetRevisionNumber(), height.GetRevisionHeight()+diff)
}
