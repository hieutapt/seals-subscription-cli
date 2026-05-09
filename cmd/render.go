package cmd

import (
	"fmt"

	"github.com/hieutapt/seals-subscription-cli/internal/output"
)

// renderSubscriptionDetail dispatches to the correct renderer based on active output mode.
func renderSubscriptionDetail(data []byte, jsonOut, agentOut bool) error {
	switch {
	case agentOut:
		return output.AgentSubscriptionDetail(data)
	case jsonOut:
		return output.JSON(data)
	default:
		return output.SubscriptionDetail(data)
	}
}

// renderSubscriptionList dispatches list rendering.
func renderSubscriptionList(data []byte, jsonOut, agentOut bool) error {
	switch {
	case agentOut:
		return output.AgentSubscriptionList(data)
	case jsonOut:
		return output.JSON(data)
	default:
		return output.SubscriptionList(data)
	}
}

// renderSuccess dispatches mutation success output.
// action is a short verb ("cancel", "pause", etc.) used in agent acks.
// id is the subscription/attempt ID for the agent ack.
// fallback is the human-readable ✓ message.
func renderSuccess(data []byte, jsonOut, agentOut bool, action string, id int, fallback string) error {
	switch {
	case agentOut:
		return output.AgentSuccess(data, action, id)
	case jsonOut:
		return output.JSON(data)
	default:
		return output.Success(data, fallback)
	}
}

// renderError is called when the API returns an error in agent mode — it writes
// the tee-on-failure envelope to stderr and returns the original error.
func renderAgentError(body []byte, statusCode int) {
	envelope := output.AgentError(body, statusCode)
	fmt.Printf("%s\n", envelope)
}
