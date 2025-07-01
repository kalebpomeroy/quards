import streamlit as st
import math
from collections import Counter
import requests

cards = None
cards_by_name = None


def load_cards():
    global cards, cards_by_name
    resp = requests.get("https://api.lorcana-api.com/bulk/cards")
    resp.raise_for_status()
    cards = resp.json()
    cards_by_name = {card["Name"].lower(): card for card in cards}


def by_name(name):
    if cards_by_name is None:
        load_cards()
    return cards_by_name.get(name.lower(), None)


A = "A"
B = "B"
LIGHT = "light"
DARK = "dark"
SHARED = "shared"

COLORS = {
    A: {LIGHT: "#295e66", DARK: "#1b3c42"},
    B: {LIGHT: "#5c3f6e", DARK: "#3a2a45"},
    SHARED: "#2b2b33",
}


def render_card_image(img_url, count, total=0):
    helper = "inline" if total > 0 else "none"
    return f"""<div style="display: inline-block; position: relative; margin: 4px;">
        <img src="{img_url}" class="hover-zoom" width="100" style="border-radius: 4px;" />
        <div style="position: absolute; top: 30px; left: 50px; background: rgba(0,0,0,0.6); 
                    color: white; padding: 2px 6px; border-radius: 4px; font-size: 24px;">
            {count}
        </div> 
        <div style="display: {helper}; position: absolute; top: 70px; left: 20px; background: rgba(0,0,0,0.6); 
                    color: white; padding: 2px 6px; border-radius: 4px; font-size: 12px;">
            up to {total} total
        </div>        
    </div>"""


def parse_deck(text):
    deck = Counter()
    for line in text.strip().splitlines():
        if not line.strip():
            continue
        count, name = line.strip().split(" ", 1)
        deck[name.strip()] += int(count)
    return deck


def get_diff(deck1, deck2):
    all_cards = set(deck1) | set(deck2)
    diff, same, drop_add = [], [], []
    for card in sorted(all_cards):
        a, b = deck1.get(card, 0), deck2.get(card, 0)
        if a == 0 or b == 0:
            drop_add.append((card, a, b))
        elif a != b:
            shared = min(a, b)
            same.append((card, shared, shared))
            diff.append((card, a, b))
        else:
            same.append((card, a, b))
    return diff, same, drop_add


def get_deck_stats(deck):
    ink, nonink = 0, 0
    types = {}
    for name, q in deck.items():
        card = by_name(name)
        if card is None:
            continue
        if card["Inkable"]:
            ink += q
        else:
            nonink += q
        types[card["Type"]] = types.get(card["Type"], 0) + q
    return (nonink, ink), types


def render_deck_input(label, cls_name):

    name = st.text_input(f"{label} Name") or label
    text = st.text_area(f"{label} List", height=300)
    return name, parse_deck(text) if text else None


def render_ink_bar(inks):
    uninkable, inkable = inks
    total = sum(inks)
    ink_pct = (inkable / total) * 100
    percent = 100 - ink_pct

    st.markdown(
        f"""    <div style="display: flex; flex-direction: row; align-items: center; margin: 8px;">
        <div style="width: 100%; height: 150px; display: flex; flex-direction: column-reverse; border-radius: 8px; overflow: hidden; border: 1px solid #444;">
            <div style="height: {100 - percent}%; background-color: #4ab1f1; display: flex; flex-direction: column; justify-content: flex-end; text-align: center; font-size: 20px; color: black;">
                <div>{inkable}</div>
                <div style="font-style: italic; padding: 4px 0 8px 0; font-size: 14px;">
                    {total} cards ({percent}% inkable)
                </div>
            </div>
            <div style="flex: 1; background-color: #f4a6a6; text-align: center; font-size: 20px; color: black;">
                {uninkable}
            </div>
        </div>
    </div>
    """,
        unsafe_allow_html=True,
    )


def render_type_bar(types, label=""):
    total = sum(types.values())
    max_height = 150
    all_types = sorted(types.keys())
    colors = [
        "#4ab1f1",
        "#f4a6a6",
        "#a6d884",
        "#f7d084",
        "#bba6f4",
        "#ffa07a",
        "#98e2e2",
        "#e2a698",
        "#b9e296",
        "#f4c4f1",
    ]

    bar_segments = ""
    for i, t in enumerate(all_types):
        count = types[t]
        height = int((count / total) * max_height)
        color = colors[i % len(colors)]
        bar_segments += f"""<div title="{t}: {count}" style="
                height: {height}px;
                background-color: {color};
                width: 100%;
                display: flex;
                align-items: flex-end;
                justify-content: center;
                font-size: 12px;
                color: black;
            ">{count if count > 0 else ''}</div>
        """

    st.markdown(
        f"""<div style="text-align: center; margin: 0 16px;">
            <div style="height: {max_height}px; width: 40px; display: flex; flex-direction: column-reverse;
                        border-radius: 6px; overflow: hidden; border: 1px solid #444;">{bar_segments}</div>
            <div style="margin-top: 6px; font-weight: bold;">{label}</div>
        </div>
        """,
        unsafe_allow_html=True,
    )


def render_type_pie(types, label=""):
    total = sum(types.values())
    all_types = sorted(types.keys())
    colors = [
        "#4ab1f1",
        "#f4a6a6",
        "#a6d884",
        "#f7d084",
        "#bba6f4",
        "#ffa07a",
        "#98e2e2",
        "#e2a698",
        "#b9e296",
        "#f4c4f1",
    ]

    # Build conic gradient and label overlays
    start = 0
    segments = []
    labels_html = []
    legend_items = []

    for i, t in enumerate(all_types):
        count = types[t]
        pct = count / total * 100
        color = colors[i % len(colors)]
        end = start + pct
        mid_angle = (start + end) / 2

        # Polar to cartesian for label placement
        x = 50 + 30 * math.cos(math.radians(mid_angle * 3.6 - 90))
        y = 50 + 30 * math.sin(math.radians(mid_angle * 3.6 - 90))

        labels_html.append(
            f"""<div style="
                position: absolute; 
                left: {x:.1f}%; 
                top: {y:.1f}%;
                transform: translate(-50%, -50%);
                font-size: 12px;
                color: black;
                font-weight: bold;">
                {count}
            </div>"""
        )

        segments.append(f"{color} {start:.2f}% {end:.2f}%")
        legend_items.append(
            f"""<div style="display: flex; align-items: center; margin: 4px 0;">
                    <div style="width: 12px; height: 12px; background: {color}; margin-right: 6px; border-radius: 2px;"></div>
                    <div style="font-size: 14px;">{t}: {count}</div>
                </div>"""
        )
        start = end

    gradient = ", ".join(segments)
    legend_html = "".join(legend_items)
    label_overlay = "".join(labels_html)

    st.markdown(
        f"""
        <div style="display: flex; align-items: center; justify-content: center; gap: 24px; margin: 16px;">
            <div style="text-align: left;">
                {legend_html}
            </div>
            <div style="position: relative; width: 120px; height: 120px;">
                <div style="width: 100%; height: 100%; border-radius: 50%;
                            background: conic-gradient({gradient});
                            border: 2px solid #444;">
                </div>
                {label_overlay}
            </div>
        </div>
        """,
        unsafe_allow_html=True,
    )


def render_card_grid(titles, card_tuples, get_image_fn, diff_mode=False):
    columns = st.columns(len(titles))

    c = COLORS[A][DARK] if len(titles) > 1 else COLORS[SHARED]
    with columns[0]:
        st.markdown(f"**{titles[0]}**")
        st.markdown(
            f"<div style='background-color:{c}; display: flex; flex-wrap: wrap;gap: 8px'>"
            + "".join(
                [
                    render_card_image(
                        get_image_fn(card), f"+{a - b}" if diff_mode else f"{a}", a
                    )
                    for card, a, b in card_tuples
                    if (a - b > 0 if diff_mode else a > 0) and get_image_fn(card)
                ]
            )
            + "</div>",
            unsafe_allow_html=True,
        )
    if len(titles) == 1:
        return

    with columns[1]:
        st.markdown(f"**{titles[1]}**")
        st.markdown(
            f"<div style='background-color:{COLORS[B][DARK]}; display: flex; flex-wrap: wrap;gap: 8px'>"
            + "".join(
                [
                    render_card_image(
                        get_image_fn(card), f"+{b - a}" if diff_mode else f"{b}", b
                    )
                    for card, a, b in card_tuples
                    if (b - a > 0 if diff_mode else b > 0) and get_image_fn(card)
                ]
            )
            + "</div>",
            unsafe_allow_html=True,
        )


# Main App UI
st.title("Lorcana Deck Comparator")
st.markdown(
    f"""<style>
    .hover-zoom {{
        transition: transform 0.2s ease;
    }}
    .hover-zoom:hover {{
        transform: scale(2.5);
        z-index: 100;
        position: relative;
    }}
    .deck_a textarea {{
            background-color: {COLORS[A][LIGHT]} !important;
            color: white !important;
    }}
    .deck_b textarea {{
            background-color: {COLORS[B][LIGHT]} !important;
            color: white !important;
    }}
    </style>""",
    unsafe_allow_html=True,
)
columns = st.columns(2)
with columns[0]:
    deck1_name, deck1 = render_deck_input("Deck A", "deck_a")
with columns[1]:
    deck2_name, deck2 = render_deck_input("Deck B", "deck_b")

if deck1 and deck2:
    diff, same, drops = get_diff(deck1, deck2)
    d1_ink, d1_types = get_deck_stats(deck1)
    d2_ink, d2_types = get_deck_stats(deck2)

    with columns[0]:
        render_ink_bar(d1_ink)
        render_type_pie(d1_types)
    with columns[1]:
        render_ink_bar(d2_ink)
        render_type_pie(d2_types)

    render_card_grid(
        [f"Only in {deck1_name}", f"Only in {deck2_name}"],
        drops,
        lambda name: by_name(name)["Image"] if by_name(name) else None,
    )
    render_card_grid(
        [f"{deck1_name} has more...", f"{deck2_name} has more..."],
        diff,
        lambda name: by_name(name)["Image"] if by_name(name) else None,
        diff_mode=True,
    )

    render_card_grid(
        ["Both Decks have these same cards"],
        same,
        lambda name: by_name(name)["Image"] if by_name(name) else None,
    )
