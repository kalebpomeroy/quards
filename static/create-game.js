let decks = [];
let gameType = 'generated'; // 'generated' or 'recorded'

// Load data on page load
window.addEventListener('load', async () => {
    await loadDecks();
    updateDeckDropdowns();
    setupEventListeners();
});

async function loadDecks() {
    try {
        const response = await fetch('/api/decks');
        const data = await response.json();
        decks = data.data || [];
        console.log(`Loaded ${decks.length} decks`);
    } catch (error) {
        console.error('Failed to load decks:', error);
        decks = [];
    }
}

function updateDeckDropdowns() {
    const player1Select = document.getElementById('player1Deck');
    const player2Select = document.getElementById('player2Deck');
    
    // Clear existing options (except first)
    player1Select.innerHTML = '<option value="">Select a deck...</option>';
    player2Select.innerHTML = '<option value="">Select a deck...</option>';
    
    decks.forEach(deck => {
        const option1 = document.createElement('option');
        option1.value = deck.name;
        option1.textContent = `${deck.name} (${deck.cardCount} cards)`;
        player1Select.appendChild(option1);
        
        const option2 = document.createElement('option');
        option2.value = deck.name;
        option2.textContent = `${deck.name} (${deck.cardCount} cards)`;
        player2Select.appendChild(option2);
    });
}

function setupEventListeners() {
    // Deck selection
    document.getElementById('player1Deck').addEventListener('change', handleDeckSelection);
    document.getElementById('player2Deck').addEventListener('change', handleDeckSelection);
    
    
    // Seed generation
    document.getElementById('generateSeedBtn').addEventListener('click', generateRandomSeed);
    
    // Create game
    document.getElementById('createGameBtn').addEventListener('click', createGame);
    
    // File upload
    document.getElementById('gameLogFile').addEventListener('change', handleFileUpload);
}

function switchTab(tabName) {
    // Hide all tabs
    document.getElementById('generateTab').classList.remove('active');
    document.getElementById('uploadTab').classList.remove('active');
    
    // Remove active class from all buttons
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    
    // Show selected tab
    document.getElementById(tabName + 'Tab').classList.add('active');
    
    // Add active class to clicked button
    event.target.classList.add('active');
}

function handleFileUpload() {
    const fileInput = document.getElementById('gameLogFile');
    const uploadBtn = document.getElementById('uploadGameBtn');
    
    if (fileInput.files.length > 0) {
        uploadBtn.disabled = false;
    } else {
        uploadBtn.disabled = true;
    }
}


function selectGameType(type) {
    gameType = type;
    
    // Update UI
    document.getElementById('gameTypeGenerated').classList.toggle('selected', type === 'generated');
    document.getElementById('gameTypeRecorded').classList.toggle('selected', type === 'recorded');
    
    // Show/hide seed section
    const seedSection = document.getElementById('seedSection');
    if (type === 'generated') {
        seedSection.classList.remove('hidden');
    } else {
        seedSection.classList.add('hidden');
    }
    
    validateForm();
}

function handleDeckSelection() {
    const player1Deck = document.getElementById('player1Deck').value;
    const player2Deck = document.getElementById('player2Deck').value;
    
    // Update deck info displays
    updateDeckInfo('player1Info', player1Deck);
    updateDeckInfo('player2Info', player2Deck);
    
    validateForm();
}

function updateDeckInfo(elementId, deckName) {
    const element = document.getElementById(elementId);
    
    if (!deckName) {
        element.style.display = 'none';
        return;
    }
    
    const deck = decks.find(d => d.name === deckName);
    if (deck) {
        const createdDate = new Date(deck.created).toLocaleDateString();
        
        element.innerHTML = `
            <div><strong>${deck.name}</strong></div>
            <div style="margin: 5px 0; color: #ccc;">${deck.description || 'No description'}</div>
            <div style="font-size: 14px; color: #888;">
                ${deck.cardCount} cards • Created ${createdDate}
            </div>
        `;
        element.style.display = 'block';
    }
}

function validateForm() {
    const player1Deck = document.getElementById('player1Deck').value;
    const player2Deck = document.getElementById('player2Deck').value;
    
    const isValid = player1Deck && player2Deck;
    document.getElementById('createGameBtn').disabled = !isValid;
}

function generateRandomSeed() {
    // Generate a UUID-like string
    const chars = '0123456789abcdef';
    let seed = '';
    for (let i = 0; i < 8; i++) {
        seed += chars[Math.floor(Math.random() * chars.length)];
    }
    seed += '-';
    for (let i = 0; i < 4; i++) {
        seed += chars[Math.floor(Math.random() * chars.length)];
    }
    seed += '-4'; // Version 4 UUID
    for (let i = 0; i < 3; i++) {
        seed += chars[Math.floor(Math.random() * chars.length)];
    }
    seed += '-';
    seed += chars[8 + Math.floor(Math.random() * 4)]; // Variant bits
    for (let i = 0; i < 3; i++) {
        seed += chars[Math.floor(Math.random() * chars.length)];
    }
    seed += '-';
    for (let i = 0; i < 12; i++) {
        seed += chars[Math.floor(Math.random() * chars.length)];
    }
    
    document.getElementById('gameSeed').value = seed;
}

async function createGame() {
    const player1Deck = document.getElementById('player1Deck').value;
    const player2Deck = document.getElementById('player2Deck').value;
    const gameSeed = document.getElementById('gameSeed').value.trim();
    
    if (!player1Deck || !player2Deck) {
        alert('Please select decks for both players');
        return;
    }
    
    const gameData = {
        player1Deck: player1Deck,
        player2Deck: player2Deck,
        seed: gameSeed || null
    };
    
    try {
        // Show loading state
        const createBtn = document.getElementById('createGameBtn');
        const originalText = createBtn.textContent;
        createBtn.textContent = 'Creating...';
        createBtn.disabled = true;
        
        const response = await fetch('/api/games', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(gameData)
        });
        
        if (response.ok) {
            const result = await response.json();
            // Redirect to game viewer
            window.location.href = `/?game=${result.data.name}`;
        } else {
            const error = await response.json();
            throw new Error(error.error || 'Failed to create game');
        }
        
    } catch (error) {
        console.error('Failed to create game:', error);
        alert(`Failed to create game: ${error.message}`);
        
        // Restore button state
        const createBtn = document.getElementById('createGameBtn');
        createBtn.textContent = originalText;
        createBtn.disabled = false;
    }
}

async function uploadGame() {
    const fileInput = document.getElementById('gameLogFile');
    
    if (!fileInput.files[0]) {
        alert('Please select a game log file');
        return;
    }
    
    const file = fileInput.files[0];
    
    try {
        // Show loading state
        const uploadBtn = document.getElementById('uploadGameBtn');
        const originalText = uploadBtn.textContent;
        uploadBtn.textContent = 'Uploading...';
        uploadBtn.disabled = true;
        
        // Read file content
        const fileContent = await readFileContent(file);
        
        // Parse log to extract deck names (basic parsing)
        const deckInfo = parseLogForDecks(fileContent);
        
        const gameData = {
            player1Deck: deckInfo.player1Deck || 'Unknown Deck',
            player2Deck: deckInfo.player2Deck || 'Unknown Deck',
            logContent: fileContent
        };
        
        const response = await fetch('/api/games', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(gameData)
        });
        
        if (response.ok) {
            const result = await response.json();
            // Navigate to game viewer
            window.location.href = `/?game=${result.data.name}`;
        } else {
            const error = await response.json();
            throw new Error(error.error || 'Failed to upload game');
        }
        
    } catch (error) {
        console.error('Failed to upload game:', error);
        alert(`Failed to upload game: ${error.message}`);
        
        // Restore button state
        const uploadBtn = document.getElementById('uploadGameBtn');
        uploadBtn.textContent = 'Upload Game';
        uploadBtn.disabled = false;
    }
}

function readFileContent(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = e => resolve(e.target.result);
        reader.onerror = reject;
        reader.readAsText(file);
    });
}

function parseLogForDecks(logContent) {
    // Simple parsing to extract deck names from start_game action
    const lines = logContent.split('\n');
    for (const line of lines) {
        if (line.includes('start_game')) {
            try {
                const match = line.match(/\{.*\}/);
                if (match) {
                    const data = JSON.parse(match[0]);
                    return {
                        player1Deck: data.player1_deck,
                        player2Deck: data.player2_deck
                    };
                }
            } catch (e) {
                // Ignore parsing errors
            }
        }
    }
    return { player1Deck: null, player2Deck: null };
}