package output

// Slim structs used exclusively for --agent output. Every field has omitempty
// so absent/zero values are never emitted. Treat the exported field set as a
// stable contract — adding fields is fine, removing/renaming is breaking.

// agentSubscription is the whitelisted view of a single subscription record.
type agentSubscription struct {
	ID            int64           `json:"id"`
	Status        string          `json:"status,omitempty"`
	CustomerEmail string          `json:"customer_email,omitempty"`
	CustomerName  string          `json:"customer_name,omitempty"`
	Currency      string          `json:"currency,omitempty"`
	Total         string          `json:"total,omitempty"`
	Interval      string          `json:"interval,omitempty"`
	Address       string          `json:"address,omitempty"`
	CardSummary   string          `json:"card,omitempty"`
	Items         []agentItem     `json:"items,omitempty"`
	NextAttempts  []agentAttempt  `json:"next_attempts,omitempty"`  // auto_charge payment type
	Invoices      []agentInvoice  `json:"invoices,omitempty"`        // recurring_invoice payment type
}

// agentListItem is the slimmer per-row view used in list NDJSON output.
type agentListItem struct {
	ID            int64  `json:"id"`
	Status        string `json:"status,omitempty"`
	CustomerEmail string `json:"customer_email,omitempty"`
	Currency      string `json:"currency,omitempty"`
	Total         string `json:"total,omitempty"`
	Interval      string `json:"interval,omitempty"`
}

// agentItem is one subscription line item.
type agentItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title,omitempty"`
	Qty   int    `json:"qty,omitempty"`
	Price string `json:"price,omitempty"`
}

// agentAttempt is a billing attempt entry (auto_charge payment type).
type agentAttempt struct {
	ID     int64  `json:"id"`
	Date   string `json:"date,omitempty"`
	Status string `json:"status,omitempty"`
}

// agentInvoice is an invoice entry (recurring_invoice payment type).
type agentInvoice struct {
	ID            int64  `json:"id"`
	Date          string `json:"date,omitempty"`
	PaymentStatus string `json:"payment_status,omitempty"`
}

// agentMeta is the final line of a list NDJSON stream.
type agentMeta struct {
	Meta struct {
		Page    int  `json:"page"`
		HasMore bool `json:"has_more"`
		Count   int  `json:"count"`
	} `json:"_meta"`
}

// agentSuccessAck is the mutation acknowledgement.
type agentSuccessAck struct {
	Ok     bool   `json:"ok"`
	ID     int    `json:"id"`
	Action string `json:"action,omitempty"`
}

// agentErrorAck is the error envelope written on non-2xx in --agent mode.
type agentErrorAck struct {
	Ok     bool   `json:"ok"`
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
	Detail string `json:"detail,omitempty"`
}
