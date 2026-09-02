package sheets

import "context"

// Noop — Google Sheets is optional sync/export only. No Apps Script.
// SQLite remains the source of truth even when this adapter is enabled later.
type Noop struct {
	on bool
}

func New(enabled bool) *Noop { return &Noop{on: enabled} }

func (n *Noop) Enabled() bool { return n.on }

func (n *Noop) Sync(context.Context) error {
	if !n.on {
		return nil
	}
	return nil
}
