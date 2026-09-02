package transcript

import (
	"context"
	"fmt"
)

// Stub — transcript provider is unresolved in the architecture map
// (Supadata requested, GoClaw current). MVP returns a local note.
type Stub struct{}

func New() *Stub { return &Stub{} }

func (s *Stub) Transcript(_ context.Context, youtubeVideoID string) (string, error) {
	return fmt.Sprintf("[transcript stub] video %s — provider undecided (Supadata requested / GoClaw current). Business logic still owns outlier scoring without captions.", youtubeVideoID), nil
}
