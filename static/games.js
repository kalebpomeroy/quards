let games = [];
let decks = [];
let filteredGames = [];
let currentFilters = {
    deck: null,
    player: null,
    date: null
};

// Load data on page load
window.addEventListener('load', async () => {
    parseUrlParams();
    await loadDecks();
    await loadGames();
    setupEventListeners();
    applyFilters();
});

function parseUrlParams() {
    const urlParams = new URLSearchParams(window.location.search);
    const deckParam = urlParams.get('deck');
    
    if (deckParam) {
        currentFilters.deck = deckParam;
        document.getElementById('filteredMessage').innerHTML = `Showing games using deck: <strong>${deckParam}</strong>`;
        document.getElementById('filteredMessage').style.display = 'block';
    }
}

async function loadDecks() {
    try {
        const response = await fetch('/api/decks');
        const data = await response.json();
        decks = data.data || [];
        
        // Populate deck filter dropdown
        const deckFilter = document.getElementById('deckFilter');
        decks.forEach(deck => {
            const option = document.createElement('option');
            option.value = deck.name;
            option.textContent = deck.name;
            deckFilter.appendChild(option);
        });
        
        // Set initial deck filter if from URL
        if (currentFilters.deck) {
            deckFilter.value = currentFilters.deck;
        }
        
        console.log(`Loaded ${decks.length} decks`);
    } catch (error) {
        console.error('Failed to load decks:', error);
        decks = [];
    }
}

async function loadGames() {
    try {
        const response = await fetch('/api/games');
        const data = await response.json();
        games = data.data || [];
        
        // Convert API format to frontend format
        games = games.map(game => ({
            id: game.name || `game-${game.id}`, // Fallback to ID-based name
            name: game.name,
            player1Deck: game.player1Deck,
            player2Deck: game.player2Deck,
            player1DeckDescription: getDecksDescription(game.player1Deck),
            player2DeckDescription: getDecksDescription(game.player2Deck),
            created: game.created,
            turns: game.turns,
            winner: game.winner,
            seed: game.seed
        }));
        
        console.log(`Loaded ${games.length} games`);
    } catch (error) {
        console.error('Failed to load games:', error);
        games = [];
    }
}

function getDecksDescription(deckName) {
    const deck = decks.find(d => d.name === deckName);
    return deck ? deck.description || 'No description' : 'Unknown deck';
}


function setupEventListeners() {
    document.getElementById('deckFilter').addEventListener('change', handleFilterChange);
    document.getElementById('playerFilter').addEventListener('change', handleFilterChange);
    document.getElementById('dateFilter').addEventListener('change', handleFilterChange);
}

function handleFilterChange() {
    currentFilters.deck = document.getElementById('deckFilter').value || null;
    currentFilters.player = document.getElementById('playerFilter').value || null;
    currentFilters.date = document.getElementById('dateFilter').value || null;
    
    applyFilters();
}

function applyFilters() {
    filteredGames = games.filter(game => {
        // Deck filter
        if (currentFilters.deck) {
            if (game.player1Deck !== currentFilters.deck && game.player2Deck !== currentFilters.deck) {
                return false;
            }
        }
        
        // Player filter
        if (currentFilters.player) {
            const playerNum = parseInt(currentFilters.player);
            if (playerNum === 1 && game.player1Deck !== currentFilters.deck) {
                return false;
            }
            if (playerNum === 2 && game.player2Deck !== currentFilters.deck) {
                return false;
            }
        }
        
        // Date filter
        if (currentFilters.date) {
            const gameDate = new Date(game.created);
            const now = new Date();
            
            switch (currentFilters.date) {
                case 'today':
                    if (gameDate.toDateString() !== now.toDateString()) {
                        return false;
                    }
                    break;
                case 'week':
                    const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
                    if (gameDate < weekAgo) {
                        return false;
                    }
                    break;
                case 'month':
                    const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
                    if (gameDate < monthAgo) {
                        return false;
                    }
                    break;
            }
        }
        
        return true;
    });
    
    renderGames();
}

function renderGames() {
    const grid = document.getElementById('gamesGrid');
    grid.innerHTML = '';
    
    if (filteredGames.length === 0) {
        const hasFilters = currentFilters.deck || currentFilters.player || currentFilters.date;
        const message = hasFilters ? 'No games match the current filters.' : 'No games found. Create your first game to get started!';
        
        grid.innerHTML = `<div class="no-games">${message}</div>`;
        return;
    }
    
    filteredGames.forEach(game => {
        const gameCard = document.createElement('div');
        gameCard.className = 'game-card';
        gameCard.onclick = () => viewGame(game.id);
        
        const createdDate = new Date(game.created).toLocaleDateString();
        const createdTime = new Date(game.created).toLocaleTimeString();
        
        gameCard.innerHTML = `
            <div class="game-header">
                <div class="game-id">${game.name || game.id}</div>
                <div class="game-date">${createdDate}</div>
            </div>
            
            <div class="game-matchup">
                <div class="player-info player1">
                    <div class="deck-name">${game.player1Deck}</div>
                    <div class="deck-description">${game.player1DeckDescription}</div>
                </div>
                
                <div class="vs">VS</div>
                
                <div class="player-info player2">
                    <div class="deck-name">${game.player2Deck}</div>
                    <div class="deck-description">${game.player2DeckDescription}</div>
                </div>
            </div>
            
            <div class="game-meta">
                <span>${game.turns} turns</span>
                <span>Winner: Player ${game.winner}</span>
                ${game.seed ? `<span>Seed: ${game.seed}</span>` : '<span>No seed</span>'}
            </div>
        `;
        
        grid.appendChild(gameCard);
    });
}

function viewGame(gameId) {
    // Navigate to game viewer with this game's log
    window.location.href = `/?game=${gameId}`;
}

function clearFilters() {
    currentFilters = { deck: null, player: null, date: null };
    
    document.getElementById('deckFilter').value = '';
    document.getElementById('playerFilter').value = '';
    document.getElementById('dateFilter').value = '';
    
    // Hide filtered message unless it's from URL
    const urlParams = new URLSearchParams(window.location.search);
    if (!urlParams.get('deck')) {
        document.getElementById('filteredMessage').style.display = 'none';
    }
    
    applyFilters();
}