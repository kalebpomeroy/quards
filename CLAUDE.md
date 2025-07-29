# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

**Current Development**: New implementation in Go
**Legacy Code**: The `python/` directory contains the original prototype and should be treated as reference material only

## Development Commands

For new Go development:
```bash
go mod init quards
go mod tidy
go run main.go
```

### Legacy Python Reference

The Python prototype can be referenced for understanding the architecture:
```bash
cd python/
python -m venv .venv
. .venv/bin/activate  
pip install -r requirements.txt
python main.py  # Reference implementation
```

## Architecture Overview (from Python prototype)

Quards is a modular state-space explorer that generates a complete game tree by exploring every possible action sequence. The Go implementation should follow these core architectural principles:

### Core Components to Implement in Go

**State Machine Engine**:
- **State**: Game state nodes with deterministic signatures for deduplication
- **Action**: Edges between states, stored as pending until resolved  
- **Explorer**: Breadth-first exploration across turn depths

**Evaluators**:
- Game-agnostic interface requiring `GetActions(stateData)` and `Execute(stateData, action, params)`
- Lorcana evaluator for Disney Lorcana card game rules

### Key Design Patterns from Prototype

- **Functional Evaluators**: Work with plain data structures, avoid tight coupling
- **Lazy State Loading**: Load states from storage only when needed
- **Breadth-First Exploration**: Process all actions at depth N before moving to N+1
- **Pruning Logic**: Detect and terminate bad move branches early
- **Database Storage**: States and actions persisted with PostgreSQL

### Go Implementation Notes

- Use interfaces for evaluators to maintain modularity
- Consider using Go's concurrency features for parallel exploration
- Implement proper error handling for database operations
- Use structured logging for exploration progress
- JSON serialization for state data compatibility