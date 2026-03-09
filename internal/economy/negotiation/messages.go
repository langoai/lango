package negotiation

import (
	"encoding/json"
	"time"
)

// ProposalAction is the action type in a proposal.
type ProposalAction string

const (
	ActionPropose ProposalAction = "propose"
	ActionCounter ProposalAction = "counter"
	ActionAccept  ProposalAction = "accept"
	ActionReject  ProposalAction = "reject"
)

// Proposal is a single offer or counter-offer in a negotiation.
type Proposal struct {
	Action    ProposalAction `json:"action"`
	SenderDID string         `json:"senderDid"`
	Terms     Terms          `json:"terms"`
	Round     int            `json:"round"`
	Reason    string         `json:"reason,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// NegotiatePayload is the P2P message payload for negotiation.
type NegotiatePayload struct {
	SessionID string   `json:"sessionId"`
	Proposal  Proposal `json:"proposal"`
}

// Marshal serializes NegotiatePayload to JSON.
func (np *NegotiatePayload) Marshal() ([]byte, error) {
	return json.Marshal(np)
}

// UnmarshalNegotiatePayload deserializes from JSON.
func UnmarshalNegotiatePayload(data []byte) (*NegotiatePayload, error) {
	var np NegotiatePayload
	if err := json.Unmarshal(data, &np); err != nil {
		return nil, err
	}
	return &np, nil
}
