// Package protocol holds the wire types shared by discovery, server and client.
// See PROTOCOL.md at the repository root for the full description.
package protocol

const (
	Version = 1

	// DiscoveryPort is the UDP port used for announce broadcasts.
	DiscoveryPort = 47820
	// DefaultTransferPort is the TCP port of the HTTPS transfer server.
	DefaultTransferPort = 47821
	// MulticastGroup is a fallback discovery channel next to directed broadcast.
	MulticastGroup = "239.255.47.82"
)

// Announce is the UDP discovery datagram (JSON encoded).
type Announce struct {
	Magic       string `json:"magic"` // always "patchdrop"
	V           int    `json:"v"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	OS          string `json:"os"`
	Port        int    `json:"port"`
	Fingerprint string `json:"fp"`
	Link        string `json:"link,omitempty"` // "wired", "wifi" or "" – interface the datagram was sent on
	Reply       bool   `json:"reply,omitempty"`
	Bye         bool   `json:"bye,omitempty"` // sent on shutdown
}

const Magic = "patchdrop"

// Info is returned by GET /api/info.
type Info struct {
	V           int    `json:"v"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	OS          string `json:"os"`
	Port        int    `json:"port"`
	Fingerprint string `json:"fp"`
}

// FileMeta describes one file in an offer. Name is a slash-separated relative path.
type FileMeta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// Offer is the body of POST /api/offer.
type Offer struct {
	Sender Info       `json:"sender"`
	Files  []FileMeta `json:"files"`
}

// OfferResponse is returned when an offer is accepted.
type OfferResponse struct {
	Session string            `json:"session"`
	Tokens  map[string]string `json:"tokens"` // file id -> upload token
}
