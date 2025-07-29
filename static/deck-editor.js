let isEditMode = false;
let originalDeckName = null;
let cardDatabase = {};

// Load data on page load
window.addEventListener('load', async () => {
    parseUrlParams();
    await loadCardDatabase();
    setupEventListeners();
    
    if (isEditMode) {
        await loadDeckForEditing();
    }
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

function parseUrlParams() {
    const urlParams = new URLSearchParams(window.location.search);
    const deckName = urlParams.get('deck');
    
    if (deckName) {
        isEditMode = true;
        originalDeckName = deckName;
        document.getElementById('pageTitle').textContent = `Edit Deck: ${deckName}`;
        document.getElementById('deckName').value = deckName;
    }
}

async function loadDeckForEditing() {
    try {
        const response = await fetch(`/api/decks/${originalDeckName}`);
        const data = await response.json();
        const deck = data.data;
        
        document.getElementById('deckName').value = deck.name;
        document.getElementById('deckDescription').value = deck.description || '';
        
        // Convert cards object to text format (from IDs to names)
        const cardListText = Object.entries(deck.cards)
            .map(([cardId, count]) => {
                const cardData = cardDatabase[cardId];
                const cardName = cardData ? cardData.Name : cardId;
                return `${count} ${cardName}`;
            })
            .join('\n');
        
        document.getElementById('cardList').value = cardListText;
        updateCardCount();
        
    } catch (error) {
        console.error('Failed to load deck:', error);
        alert('Failed to load deck for editing');
        window.location.href = '/decks.html';
    }
}

function setupEventListeners() {
    // Real-time card count update
    document.getElementById('cardList').addEventListener('input', updateCardCount);
    
    // Save deck
    document.getElementById('saveDeckBtn').addEventListener('click', saveDeck);
    
    // Form validation
    document.getElementById('deckName').addEventListener('input', validateForm);
}

function updateCardCount() {
    const cardListText = document.getElementById('cardList').value;
    const { cards, totalCount, errors } = parseCardList(cardListText);
    
    // Update counter
    document.getElementById('cardCount').textContent = totalCount;
    document.getElementById('cardCount').style.color = totalCount === 60 ? '#4CAF50' : '#ff6666';
    
    // Show validation errors
    const errorContainer = document.getElementById('validationErrors');
    const errorList = document.getElementById('errorList');
    
    if (errors.length > 0) {
        errorList.innerHTML = errors.map(error => `<li>${error}</li>`).join('');
        errorContainer.style.display = 'block';
    } else {
        errorContainer.style.display = 'none';
    }
    
    validateForm();
}

function parseCardList(text) {
    const lines = text.split('\n').filter(line => line.trim());
    const cards = {};
    const errors = [];
    let totalCount = 0;
    
    lines.forEach((line, index) => {
        const trimmed = line.trim();
        if (!trimmed) return;
        
        // Parse format: "3 Huey - Reliable Leader"
        const match = trimmed.match(/^(\d+)\s+(.+)$/);
        if (!match) {
            errors.push(`Line ${index + 1}: Invalid format. Use "Quantity Card Name"`);
            return;
        }
        
        const [, quantityStr, cardName] = match;
        const quantity = parseInt(quantityStr);
        
        if (isNaN(quantity) || quantity <= 0) {
            errors.push(`Line ${index + 1}: Invalid quantity "${quantityStr}"`);
            return;
        }
        
        if (quantity > 4) {
            errors.push(`Line ${index + 1}: Cannot have more than 4 copies of a card`);
            return;
        }
        
        // Find card ID by name (case insensitive)
        const cardData = Object.values(cardDatabase).find(card => 
            card.Name.toLowerCase() === cardName.toLowerCase()
        );
        if (!cardData) {
            errors.push(`Line ${index + 1}: Card "${cardName}" not found in database`);
            return;
        }
        
        const cardId = cardData.Unique_ID;
        
        if (cards[cardId]) {
            errors.push(`Line ${index + 1}: Duplicate card "${cardName}"`);
            return;
        }
        
        cards[cardId] = quantity;
        totalCount += quantity;
    });
    
    if (totalCount !== 60 && totalCount > 0) {
        errors.push(`Total cards must be exactly 60 (currently ${totalCount})`);
    }
    
    return { cards, totalCount, errors };
}

function validateForm() {
    const deckName = document.getElementById('deckName').value.trim();
    const cardListText = document.getElementById('cardList').value.trim();
    const { errors } = parseCardList(cardListText);
    
    const isValid = deckName && cardListText && errors.length === 0;
    document.getElementById('saveDeckBtn').disabled = !isValid;
}

async function saveDeck() {
    const deckName = document.getElementById('deckName').value.trim();
    const deckDescription = document.getElementById('deckDescription').value.trim();
    const cardListText = document.getElementById('cardList').value.trim();
    
    const { cards, errors } = parseCardList(cardListText);
    
    if (errors.length > 0) {
        alert('Please fix validation errors before saving');
        return;
    }
    
    if (!deckName) {
        alert('Please enter a deck name');
        return;
    }
    
    const deckData = {
        name: deckName,
        description: deckDescription,
        cards: cards
    };
    
    try {
        // Show loading state
        const saveBtn = document.getElementById('saveDeckBtn');
        const originalText = saveBtn.textContent;
        saveBtn.textContent = 'Saving...';
        saveBtn.disabled = true;
        
        let response;
        if (isEditMode) {
            response = await fetch(`/api/decks/${originalDeckName}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(deckData)
            });
        } else {
            response = await fetch('/api/decks', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(deckData)
            });
        }
        
        if (response.ok) {
            // Redirect back to decks page
            window.location.href = '/decks.html';
        } else {
            const error = await response.json();
            throw new Error(error.error || 'Failed to save deck');
        }
        
    } catch (error) {
        console.error('Failed to save deck:', error);
        alert(`Failed to save deck: ${error.message}`);
        
        // Restore button state
        const saveBtn = document.getElementById('saveDeckBtn');
        saveBtn.textContent = originalText;
        saveBtn.disabled = false;
    }
}