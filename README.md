# Quards

Modular state-space explorer in Python.

# Quards Structure

Quards is a state explorer that will iterate through every possible option
for a state. It is very agnostic to the contents (see #evaluators below). It
is functionally a graph generator. There's a few components that are important
for the engine to run, but must be agnostic to the evaluator.

## State

Handles the state of the world, generally after an action is executed. In
technical graph terms, a state is a node. Nodes are created after an action

## Action

This is an action that can be taken. Actions are added as incomplete, and are
processed separately. This results in graph edgeds that are unresolved.

When an action is evaluated, the game state MUST change, otherwise it's not a
relevant action. The new state (node) is created and the action is marked as
pointing to the new state (resolved)

## Explorer

Manages the turn depth, iteration and creating new actions.

# Evaluators

Evaluators should be responsible for handling two methods.

def get_actions(state_data) -> [ (name, params)... ]
def execute(state_data, action, params) -> new_state_data

These methods allow the state and quards engine to parse through all of the
possibilities and scale independantly. This is a very hard boundary.

Evaluators should NEVER import state/actions and be considered functional. This
will allow us to create easily mockable tests based on simple types. Only dumb
dictionaries should be passed around at this level.

## Lorcana Evaluator

TODO: Fill out this second
def get_initial_state -> state_data
def get_turn_summary -> ??

# Usage

```
git clone git@github.com:kalebpomeroy/quards.git
cd quards
python -m venv .venv
. .venv/bin/activate
pip install -r requirements.txt
python main.py
```
