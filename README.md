# Your Project

Modular state-space explorer in Python.

# Evaluator

This is the game specific evaluator. It should:

- Start from existing state (json definition?)
- State includes a list of valid selections
- Given a state

# State Generator

This should generate a state from a json file.

# Explorer

This takes a state, and evaluates every choice, resulting in many different states that can each be evaluated

# Flow

- On start-job

  - Creates a valid starting state and seed ID
  - Create an edge/action for the seed-state to "start game"

- While true:
  - Explorer looks for any "open" actions (which include a seed/state/start + action)
  - Explorer asks the evaluator to execute a single action, resulting in:
    - New state created (if necessary)
    - Mark the action as "closed"
