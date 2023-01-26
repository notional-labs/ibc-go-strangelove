(window.webpackJsonp=window.webpackJsonp||[]).push([[114],{675:function(e,t,l){"use strict";l.r(t);var o=l(1),c=Object(o.a)({},(function(){var e=this,t=e.$createElement,l=e._self._c||t;return l("ContentSlotsDistributor",{attrs:{"slot-key":e.$parent.slotKey}},[l("h1",{attrs:{id:"setup"}},[l("a",{staticClass:"header-anchor",attrs:{href:"#setup"}},[e._v("#")]),e._v(" Setup")]),e._v(" "),l("p",{attrs:{synopsis:""}},[e._v("Learn how to configure light client modules and create clients using core IBC and the "),l("code",[e._v("02-client")]),e._v(" submodule.")]),e._v(" "),l("h2",{attrs:{id:"configuring-a-light-client-module"}},[l("a",{staticClass:"header-anchor",attrs:{href:"#configuring-a-light-client-module"}},[e._v("#")]),e._v(" Configuring a light client module")]),e._v(" "),l("p",[e._v("An IBC light client module must implement the "),l("a",{attrs:{href:"https://github.com/cosmos/cosmos-sdk/blob/main/types/module/module.go#L50",target:"_blank",rel:"noopener noreferrer"}},[l("code",[e._v("AppModuleBasic")]),l("OutboundLink")],1),e._v(" interface in order to register its concrete types against the core IBC interfaces defined in "),l("code",[e._v("modules/core/exported")]),e._v(". This is accomplished via the "),l("code",[e._v("RegisterInterfaces")]),e._v(" method which provides the light client module with the opportunity to register codec types using the chain's "),l("code",[e._v("InterfaceRegistery")]),e._v(". Please refer to the "),l("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/02-client-refactor-beta1/modules/light-clients/07-tendermint/codec.go#L11",target:"_blank",rel:"noopener noreferrer"}},[l("code",[e._v("07-tendermint")]),e._v(" codec registration"),l("OutboundLink")],1),e._v(".")]),e._v(" "),l("p",[e._v("The "),l("code",[e._v("AppModuleBasic")]),e._v(" interface may also be leveraged to install custom CLI handlers for light client module users. Light client modules can safely no-op for interface methods which it does not wish to implement.")]),e._v(" "),l("p",[e._v("Please refer to the "),l("RouterLink",{attrs:{to:"/ibc/integration.html#integrating-light-clients"}},[e._v("core IBC documentation")]),e._v(" for how to configure additional light client modules alongside "),l("code",[e._v("07-tendermint")]),e._v(" in "),l("code",[e._v("app.go")]),e._v(".")],1),e._v(" "),l("p",[e._v("See below for an example of the "),l("code",[e._v("07-tendermint")]),e._v(" implementation of "),l("code",[e._v("AppModuleBasic")]),e._v(".")]),e._v(" "),l("tm-code-block",{staticClass:"codeblock",attrs:{language:"go",base64:"dmFyIF8gbW9kdWxlLkFwcE1vZHVsZUJhc2ljID0gQXBwTW9kdWxlQmFzaWN7fQoKLy8gQXBwTW9kdWxlQmFzaWMgZGVmaW5lcyB0aGUgYmFzaWMgYXBwbGljYXRpb24gbW9kdWxlIHVzZWQgYnkgdGhlIHRlbmRlcm1pbnQgbGlnaHQgY2xpZW50LgovLyBPbmx5IHRoZSBSZWdpc3RlckludGVyZmFjZXMgZnVuY3Rpb24gbmVlZHMgdG8gYmUgaW1wbGVtZW50ZWQuIEFsbCBvdGhlciBmdW5jdGlvbiBwZXJmb3JtCi8vIGEgbm8tb3AuCnR5cGUgQXBwTW9kdWxlQmFzaWMgc3RydWN0e30KCi8vIE5hbWUgcmV0dXJucyB0aGUgdGVuZGVybWludCBtb2R1bGUgbmFtZS4KZnVuYyAoQXBwTW9kdWxlQmFzaWMpIE5hbWUoKSBzdHJpbmcgewoJcmV0dXJuIE1vZHVsZU5hbWUKfQoKLy8gUmVnaXN0ZXJMZWdhY3lBbWlub0NvZGVjIHBlcmZvcm1zIGEgbm8tb3AuIFRoZSBUZW5kZXJtaW50IGNsaWVudCBkb2VzIG5vdCBzdXBwb3J0IGFtaW5vLgpmdW5jIChBcHBNb2R1bGVCYXNpYykgUmVnaXN0ZXJMZWdhY3lBbWlub0NvZGVjKCpjb2RlYy5MZWdhY3lBbWlubykge30KCi8vIFJlZ2lzdGVySW50ZXJmYWNlcyByZWdpc3RlcnMgbW9kdWxlIGNvbmNyZXRlIHR5cGVzIGludG8gcHJvdG9idWYgQW55LiBUaGlzIGFsbG93cyBjb3JlIElCQwovLyB0byB1bm1hcnNoYWwgdGVuZGVybWludCBsaWdodCBjbGllbnQgdHlwZXMuCmZ1bmMgKEFwcE1vZHVsZUJhc2ljKSBSZWdpc3RlckludGVyZmFjZXMocmVnaXN0cnkgY29kZWN0eXBlcy5JbnRlcmZhY2VSZWdpc3RyeSkgewoJUmVnaXN0ZXJJbnRlcmZhY2VzKHJlZ2lzdHJ5KQp9CgovLyBEZWZhdWx0R2VuZXNpcyBwZXJmb3JtcyBhIG5vLW9wLiBHZW5lc2lzIGlzIG5vdCBzdXBwb3J0ZWQgZm9yIHRoZSB0ZW5kZXJtaW50IGxpZ2h0IGNsaWVudC4KZnVuYyAoQXBwTW9kdWxlQmFzaWMpIERlZmF1bHRHZW5lc2lzKGNkYyBjb2RlYy5KU09OQ29kZWMpIGpzb24uUmF3TWVzc2FnZSB7CglyZXR1cm4gbmlsCn0KCi8vIFZhbGlkYXRlR2VuZXNpcyBwZXJmb3JtcyBhIG5vLW9wLiBHZW5lc2lzIGlzIG5vdCBzdXBwb3J0ZWQgZm9yIHRoZSB0ZW5kZXJtaW50IGxpZ2h0IGNpbGVudC4KZnVuYyAoQXBwTW9kdWxlQmFzaWMpIFZhbGlkYXRlR2VuZXNpcyhjZGMgY29kZWMuSlNPTkNvZGVjLCBjb25maWcgY2xpZW50LlR4RW5jb2RpbmdDb25maWcsIGJ6IGpzb24uUmF3TWVzc2FnZSkgZXJyb3IgewoJcmV0dXJuIG5pbAp9CgovLyBSZWdpc3RlckdSUENHYXRld2F5Um91dGVzIHBlcmZvcm1zIGEgbm8tb3AuCmZ1bmMgKEFwcE1vZHVsZUJhc2ljKSBSZWdpc3RlckdSUENHYXRld2F5Um91dGVzKGNsaWVudEN0eCBjbGllbnQuQ29udGV4dCwgbXV4ICpydW50aW1lLlNlcnZlTXV4KSB7fQoKLy8gR2V0VHhDbWQgcGVyZm9ybXMgYSBuby1vcC4gUGxlYXNlIHNlZSB0aGUgMDItY2xpZW50IGNsaSBjb21tYW5kcy4KZnVuYyAoQXBwTW9kdWxlQmFzaWMpIEdldFR4Q21kKCkgKmNvYnJhLkNvbW1hbmQgewoJcmV0dXJuIG5pbAp9CgovLyBHZXRRdWVyeUNtZCBwZXJmb3JtcyBhIG5vLW9wLiBQbGVhc2Ugc2VlIHRoZSAwMi1jbGllbnQgY2xpIGNvbW1hbmRzLgpmdW5jIChBcHBNb2R1bGVCYXNpYykgR2V0UXVlcnlDbWQoKSAqY29icmEuQ29tbWFuZCB7CglyZXR1cm4gbmlsCn0K"}}),e._v(" "),l("h2",{attrs:{id:"creating-clients"}},[l("a",{staticClass:"header-anchor",attrs:{href:"#creating-clients"}},[e._v("#")]),e._v(" Creating clients")]),e._v(" "),l("p",[e._v("A client is created by executing a new "),l("code",[e._v("MsgCreateClient")]),e._v(" transaction composed with a valid "),l("code",[e._v("ClientState")]),e._v(" and initial "),l("code",[e._v("ConsensusState")]),e._v(" encoded as protobuf "),l("code",[e._v("Any")]),e._v("s.\nGenerally, this is normally done by an off-chain process known as an "),l("a",{attrs:{href:"https://github.com/cosmos/ibc/tree/main/spec/relayer/ics-018-relayer-algorithms",target:"_blank",rel:"noopener noreferrer"}},[e._v("IBC relayer"),l("OutboundLink")],1),e._v(" however, this is not a strict requirement.")]),e._v(" "),l("p",[e._v("See below for a list of IBC relayer implementations:")]),e._v(" "),l("ul",[l("li",[l("a",{attrs:{href:"https://github.com/cosmos/relayer",target:"_blank",rel:"noopener noreferrer"}},[e._v("cosmos/relayer"),l("OutboundLink")],1)]),e._v(" "),l("li",[l("a",{attrs:{href:"https://github.com/informalsystems/hermes",target:"_blank",rel:"noopener noreferrer"}},[e._v("informalsystems/hermes"),l("OutboundLink")],1)]),e._v(" "),l("li",[l("a",{attrs:{href:"https://github.com/confio/ts-relayer",target:"_blank",rel:"noopener noreferrer"}},[e._v("confio/ts-relayer"),l("OutboundLink")],1)])]),e._v(" "),l("p",[e._v("Stateless checks are performed within the "),l("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/02-client-refactor-beta1/modules/core/02-client/types/msgs.go#L48",target:"_blank",rel:"noopener noreferrer"}},[l("code",[e._v("ValidateBasic")]),l("OutboundLink")],1),e._v(" method of "),l("code",[e._v("MsgCreateClient")]),e._v(".")]),e._v(" "),l("tm-code-block",{staticClass:"codeblock",attrs:{language:"protobuf",base64:"Ly8gTXNnQ3JlYXRlQ2xpZW50IGRlZmluZXMgYSBtZXNzYWdlIHRvIGNyZWF0ZSBhbiBJQkMgY2xpZW50Cm1lc3NhZ2UgTXNnQ3JlYXRlQ2xpZW50IHsKICBvcHRpb24gKGdvZ29wcm90by5lcXVhbCkgICAgICAgICAgID0gZmFsc2U7CiAgb3B0aW9uIChnb2dvcHJvdG8uZ29wcm90b19nZXR0ZXJzKSA9IGZhbHNlOwoKICAvLyBsaWdodCBjbGllbnQgc3RhdGUKICBnb29nbGUucHJvdG9idWYuQW55IGNsaWVudF9zdGF0ZSA9IDEgWyhnb2dvcHJvdG8ubW9yZXRhZ3MpID0gJnF1b3Q7eWFtbDpcJnF1b3Q7Y2xpZW50X3N0YXRlXCZxdW90OyZxdW90O107CiAgLy8gY29uc2Vuc3VzIHN0YXRlIGFzc29jaWF0ZWQgd2l0aCB0aGUgY2xpZW50IHRoYXQgY29ycmVzcG9uZHMgdG8gYSBnaXZlbgogIC8vIGhlaWdodC4KICBnb29nbGUucHJvdG9idWYuQW55IGNvbnNlbnN1c19zdGF0ZSA9IDIgWyhnb2dvcHJvdG8ubW9yZXRhZ3MpID0gJnF1b3Q7eWFtbDpcJnF1b3Q7Y29uc2Vuc3VzX3N0YXRlXCZxdW90OyZxdW90O107CiAgLy8gc2lnbmVyIGFkZHJlc3MKICBzdHJpbmcgc2lnbmVyID0gMzsKfQo="}}),e._v(" "),l("p",[e._v("Leveraging protobuf "),l("code",[e._v("Any")]),e._v(" encoding allows core IBC to "),l("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/02-client-refactor-beta1/modules/core/keeper/msg_server.go#L28-L36",target:"_blank",rel:"noopener noreferrer"}},[e._v("unpack"),l("OutboundLink")],1),e._v(" both the "),l("code",[e._v("ClientState")]),e._v(" and "),l("code",[e._v("ConsensusState")]),e._v(" into their respective interface types registered previously using the light client module's "),l("code",[e._v("RegisterInterfaces")]),e._v(" method.")]),e._v(" "),l("p",[e._v("Within the "),l("code",[e._v("02-client")]),e._v(" submodule, the "),l("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/02-client-refactor-beta1/modules/core/02-client/keeper/client.go#L30-L34",target:"_blank",rel:"noopener noreferrer"}},[l("code",[e._v("ClientState")]),e._v(" is then initialized"),l("OutboundLink")],1),e._v(" with its own isolated key-value store, namespaced using a unique client identifier.")]),e._v(" "),l("p",[e._v("In order to successfully create an IBC client using a new client type it "),l("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/02-client-refactor-beta1/modules/core/02-client/keeper/client.go#L18-L24",target:"_blank",rel:"noopener noreferrer"}},[e._v("must be supported"),l("OutboundLink")],1),e._v(". Light client support in IBC is gated by on-chain governance. The allow list may be updated by submitting a new governance proposal to update the "),l("code",[e._v("02-client")]),e._v(" parameter "),l("code",[e._v("AllowedClients")]),e._v(".")]),e._v(" "),l("p",[e._v("See below for example:")]),e._v(" "),l("tm-code-block",{staticClass:"codeblock",attrs:{language:"",base64:"JCAlcyB0eCBnb3Ygc3VibWl0LXByb3Bvc2FsIHBhcmFtLWNoYW5nZSAmbHQ7cGF0aC90by9wcm9wb3NhbC5qc29uJmd0OyAtLWZyb209Jmx0O2tleV9vcl9hZGRyZXNzJmd0OwpXaGVyZSBwcm9wb3NhbC5qc29uIGNvbnRhaW5zOgp7CiAgJnF1b3Q7dGl0bGUmcXVvdDs6ICZxdW90O0lCQyBDbGllbnRzIFBhcmFtIENoYW5nZSZxdW90OywKICAmcXVvdDtkZXNjcmlwdGlvbiZxdW90OzogJnF1b3Q7VXBkYXRlIGFsbG93ZWQgY2xpZW50cyZxdW90OywKICAmcXVvdDtjaGFuZ2VzJnF1b3Q7OiBbCiAgICB7CiAgICAgICZxdW90O3N1YnNwYWNlJnF1b3Q7OiAmcXVvdDtpYmMmcXVvdDssCiAgICAgICZxdW90O2tleSZxdW90OzogJnF1b3Q7QWxsb3dlZENsaWVudHMmcXVvdDssCiAgICAgICZxdW90O3ZhbHVlJnF1b3Q7OiBbJnF1b3Q7MDYtc29sb21hY2hpbmUmcXVvdDssICZxdW90OzA3LXRlbmRlcm1pbnQmcXVvdDssICZxdW90OzB4LW5ldy1jbGllbnQmcXVvdDtdCiAgICB9CiAgXSwKICAmcXVvdDtkZXBvc2l0JnF1b3Q7OiAmcXVvdDsxMDAwc3Rha2UmcXVvdDsKfQo="}})],1)}),[],!1,null,null,null);t.default=c.exports}}]);