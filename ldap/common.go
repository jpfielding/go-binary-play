package main

import "net"

type ProtocolOp string

// https://en.wikipedia.org/wiki/Lightweight_Directory_Access_Protocol#Search_and_compare
// LDAP Requests https://datatracker.ietf.org/doc/html/rfc4511
const (
	BindRequest     ProtocolOp = "BindRequest"
	BindResponse    ProtocolOp = "BindResponse"
	UnbindRequest   ProtocolOp = "UnbindRequest" // kill the connection (not de-auth)
	SearchRequest   ProtocolOp = "SearchRequest"
	SearchResEntry  ProtocolOp = "SearchResultEntry"
	SearchResDone   ProtocolOp = "SearchResultDone"
	SearchResRef    ProtocolOp = "SearchResultReference"
	ModifyRequest   ProtocolOp = "ModifyRequest"
	ModifyResponse  ProtocolOp = "ModifyResponse"
	AddRequest      ProtocolOp = "AddRequest"
	AddResponse     ProtocolOp = "AddResponse"
	DelRequest      ProtocolOp = "DelRequest"
	DelResponse     ProtocolOp = "DelResponse"
	DodDNRequest    ProtocolOp = "ModifyDNRequest"
	ModDNResponse   ProtocolOp = "ModifyDNResponse"
	CompareRequest  ProtocolOp = "CompareRequest"
	CompareResponse ProtocolOp = "CompareResponse"
	AbandonRequest  ProtocolOp = "AbandonRequest"
	ExtendedReq     ProtocolOp = "ExtendedRequest" // StartTLS, Cancel, Password Modify
	ExtendedResp    ProtocolOp = "ExtendedResponse"
)

type Message struct {
	MessageID  int64
	ProtocolOp ProtocolOp
}

type Session struct {
}

// Serve the connections requests
func (s *Session) Serve(l net.Conn) error {
	defer l.Close()
	return nil
}
