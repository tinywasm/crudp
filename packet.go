package crudp

// Packet represents both requests and responses of the protocol
type Packet struct {
	Action    byte     `json:"action"`
	HandlerID uint8    `json:"handler_id"`
	ReqID     string   `json:"req_id"`
	Data      [][]byte `json:"data"`
}

// BatchRequest is what is sent in the POST /sync
type BatchRequest struct {
	Packets []Packet `json:"packets"`
}

// BatchResponse is what is received by SSE or as HTTP response
type BatchResponse struct {
	Results []PacketResult `json:"results"`
}

type PacketResult struct {
	Packet             // Embed Packet complete for symmetry with BatchRequest
	MessageType uint8  `json:"message_type"` // 0=Normal, 1=Info, 2=Error, 3=Warning, 4=Success
	Message     string `json:"message"`      // Message for the user
}

// Request represents a single operation request for automatic endpoints
type Request struct {
	ReqID string   `json:"req_id"`
	Data  [][]byte `json:"data"`
}

// Response represents a single operation response for automatic endpoints
type Response struct {
	ReqID       string   `json:"req_id"`
	Data        [][]byte `json:"data"`
	MessageType uint8    `json:"message_type"`
	Message     string   `json:"message"`
}
