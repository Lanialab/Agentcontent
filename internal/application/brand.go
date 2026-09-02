package application

import (
	"context"
	"time"

	"github.com/grokbot-2/agentcontent/internal/domain/brand"
	"github.com/grokbot-2/agentcontent/internal/ports"
)

type BrandService struct {
	store ports.BrandStore
	trace ports.Tracer
}

func NewBrandService(store ports.BrandStore, trace ports.Tracer) *BrandService {
	return &BrandService{store: store, trace: trace}
}

func (s *BrandService) Get(ctx context.Context) (brand.Blueprint, error) {
	s.pulse("get-blueprint")
	return s.store.GetBlueprint(ctx)
}

func (s *BrandService) Save(ctx context.Context, bp brand.Blueprint) (brand.Blueprint, error) {
	bp.UpdatedAt = time.Now().UTC()
	s.pulse("save-blueprint")
	if s.trace != nil {
		s.trace.Emit(ports.TraceEvent{
			Path:  "brand",
			Kind:  ports.TracePersist,
			Label: "save-blueprint",
			Nodes: []string{"uc-brand", "port-brand-store", "adp-sqlite"},
		})
	}
	if err := s.store.SaveBlueprint(ctx, bp); err != nil {
		return brand.Blueprint{}, err
	}
	return bp, nil
}

func (s *BrandService) pulse(label string) {
	if s.trace == nil {
		return
	}
	s.trace.Emit(ports.TraceEvent{
		Path:  "brand",
		Kind:  ports.TraceCommand,
		Label: label,
		Nodes: []string{"ui-brand", "http-brand", "port-brand", "uc-brand", "dom-brand", "port-brand-store", "adp-sqlite"},
	})
}
