package lens

import (
	"fmt"
	"quards/internal/lens/core"
	"quards/internal/lens/services"
	"quards/internal/parser"
)

// LensFunc represents a lens function signature
type LensFunc func(entries []parser.LogEntry, services *services.LensServices) interface{}

// Processor orchestrates lens functions with their dependencies
type Processor struct {
	services *services.LensServices
	lenses   map[string]LensFunc
}

// New creates a lens processor with default services
func New() *Processor {
	cardDB := services.NewInMemoryCardDB()
	if err := cardDB.LoadFromFile("static/cards.lorcana-api.json"); err != nil {
		// In production, you'd want proper error handling here
		_ = err
	}

	return NewWithServices(&services.LensServices{
		CardDB: cardDB,
		Cache:  services.NewInMemoryCache(),
	})
}

// NewWithServices creates a processor with custom services
func NewWithServices(svc *services.LensServices) *Processor {
	return &Processor{
		services: svc,
		lenses: map[string]LensFunc{
			"zones":            core.ZonesLens,
			"playerStats":      core.PlayerStatsLens,
			"availableActions": core.AvailableActionsLens,
			"gameSteps":        core.GameStepsLens,
			"battlefield":      core.BattlefieldLens,
			"composite":        core.CompositeGameStateLens,
			"gameState":        core.GameStateLens,
			"history":          core.HistoryLens,
			"stepsNavigation":  core.StepsNavigationLens,
		},
	}
}

// Lens executes a lens by name
func (p *Processor) Lens(name string, entries []parser.LogEntry) (interface{}, error) {
	lensFunc, exists := p.lenses[name]
	if !exists {
		return nil, fmt.Errorf("lens '%s' not found", name)
	}
	return lensFunc(entries, p.services), nil
}

// LensFromContent executes a lens directly from log content
func (p *Processor) LensFromContent(name string, content string) (interface{}, error) {
	entries, err := parser.ParseLogContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log: %w", err)
	}
	return p.Lens(name, entries)
}

// RegisterLens adds a custom lens
func (p *Processor) RegisterLens(name string, lensFunc LensFunc) {
	p.lenses[name] = lensFunc
}

// AvailableLenses returns all lens names
func (p *Processor) AvailableLenses() []string {
	names := make([]string, 0, len(p.lenses))
	for name := range p.lenses {
		names = append(names, name)
	}
	return names
}

// Services returns the underlying services
func (p *Processor) Services() *services.LensServices {
	return p.services
}

// CacheStats returns cache statistics
func (p *Processor) CacheStats() map[string]interface{} {
	return p.services.Cache.GetStats()
}

// ClearCache clears all cached data
func (p *Processor) ClearCache() {
	p.services.Cache.Clear()
}