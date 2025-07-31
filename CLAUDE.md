# CLAUDE.md

This file provides guidance to Claude Code for working with the Quards codebase.

## Project Overview

**Quards** is a Disney Lorcana card game replay and analysis platform built in Go.

**Current State**: Functional game viewer with event-sourcing architecture
**Active Development**: Game state management, card drawing mechanics, UI improvements

## Quick Start

```bash
# Development server (with live reload)
air  # Uses .air.toml config, watches files, auto-rebuilds

# Manual run
go run main.go

# Server runs on http://localhost:8080
# Game viewer: http://localhost:8080/nav.html
```

## Architecture

### Core Components

- **Event-Sourcing**: Games stored as append-only logs of player actions
- **Lens System**: Pure functional data transformers that compute game state from logs
- **API Layer**: REST endpoints for game management and state queries
- **Web UI**: Static HTML/JS for game replay and analysis

### Directory Structure

```
internal/
├── api/           # REST API handlers and routing
├── game/          # Game logic, log management, turn advancement  
├── lens/          # State computation system
│   ├── core/      # Core lens implementations
│   └── services/  # Card database, caching
├── parser/        # Log parsing and event structures
├── deck/          # Deck management
└── database/      # PostgreSQL integration

static/            # Web UI (HTML/CSS/JS)
python/            # Legacy prototype (reference only)
```

## Key Systems

### 1. Event Sourcing & Game Logs

Games are stored as logs like:
```
GameStarted p1_deck="YR Pile" p2_deck="UG Something" seed=12345
DecksShuffled seed=12345
OpeningHandsDrawn p1="card1,card2,..." p2="card3,card4,..."
TurnStarted player=1 turn=1
CardInked player=1 card_id="AZS-025"
CardPlayed player=1 card_id="AZS-034" cost=1
TurnPassed player=1
CardDrawn player=2 card_id="INK-142"  # Auto-generated on turn start
TurnStarted player=2 turn=2
```

**Important**: String parameters are quoted to handle spaces in deck names.

### 2. Lens System

Lenses are pure functions: `([]LogEntry, services) -> interface{}`

Key lenses:
- `zones`: Player hands, battlefield, ink, deck/discard counts
- `history`: Human-readable event descriptions with card names  
- `stepsNavigation`: UI-optimized step data for replay
- `availableActions`: Valid actions for current player
- `composite`: Combined game state for UI

### 3. Turn Management & Card Drawing

**Rules**:
- Player 1 doesn't draw on turn 1 (opening hand only)
- Player 2 draws on all turns (including turn 1)  
- Player 1 draws on turns 2, 3, etc.

**Implementation**: `AppendActionToGameByID()` automatically handles turn advancement and card drawing.

### 4. API Endpoints

```
# Game Management
GET    /api/games                    # List games
POST   /api/games                    # Create game
GET    /api/games/{id}               # Get game details
DELETE /api/games/{id}               # Delete game
POST   /api/games/{id}/execute       # Execute action
POST   /api/games/{id}/truncate      # Truncate to specific log

# Game State (via lenses)
GET    /api/games/{id}/state         # Full composite state
GET    /api/games/{id}/history       # Human-readable history
GET    /api/games/{id}/navigation    # UI navigation data
GET    /api/games/{id}/actions       # Available actions

# Deck Management  
GET    /api/decks                    # List decks
GET    /api/decks/{deckname}         # Get deck (uses name as ID)
```

## Current Issues & TODO

### Known Issues
1. **Card Drawing Bug**: `getNextCardFromDeck()` fails with quoted deck names
   - Root cause: Log parser splits quoted strings incorrectly
   - Need to fix `parseLogLine()` in `internal/parser/parser.go`

2. **History UI Lag**: Client-side history rendering has timing issues
   - Server-side lenses work correctly
   - UI needs debugging for step synchronization

### Immediate TODOs
- [ ] Fix log parser to handle quoted parameters properly
- [ ] Switch deck references from names to IDs (game creation, UI)
- [ ] Add comprehensive API test suite (external to Go tests)
- [ ] UI improvements for game replay

## Development Notes

### Database
- PostgreSQL required
- Game and deck storage
- Connection initialized in `database/` package

### Card Data
- Cards loaded from `static/cards.lorcana-api.json`
- Card database provides name lookups for IDs
- Used in history lens for human-readable descriptions

### UI Structure
- `nav.html`: Game list and navigation
- `index.html`: Game viewer (replay interface)
- `app.js`: Main game viewer logic
- Uses server-side lenses for data (no client-side game logic)

### Testing
- No traditional Go tests (by design)
- External API test suite planned
- Manual testing via web UI during development

## Common Patterns

### Adding New Lenses
1. Create lens function in `internal/lens/core/`
2. Register in `internal/lens/processor.go`
3. Add API endpoint in `internal/api/handlers.go`
4. Update UI to consume new endpoint

### Debugging Game Issues
1. Check raw log content: `GET /api/games/{id}` 
2. Test lens output: `GET /api/games/{id}/{lensName}`
3. Use browser dev tools for UI debugging
4. Server logs in `tmp/server.log` (with Air)

### Log Format
- Event names are PascalCase: `CardPlayed`, `TurnStarted`
- Parameters are snake_case: `card_id`, `player_deck`
- String values are quoted: `deck="YR Pile"`
- Parsing handles both quoted and unquoted values

## Project Philosophy

- **Functional approach**: Lenses are pure functions, no side effects
- **Event sourcing**: All game state derived from logs, no mutations
- **API-first**: UI consumes server-provided data, minimal client logic
- **Simple tech stack**: Go backend, vanilla JS frontend, PostgreSQL storage