(window.webpackJsonp=window.webpackJsonp||[]).push([[108],{669:function(e,t,i){"use strict";i.r(t);var a=i(1),s=Object(a.a)({},(function(){var e=this,t=e.$createElement,i=e._self._c||t;return i("ContentSlotsDistributor",{attrs:{"slot-key":e.$parent.slotKey}},[i("h1",{attrs:{id:"implementing-the-clientstate-interface"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#implementing-the-clientstate-interface"}},[e._v("#")]),e._v(" Implementing the "),i("code",[e._v("ClientState")]),e._v(" interface")]),e._v(" "),i("p",[e._v("Learn how to implement the "),i("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v6.0.0/modules/core/exported/client.go#L40",target:"_blank",rel:"noopener noreferrer"}},[i("code",[e._v("ClientState")]),i("OutboundLink")],1),e._v(" interface.")]),e._v(" "),i("h2",{attrs:{id:"clienttype-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#clienttype-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("ClientType")]),e._v(" method")]),e._v(" "),i("p",[i("code",[e._v("ClientType")]),e._v(" should return a unique string identifier of the light client. This will be used when generating a client identifier.\nThe format is created as follows: "),i("code",[e._v("ClientType-{N}")]),e._v(" where "),i("code",[e._v("{N}")]),e._v(" is the unique global nonce associated with a specific client.")]),e._v(" "),i("h2",{attrs:{id:"getlatestheight-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#getlatestheight-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("GetLatestHeight")]),e._v(" method")]),e._v(" "),i("p",[i("code",[e._v("GetLatestHeight")]),e._v(" should return the latest block height that the client state represents.")]),e._v(" "),i("h2",{attrs:{id:"validate-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#validate-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("Validate")]),e._v(" method")]),e._v(" "),i("p",[i("code",[e._v("Validate")]),e._v(" should validate every client state field and should return an error if any value is invalid. The light client\nimplementer is in charge of determining which checks are required. See the "),i("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v6.0.0/modules/light-clients/07-tendermint/types/client_state.go#L101",target:"_blank",rel:"noopener noreferrer"}},[e._v("tendermint light client implementation"),i("OutboundLink")],1),e._v(" as a reference.")]),e._v(" "),i("h2",{attrs:{id:"status-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#status-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("Status")]),e._v(" method")]),e._v(" "),i("p",[i("code",[e._v("Status")]),e._v(" must return the status of the client.")]),e._v(" "),i("ul",[i("li",[e._v("An "),i("code",[e._v("Active")]),e._v(" status indicates that clients are allowed to process packets.")]),e._v(" "),i("li",[e._v("A "),i("code",[e._v("Frozen")]),e._v(" status indicates that a client is not allowed to be used.")]),e._v(" "),i("li",[e._v("An "),i("code",[e._v("Expired")]),e._v(" status indicates that a client is not allowed to be used.")]),e._v(" "),i("li",[e._v("An "),i("code",[e._v("Unknown")]),e._v(" status indicates that there was an error in determining the status of a client.")])]),e._v(" "),i("p",[e._v("All possible Status types can be found "),i("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v6.0.0/modules/core/exported/client.go#L26-L36",target:"_blank",rel:"noopener noreferrer"}},[e._v("here"),i("OutboundLink")],1),e._v(".")]),e._v(" "),i("p",[e._v("This field is returned by the gRPC "),i("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v6.0.0/modules/core/02-client/types/query.pb.go#L665",target:"_blank",rel:"noopener noreferrer"}},[e._v("QueryClientStatusResponse"),i("OutboundLink")],1),e._v(" endpoint.")]),e._v(" "),i("h2",{attrs:{id:"zerocustomfields-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#zerocustomfields-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("ZeroCustomFields")]),e._v(" method")]),e._v(" "),i("p",[i("code",[e._v("ZeroCustomFields")]),e._v(" should return a copy of the light client with all client customizable fields with their zero value. It should not mutate the fields of the light client.\nThis method is used when "),i("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v6.0.0/modules/core/02-client/keeper/proposal.go#L89",target:"_blank",rel:"noopener noreferrer"}},[e._v("scheduling upgrades"),i("OutboundLink")],1),e._v(". Upgrades are used to upgrade chain specific fields.\nIn the tendermint case, this may be the chainID or the unbonding period.\nFor more information about client upgrades see "),i("RouterLink",{attrs:{to:"/ibc/upgrades/developer-guide.html"}},[e._v("the developer guide")]),e._v(".")],1),e._v(" "),i("h2",{attrs:{id:"gettimestampatheight-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#gettimestampatheight-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("GetTimestampAtHeight")]),e._v(" method")]),e._v(" "),i("p",[i("code",[e._v("GetTimestampAtHeight")]),e._v(" must return the timestamp for the consensus state associated with the provided height.\nThis value is used to facilitate timeouts by checking the packet timeout timestamp against the returned value.")]),e._v(" "),i("h2",{attrs:{id:"initialize-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#initialize-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("Initialize")]),e._v(" method")]),e._v(" "),i("p",[e._v("Clients must validate the initial consensus state, and set the initial client state and consensus state in the provided client store.\nClients may also store any necessary client-specific metadata.")]),e._v(" "),i("p",[i("code",[e._v("Initialize")]),e._v(" is called when a "),i("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/main/modules/core/02-client/keeper/client.go#L32",target:"_blank",rel:"noopener noreferrer"}},[e._v("client is created"),i("OutboundLink")],1),e._v(".")]),e._v(" "),i("h2",{attrs:{id:"verifymembership-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#verifymembership-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("VerifyMembership")]),e._v(" method")]),e._v(" "),i("p",[i("code",[e._v("VerifyMembership")]),e._v(" must verify the existence of a value at a given CommitmentPath at the specified height. For more information about membership proofs\nsee "),i("RouterLink",{attrs:{to:"/ibc/light-clients/proofs.html"}},[e._v("the proof docs")]),e._v(".")],1),e._v(" "),i("h2",{attrs:{id:"verifynonmembership-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#verifynonmembership-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("VerifyNonMembership")]),e._v(" method")]),e._v(" "),i("p",[i("code",[e._v("VerifyNonMembership")]),e._v(" must verify the absence of a value at a given CommitmentPath at a specified height. For more information about non membership proofs\nsee "),i("RouterLink",{attrs:{to:"/ibc/light-clients/proofs.html"}},[e._v("the proof docs")]),e._v(".")],1),e._v(" "),i("h2",{attrs:{id:"verifyclientmessage-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#verifyclientmessage-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("VerifyClientMessage")]),e._v(" method")]),e._v(" "),i("p",[e._v("VerifyClientMessage must verify a ClientMessage. A ClientMessage could be a Header, Misbehaviour, or batch update.\nIt must handle each type of ClientMessage appropriately. Calls to CheckForMisbehaviour, UpdateState, and UpdateStateOnMisbehaviour\nwill assume that the content of the ClientMessage has been verified and can be trusted. An error should be returned\nif the ClientMessage fails to verify.")]),e._v(" "),i("h2",{attrs:{id:"checkformisbehaviour-method"}},[i("a",{staticClass:"header-anchor",attrs:{href:"#checkformisbehaviour-method"}},[e._v("#")]),e._v(" "),i("code",[e._v("CheckForMisbehaviour")]),e._v(" method")]),e._v(" "),i("p",[e._v("Checks for evidence of a misbehaviour in Header or Misbehaviour type. It assumes the ClientMessage\nhas already been verified.")])])}),[],!1,null,null,null);t.default=s.exports}}]);