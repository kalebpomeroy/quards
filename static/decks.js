let decks = [];
let cardDatabase = {};
let imageCache = new Map();

// Load data on page load
window.addEventListener('load', async () => {
    await loadCardDatabase();
    await loadDecks();
    renderDecks();
    setupEventListeners();
});

async function loadCardDatabase() {
    try {
        const response = await fetch('/cards.json');
        const cards = await response.json();
        
        // Create lookup by Unique_ID
        cards.forEach(card => {
            cardDatabase[card.Unique_ID] = card;
        });
        
        console.log(`Loaded ${cards.length} cards into database`);
    } catch (error) {
        console.error('Failed to load card database:', error);
    }
}

async function loadDecks() {
    try {
        // Add cache-busting to ensure fresh data
        const response = await fetch(`/api/decks?t=${Date.now()}`);
        const data = await response.json();
        decks = data.data || [];
        
        console.log(`Loaded ${decks.length} decks`);
    } catch (error) {
        console.error('Failed to load decks:', error);
        decks = [];
    }
}

function renderDecks() {
    const grid = document.getElementById('deckGrid');
    grid.innerHTML = '';
    
    if (decks.length === 0) {
        grid.innerHTML = '<p style="grid-column: 1 / -1; text-align: center; color: #888;">No decks found. Create your first deck to get started!</p>';
        return;
    }
    
    decks.forEach(deck => {
        const deckCard = document.createElement('div');
        deckCard.className = 'deck-card';
        
        const createdDate = new Date(deck.created).toLocaleDateString();
        const modifiedDate = new Date(deck.modified).toLocaleDateString();
        
        deckCard.innerHTML = `
            <h3>${deck.name}</h3>
            <div class="description">${deck.description || 'No description'}</div>
            <div class="meta">
                <span>${deck.cardCount} cards</span>
                <span>Modified: ${modifiedDate}</span>
            </div>
            <div class="deck-actions">
                <button class="btn" onclick="viewDeck('${deck.name}')">View</button>
                <button class="btn secondary" onclick="editDeck('${deck.name}')">Edit</button>
                <button class="btn danger" onclick="deleteDeck('${deck.name}')">Delete</button>
            </div>
        `;
        
        grid.appendChild(deckCard);
    });
}


async function viewDeck(deckName) {
    try {
        const response = await fetch(`/api/decks/${deckName}`);
        const data = await response.json();
        const deck = data.data;
        
        const modal = document.getElementById('deckModal');
        const modalName = document.getElementById('modalDeckName');
        const modalDescription = document.getElementById('modalDeckDescription');
        const modalCards = document.getElementById('modalDeckCards');
        
        modalName.textContent = deck.name;
        modalDescription.textContent = deck.description || 'No description';
        
        // Add games button if not already exists
        let gamesButton = document.getElementById('gamesWithDeckBtn');
        if (!gamesButton) {
            gamesButton = document.createElement('button');
            gamesButton.id = 'gamesWithDeckBtn';
            gamesButton.className = 'btn secondary';
            gamesButton.style.marginTop = '10px';
            gamesButton.onclick = () => viewGamesWithDeck(deck.name);
            modalDescription.parentNode.insertBefore(gamesButton, modalCards);
        }
        gamesButton.textContent = `View Games with ${deck.name}`;
        
        // Render card grid - one image per unique card with count overlay
        modalCards.innerHTML = '';
        const cardGrid = document.createElement('div');
        cardGrid.style.cssText = 'display: grid; grid-template-columns: repeat(auto-fill, minmax(100px, 1fr)); gap: 10px; margin-top: 20px;';
        
        Object.entries(deck.cards).forEach(([cardId, count]) => {
            const cardData = cardDatabase[cardId];
            const cardItem = document.createElement('div');
            cardItem.style.cssText = 'position: relative; aspect-ratio: 5/7; border-radius: 6px; overflow: hidden;';
            
            if (cardData && cardData.Image) {
                // Create image element with caching
                const img = createCachedImage(cardData.Image, cardData.Name);
                cardItem.appendChild(img);
                
                const countOverlay = document.createElement('div');
                countOverlay.style.cssText = 'position: absolute; bottom: 4px; right: 4px; background: rgba(0,0,0,0.9); color: white; padding: 4px 8px; border-radius: 12px; font-size: 16px; font-weight: bold; min-width: 24px; text-align: center;';
                countOverlay.textContent = count;
                cardItem.appendChild(countOverlay);
            } else {
                cardItem.innerHTML = `
                    <div style="width: 100%; height: 100%; background: #654321; display: flex; align-items: center; 
                                justify-content: center; font-size: 10px; color: white; text-align: center; padding: 4px;">
                        ${cardId}
                    </div>
                    <div style="position: absolute; bottom: 4px; right: 4px; background: rgba(0,0,0,0.9); color: white; 
                                padding: 4px 8px; border-radius: 12px; font-size: 16px; font-weight: bold; min-width: 24px; text-align: center;">
                        ${count}
                    </div>
                `;
            }
            
            cardGrid.appendChild(cardItem);
        });
        
        modalCards.appendChild(cardGrid);
        modal.style.display = 'block';
        
    } catch (error) {
        console.error('Failed to load deck:', error);
        alert('Failed to load deck details');
    }
}

function editDeck(deckName) {
    window.location.href = `/deck-editor.html?deck=${deckName}`;
}

async function deleteDeck(deckName) {
    if (!confirm(`Are you sure you want to delete the deck "${deckName}"?`)) {
        return;
    }
    
    try {
        const response = await fetch(`/api/decks/${deckName}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            await loadDecks();
            renderDecks();
        } else {
            throw new Error('Failed to delete deck');
        }
    } catch (error) {
        console.error('Failed to delete deck:', error);
        alert('Failed to delete deck');
    }
}

function createDeck() {
    window.location.href = '/deck-editor.html';
}


function setupEventListeners() {
    // Modal close
    document.getElementById('closeModal').addEventListener('click', () => {
        document.getElementById('deckModal').style.display = 'none';
    });
    
    // Close modal when clicking outside
    window.addEventListener('click', (event) => {
        const modal = document.getElementById('deckModal');
        if (event.target === modal) {
            modal.style.display = 'none';
        }
    });
    
    // Create deck button
    document.getElementById('createDeckBtn').addEventListener('click', createDeck);
}


function viewGamesWithDeck(deckName) {
    window.location.href = `/games.html?deck=${deckName}`;
}

function createCachedImage(src, alt) {
    const img = document.createElement('img');
    img.alt = alt;
    img.style.cssText = 'width: 100%; height: 100%; object-fit: cover;';
    img.loading = 'lazy';
    
    // Check if image is cached
    if (imageCache.has(src)) {
        const cachedData = imageCache.get(src);
        if (cachedData.loaded) {
            img.src = src;
        } else {
            // Image is still loading, add to waiting list
            cachedData.waiting.push(img);
        }
    } else {
        // First time loading this image
        imageCache.set(src, { loaded: false, waiting: [img] });
        
        // Create a temporary image to preload
        const tempImg = new Image();
        tempImg.onload = () => {
            const cacheEntry = imageCache.get(src);
            cacheEntry.loaded = true;
            
            // Update all waiting images
            cacheEntry.waiting.forEach(waitingImg => {
                waitingImg.src = src;
            });
            cacheEntry.waiting = [];
        };
        tempImg.onerror = () => {
            // Remove from cache on error
            imageCache.delete(src);
        };
        tempImg.src = src;
    }
    
    return img;
}