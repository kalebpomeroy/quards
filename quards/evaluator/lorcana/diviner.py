from quards.evaluator.lorcana import cards


def list_actions(state):

    player = state["current_player"]

    # Set a fake victory condition:
    if len(state["zones"]["ink"][player]) > 1:
        print(f"{player} WON BY HAVING 4 ink in play")
        return []

    actions = []
    if state["ink_drops_available"] > 0:
        for card_id in state["zones"]["hands"][player]:
            card = cards.CardIndex().get_by_id(card_id)
            if card["Inkable"]:
                actions.append(("ink", {"card_id": card_id, "player": player}))

    # This needs to be more sophisticated. This forces the player to take every
    # action before they are allowed to pass
    if len(actions) == 0:
        return [("pass", None)]
    return actions

    # For each game_obj
    # quest?
    # challenge? * targets
    # ability? (mult)
