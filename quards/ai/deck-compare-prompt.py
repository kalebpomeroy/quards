def build_prompt(d1_name, d1, d2_name, d2):
    header = f"""
        You are a Lorcana strategy expert. Evaluate how changes between decks impact gameplay. 
        Neither decklist is necessarily superior or first. Be as objective as possible. 
        
        Use any outside knowledge you have available to you (including searching the web). Give 
        for each change, generate a paragraph of hypothetical positives as well as what we lose by cutting it. 

        If a card has a very high ceiling but low floor call that out (and vice versa). If the changes are 
        pretty even (for example, two characters with matching costs/stats, but slightly different ability). 

        At the end, supply a summary of A vs B. Pay special attention to curve (total/average costs of cards),
        inkable vs non-inkable, and call out anything that seems specifically different or consistent with the chaannges. 

        {d1_name}
        {"\n".join([f"{count}x {card}"for card, count in d1.items()])}

        {d2_name}
        {"\n".join([f"{count}x {card}"for card, count in d2.items()])}

        
    """

    diff, same, drops = get_diff(deck1, deck2)

    diff_text = "\nChanges:\n" + "\n".join(f"{card}: {a} → {b}" for card, a, b in diff)

    same_text = "\nStays the same:\n" + "\n".join(f"{a} {card}" for card, a, _ in same)
    drops_text = "\nDrop/Add:\n" + "\n".join(
        f"{card} {a} x {b}" for card, a, b in drops
    )

    card_texts = "\n".join(
        format_card(by_name(card)) for card, _, _ in diff if by_name(card) is not None
    )
    same_card_texts = "\n".join(
        format_card(by_name(card)) for card, _, _ in same if by_name(card) is not None
    )

    return f"""{header}

{same_text}

{drops_text}

{diff_text}

Card Definitions
{card_texts}

(Other cards that didn't change but might be helpful context)
{same_card_texts}

Explain the implications."""
