let gameSteps = [];
let currentStep = 0;
let cardDatabase = {};
let isPlaying = false;
let playInterval;
let availableActions = [];
let currentGameID = null;

// Load game data on page load
window.addEventListener('load', async () => {
    console.log('Page load event triggered');
    console.log('Loading card database...');
    await loadCardDatabase();
    console.log('Card database loaded, loading game steps...');
    await loadGameSteps();
    
    // Only render if we have game steps
    if (gameSteps.length > 0) {
        console.log('Rendering current step after all data loaded');
        await renderCurrentStep(); // This will call loadAvailableActions internally
    } else {
        console.log('No game steps to render');
    }
});

async function loadCardDatabase() {
    try {
        const response = await fetch('/cards.lorcana-api.json');
        const cards = await response.json();
        
        // Create lookup map by Unique_ID
        cards.forEach(card => {
            cardDatabase[card.Unique_ID] = card;
        });
        
        const totalCards = Object.keys(cardDatabase).length;
        
        console.log(`Loaded ${totalCards} cards into database`);
    } catch (error) {
        console.error('Failed to load card database:', error);
    }
}

async function fetchGameStateForStep(stepNumber) {
    try {
        if (!currentGameID) {
            console.error('No current game ID');
            return null;
        }
        
        const response = await fetch(`/api/games/${currentGameID}/state?step=${stepNumber}`);
        
        if (!response.ok) {
            console.error('Failed to fetch game state data');
            return null;
        }
        
        const data = await response.json();
        return data.data;
    } catch (error) {
        console.error('Error fetching game state:', error);
        return null;
    }
}

// Note: Player choice detection is now handled server-side by the navigation API

async function loadGameSteps() {
    try {
        // Check URL parameters for game name
        const urlParams = new URLSearchParams(window.location.search);
        const gameID = urlParams.get('game');
        
        console.log('URL params:', window.location.search);
        console.log('Extracted gameID:', gameID);
        
        if (!gameID) {
            console.error('No game specified in URL parameters');
            gameSteps = [];
            updateStepCounter();
            return;
        }
        
        currentGameID = gameID;
        const apiUrl = `/api/games/${gameID}/navigation`;
        
        console.log('Fetching navigation data from API URL:', apiUrl);
        const response = await fetch(apiUrl);
        console.log('API response status:', response.status);
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        console.log('Navigation API response data:', data);
        
        if (data.error) {
            throw new Error(data.error);
        }
        
        const allSteps = data.data || [];
        
        // Filter to only show player choice actions using server-provided flag
        const playerChoiceSteps = allSteps.filter(step => step.isPlayerChoice);
        
        // If there are no player choice steps yet, use the last framework step
        // This handles newly created games that haven't had any player actions yet
        if (playerChoiceSteps.length === 0 && allSteps.length > 0) {
            // Use the last framework step (e.g., turn_start) and mark it as a choice point
            const lastStep = allSteps[allSteps.length - 1];
            lastStep.originalStepNumber = lastStep.step + 1; // Server provides step numbers
            gameSteps = [lastStep];
            currentStep = 0;
            console.log(`No player choice steps found, using last framework step: ${lastStep.action}`);
        } else {
            gameSteps = playerChoiceSteps;
            
            // Server already provides step numbers, just copy them
            gameSteps.forEach((step) => {
                step.originalStepNumber = step.step + 1; // Convert to 1-indexed for API
            });
            
            // Start at the last step instead of the first
            if (gameSteps.length > 0) {
                currentStep = gameSteps.length - 1;
                console.log(`Set current step to ${currentStep + 1} (last player choice step)`);
            } else {
                currentStep = 0;
                console.log('No steps loaded, staying at step 0');
            }
        }
        
        updateStepCounter();
        console.log(`Loaded ${gameSteps.length} player choice steps for game: ${gameID}, starting at step ${currentStep + 1}`);
        
        // Ensure we render the current step after loading
        if (gameSteps.length > 0) {
            console.log('About to render current step after loading');
            // Don't call renderCurrentStep here since it's called in the main load handler
        }
    } catch (error) {
        console.error('Failed to load game steps:', error);
        console.error('Error details:', error.message);
        gameSteps = [];
        updateStepCounter();
        
        // Show user-friendly error
        const stepInfo = document.getElementById('stepInfo');
        if (stepInfo) {
            stepInfo.innerHTML = `<strong>Error:</strong> Failed to load game. ${error.message}`;
        }
    }
}

function updateStepCounter() {
    document.getElementById('stepCounter').textContent = 
        `Step ${currentStep + 1} / ${gameSteps.length}`;
    
    // Update timeline
    const progress = ((currentStep + 1) / gameSteps.length) * 100;
    document.getElementById('timelineProgress').style.width = `${progress}%`;
}

async function loadAvailableActions() {
    try {
        if (!currentGameID) {
            console.log('No game loaded, skipping available actions');
            availableActions = [];
            renderActions();
            return;
        }
        
        // Check if we're on the most recent step for action execution
        const isLastStep = currentStep === gameSteps.length - 1;
        
        // Calculate actions based on historical context (up to original step number)
        const currentGameStep = gameSteps[currentStep];
        const stepNumber = currentGameStep ? currentGameStep.originalStepNumber : currentStep + 1;
        const apiUrl = `/api/games/${currentGameID}/actions?step=${stepNumber}`;
        const response = await fetch(apiUrl);
        const data = await response.json();
        
        if (data.error) {
            throw new Error(data.error);
        }
        
        availableActions = data.data || [];
        renderActions(isLastStep);
        console.log(`Loaded ${availableActions.length} available actions for game: ${currentGameID} at step ${currentStep + 1}`);
    } catch (error) {
        console.error('Failed to load available actions:', error);
        availableActions = [];
        renderActions(false);
    }
}

function renderActions(isCurrentStep = false) {
    const grid = document.getElementById('actionsGrid');
    const currentPlayerDiv = document.getElementById('currentPlayer');
    
    // Always show only valid actions - no debug info for invalid actions
    const actionsToShow = availableActions.filter(action => action.valid);
    
    // For past steps, find which action was actually chosen
    let chosenAction = null;
    if (!isCurrentStep && currentStep < gameSteps.length - 1) {
        const nextStep = gameSteps[currentStep + 1];
        chosenAction = findMatchingAction(nextStep, actionsToShow);
    }
    
    if (actionsToShow.length === 0) {
        if (isCurrentStep) {
            grid.innerHTML = '<div class="action-section"><div class="action-category">No Actions</div><div class="action-description">No actions available at this time</div></div>';
        } else {
            grid.innerHTML = '<div class="action-section"><div class="action-category">No Valid Actions</div><div class="action-description">No valid actions were available at this step</div></div>';
        }
        currentPlayerDiv.textContent = 'Current Player: Unknown';
        return;
    }
    
    // Determine current player from actions
    const currentPlayer = determineCurrentPlayer();
    currentPlayerDiv.textContent = `Current Player: Player ${currentPlayer}`;
    
    // Group actions by category
    const actionCategories = {
        'pass': [],
        'ink_card': [],
        'play_card': [],
        'quest': [],
    };
    
    actionsToShow.forEach(action => {
        if (actionCategories[action.type]) {
            actionCategories[action.type].push(action);
        }
    });
    
    grid.innerHTML = '';
    
    // Always render all categories (even if empty) for consistent layout
    const allCategories = ['ink_card', 'play_card', 'quest', 'challenge', 'pass'];
    
    allCategories.forEach(category => {
        const actions = actionCategories[category] || [];
        
        const sectionDiv = document.createElement('div');
        sectionDiv.className = 'action-section';
        
        const categoryTitle = document.createElement('div');
        categoryTitle.className = 'action-category';
        categoryTitle.textContent = category.replace('_', ' ').toUpperCase();
        sectionDiv.appendChild(categoryTitle);
        
        const cardsContainer = document.createElement('div');
        cardsContainer.className = 'action-cards-container';
        
        if (actions.length === 0) {
            // Empty section placeholder
            const emptyDiv = document.createElement('div');
            emptyDiv.className = 'empty-action-section';
            emptyDiv.textContent = 'No actions available';
            cardsContainer.appendChild(emptyDiv);
        } else if (category === 'pass') {
            // Special handling for pass action
            const passAction = actions[0];
            const isChosenAction = chosenAction && actionsMatch(passAction, chosenAction);
            
            const passButton = document.createElement('div');
            passButton.className = getActionClass(isCurrentStep, isChosenAction);
            passButton.innerHTML = `
                <div class="pass-button">Pass Turn</div>
                ${getActionHint(isCurrentStep, isChosenAction)}
            `;
            
            addActionClickHandler(passButton, passAction, isCurrentStep, isChosenAction);
            cardsContainer.appendChild(passButton);
        } else {
            // Card-based actions
            actions.forEach(action => {
                const isChosenAction = chosenAction && actionsMatch(action, chosenAction);
                const cardId = action.parameters?.card_id;
                const cardInfo = cardDatabase[cardId];
                
                
                const cardElement = document.createElement('div');
                cardElement.className = `action-card-image ${getActionClass(isCurrentStep, isChosenAction)}`;
                
                if (cardInfo) {
                    cardElement.innerHTML = `
                        <img src="${cardInfo.Image}" alt="${cardInfo.Name}" />
                        <div class="card-tooltip">${cardInfo.Name}${action.parameters?.cost ? ` (${action.parameters.cost} ink)` : ''}${action.parameters?.lore ? ` (+${action.parameters.lore} lore)` : ''}</div>
                        ${getActionHint(isCurrentStep, isChosenAction)}
                    `;
                    
                    // Add hover tooltip
                    cardElement.addEventListener('mouseenter', () => showCardDetail(cardInfo));
                    cardElement.addEventListener('mouseleave', hideCardDetail);
                } else {
                    // Fallback for missing card info
                    cardElement.innerHTML = `
                        <img src="/back.png" alt="Card Back" />
                        <div class="card-tooltip">${cardId} (Card data not found)${action.parameters?.cost ? ` (${action.parameters.cost} ink)` : ''}${action.parameters?.lore ? ` (+${action.parameters.lore} lore)` : ''}</div>
                        ${getActionHint(isCurrentStep, isChosenAction)}
                    `;
                }
                
                addActionClickHandler(cardElement, action, isCurrentStep, isChosenAction);
                cardsContainer.appendChild(cardElement);
            });
        }
        
        sectionDiv.appendChild(cardsContainer);
        grid.appendChild(sectionDiv);
    });
}

function getActionClass(isCurrentStep, isChosenAction) {
    if (isCurrentStep) {
        return 'executable';
    } else if (isChosenAction) {
        return 'chosen';
    } else {
        return 'alternative';
    }
}

function getActionHint(isCurrentStep, isChosenAction) {
    if (!isCurrentStep && isChosenAction) {
        return '<div class="chosen-hint">✓ CHOSEN - Click to advance</div>';
    }
    return '';
}

function addActionClickHandler(element, action, isCurrentStep, isChosenAction) {
    if (isCurrentStep) {
        element.addEventListener('click', () => executeAction(action));
    } else if (isChosenAction) {
        element.addEventListener('click', () => nextStep());
    } else {
        // For alternative actions, just make them non-interactive for now
        element.style.cursor = 'default';
    }
}

function determineCurrentPlayer() {
    // Look for the most recent turn_start or pass action in game steps
    if (gameSteps && gameSteps.length > 0) {
        // Search backwards through steps to find the most recent turn_start or pass
        for (let i = Math.min(currentStep, gameSteps.length - 1); i >= 0; i--) {
            const step = gameSteps[i];
            if (step.action === 'turn_start' && step.parameters && step.parameters.player) {
                return step.parameters.player;
            } else if (step.action === 'pass') {
                // After a pass, it's the other player's turn
                return step.player === 1 ? 2 : 1;
            }
        }
    }
    
    // Default to player 1 if no turn_start or pass found
    return 1;
}

function executeAction(action) {
    if (!action.valid) return;
    
    console.log('Executing action:', action);
    
    // Execute action immediately without confirmation
    appendActionToLog(action);
}

// Find which action from the available actions matches the actual game step
function findMatchingAction(gameStep, availableActions) {
    for (const action of availableActions) {
        if (actionsMatch(action, gameStep)) {
            return action;
        }
    }
    return null;
}

// Check if an available action matches a game step
function actionsMatch(action, gameStep) {
    // Match by action type
    if (action.type !== gameStep.action) {
        return false;
    }
    
    // For actions with parameters, match key parameters
    if (action.parameters) {
        // For card-related actions, match the card_id
        if (action.parameters.card_id && gameStep.parameters && gameStep.parameters.card_id) {
            return action.parameters.card_id === gameStep.parameters.card_id;
        }
        
        // For play_card actions, also match cost
        if (action.type === 'play_card' && action.parameters.cost && gameStep.parameters && gameStep.parameters.cost) {
            return action.parameters.card_id === gameStep.parameters.card_id && 
                   action.parameters.cost === gameStep.parameters.cost;
        }
    }
    
    // For pass actions, just matching the type is sufficient
    if (action.type === 'pass') {
        return true;
    }
    
    return false;
}

async function truncateGameFromAction(action) {
    try {
        if (!currentGameID) {
            throw new Error('No game loaded');
        }
        
        console.log('Truncating game from action:', action);
        
        if (!confirm(`Are you sure you want to truncate the game at step ${currentStep + 1}? This will permanently delete all steps after this point.`)) {
            return;
        }
        
        // Get the game log up to the current step
        const logUpToCurrentStep = gameSteps.slice(0, currentStep + 1);
        
        // Convert steps back to log format
        const logEntries = logUpToCurrentStep.map(step => {
            return JSON.stringify({
                turn: step.turn,
                player: step.player,
                action: step.action,
                parameters: step.parameters
            });
        });
        
        const logContent = logEntries.join('\n') + '\n';
        
        // Update the current game with truncated log
        const response = await fetch(`/api/games/${currentGameID}/truncate`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                logContent: logContent
            })
        });
        
        const result = await response.json();
        
        if (!response.ok || result.error) {
            throw new Error(result.error || 'Failed to truncate game');
        }
        
        console.log('Game truncated successfully:', result);
        
        // Reload the game to show the truncated state
        await loadGameSteps();
        await renderCurrentStep();
        
    } catch (error) {
        console.error('Failed to truncate game:', error);
        alert(`Failed to truncate game: ${error.message}`);
    }
}

async function forkGameFromAction(action) {
    try {
        if (!currentGameID) {
            throw new Error('No game loaded');
        }
        
        console.log('Forking game from action:', action);
        
        // Create a new game name for the fork
        const timestamp = Date.now();
        const forkGameName = `${currentGameName}-fork-${timestamp}`;
        
        // Get the game log up to the current step
        const logUpToCurrentStep = gameSteps.slice(0, currentStep + 1);
        
        // Convert steps back to log format
        const logEntries = logUpToCurrentStep.map(step => {
            return JSON.stringify({
                turn: step.turn,
                player: step.player,
                action: step.action,
                parameters: step.parameters
            });
        });
        
        const logContent = logEntries.join('\n') + '\n';
        
        // Create the forked game
        const response = await fetch('/api/games', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: forkGameName,
                player1Deck: 'unknown', // Will be extracted from log
                player2Deck: 'unknown', // Will be extracted from log
                logContent: logContent
            })
        });
        
        const result = await response.json();
        
        if (!response.ok || result.error) {
            throw new Error(result.error || 'Failed to create forked game');
        }
        
        console.log('Forked game created:', result);
        
        // Redirect to the new forked game
        window.location.href = `/?game=${forkGameName}`;
        
    } catch (error) {
        console.error('Failed to fork game:', error);
        alert(`Failed to fork game: ${error.message}`);
    }
}

async function appendActionToLog(action) {
    try {
        if (!currentGameID) {
            throw new Error('No game loaded');
        }
        
        console.log('Executing action:', action);
        
        // Send action to backend API
        const response = await fetch(`/api/games/${currentGameID}/execute`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                type: action.type,
                parameters: action.parameters
            })
        });
        
        const result = await response.json();
        
        if (!response.ok || result.error) {
            throw new Error(result.error || 'Failed to execute action');
        }
        
        console.log('Action executed successfully:', result);
        
        // Reload the game state to show the new action
        await loadGameSteps();
        await renderCurrentStep(); // This will call loadAvailableActions internally
        
    } catch (error) {
        console.error('Failed to execute action:', error);
        alert(`Failed to execute action: ${error.message}`);
    }
}

async function renderCurrentStep() {
    if (gameSteps.length === 0) {
        console.log('No game steps to render');
        return;
    }
    
    console.log(`Rendering step ${currentStep + 1} of ${gameSteps.length}`);
    const step = gameSteps[currentStep];
    console.log('Step data:', step);
    
    // Fetch game state for the current step using the original step number
    const stepNumber = step.originalStepNumber || (currentStep + 1);
    const gameState = await fetchGameStateForStep(stepNumber);
    if (!gameState) {
        console.error('Failed to load game state for step');
        return;
    }
    console.log('Game state:', gameState);
    
    const zones = gameState.zones;
    const playerStats = gameState.playerStats;
    console.log('Zones:', zones);
    console.log('Player stats:', playerStats);
    
    // Update step info
    const stepInfo = document.getElementById('stepInfo');
    
    // Format parameters properly
    let paramDisplay = '';
    if (step.parameters && Object.keys(step.parameters).length > 0) {
        const paramEntries = Object.entries(step.parameters);
        paramDisplay = `(${paramEntries.map(([key, value]) => `${key}: ${value}`).join(', ')})`;
    }
    
    stepInfo.innerHTML = `
        <strong>Step ${currentStep + 1}:</strong> Player ${step.player} - ${step.action}
        ${paramDisplay}
    `;
    
    // Render Player 1 zones and stats
    renderPlayerZones(zones.player1, 'p1');
    renderPlayerStats(playerStats.player1, 'p1');
    
    // Render Player 2 zones and stats
    renderPlayerZones(zones.player2, 'p2');
    renderPlayerStats(playerStats.player2, 'p2');
    
    // Render history
    renderGameHistory();
    
    // Refresh available actions when step changes
    loadAvailableActions();
    
    updateStepCounter();
}

function renderPlayerZones(playerZones, playerId) {
    console.log(`Rendering zones for ${playerId}:`, playerZones);
    
    if (!playerZones) {
        console.error(`No player zones data for ${playerId}`);
        return;
    }
    
    // Update zone counts
    const deckElement = document.getElementById(`${playerId}-deck-count`);
    const discardElement = document.getElementById(`${playerId}-discard-count`);
    
    if (deckElement) {
        deckElement.textContent = playerZones.deck || 0;
        console.log(`Updated ${playerId} deck count to: ${playerZones.deck}`);
    } else {
        console.error(`Deck count element not found: ${playerId}-deck-count`);
    }
    
    if (discardElement) {
        discardElement.textContent = playerZones.discard || 0;
        console.log(`Updated ${playerId} discard count to: ${playerZones.discard}`);
        
        // Add mouseover functionality for discard pile
        const discardParent = discardElement.parentElement;
        if (discardParent && playerZones.discardPile && playerZones.discardPile.length > 0) {
            discardParent.style.cursor = 'pointer';
            discardParent.addEventListener('mouseenter', () => showZoneContents('Discard Pile', playerZones.discardPile));
            discardParent.addEventListener('mouseleave', hideZoneContents);
        }
    }
    
    // Update hand count and render face-down cards
    // Hand is now an array of card objects
    const handCount = playerZones.hand ? playerZones.hand.length : 0;
    console.log(`${playerId} hand count:`, handCount, 'hand data:', playerZones.hand);
    
    const handCountElement = document.getElementById(`${playerId}-hand-count`);
    if (handCountElement) {
        handCountElement.textContent = handCount;
        console.log(`Updated ${playerId} hand count to: ${handCount}`);
    } else {
        console.error(`Hand count element not found: ${playerId}-hand-count`);
    }
    
    renderHandCards(`${playerId}-hand`, playerZones.hand || []);
    
    // Render battlefield (face-up cards with instances and exhaustion)
    renderBattlefieldCards(`${playerId}-battlefield`, playerZones.in_play || []);
}

function renderPlayerStats(playerStats, playerId) {
    // Update player statistics
    document.getElementById(`${playerId}-lore-count`).textContent = playerStats.lore;
    
    // Update ink counter (available / total)
    const availableInk = playerStats.available_ink || 0;
    const totalInk = playerStats.total_ink || 0;
    const inkElement = document.getElementById(`${playerId}-ink-count`);
    inkElement.textContent = `${availableInk} / ${totalInk}`;
    
    // Add mouseover functionality for ink counter if there are inked cards
    const inkParent = inkElement.parentElement;
    if (inkParent && playerStats.inkwell && playerStats.inkwell.length > 0) {
        inkParent.style.cursor = 'pointer';
        inkParent.addEventListener('mouseenter', () => showZoneContents('Inkwell', playerStats.inkwell));
        inkParent.addEventListener('mouseleave', hideZoneContents);
    }
}

function renderFaceDownCards(containerId, count) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';
    
    for (let i = 0; i < count; i++) {
        const card = document.createElement('div');
        card.className = 'card';
        card.innerHTML = '<img src="/back.png" alt="Card Back" />';
        container.appendChild(card);
    }
}

function renderHandCards(containerId, handCards) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';
    
    handCards.forEach(cardData => {
        const cardId = cardData.card_id;
        const cardInfo = cardDatabase[cardId];
        
        const card = document.createElement('div');
        card.className = 'card';
        
        if (cardInfo && cardInfo.Image) {
            card.innerHTML = `
                <img src="${cardInfo.Image}" alt="${cardInfo.Name}" />
                <div class="tooltip">${cardInfo.Name}</div>
            `;
            card.addEventListener('mouseenter', () => showCardDetail(cardInfo));
            card.addEventListener('mouseleave', hideCardDetail);
        } else {
            card.innerHTML = `
                <img src="/back.png" alt="Card Back" />
                <div class="tooltip">${cardId} (Image not found)</div>
            `;
        }
        
        container.appendChild(card);
    });
}

function renderInkCards(containerId, inkCards) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';
    
    const currentTurn = gameSteps[currentStep]?.gameState?.gameState?.currentTurn || 0;
    
    inkCards.forEach(inkCard => {
        const card = document.createElement('div');
        card.className = 'card';
        
        // Show face up only on the turn it was played, otherwise show back
        const showFaceUp = inkCard.turnPlayed === currentTurn;
        const cardData = cardDatabase[inkCard.card_id];
        
        if (showFaceUp && cardData && cardData.Image) {
            card.innerHTML = `
                <img src="${cardData.Image}" alt="${cardData.Name}" />
                <div class="tooltip">${cardData.Name}</div>
            `;
            card.addEventListener('mouseenter', () => showCardDetail(cardData));
            card.addEventListener('mouseleave', hideCardDetail);
        } else {
            card.innerHTML = '<img src="/back.png" alt="Card Back" />';
            if (cardData) {
                card.innerHTML += `<div class="tooltip">${cardData.Name} (face down)</div>`;
            }
        }
        
        container.appendChild(card);
    });
}

function renderBattlefieldCards(containerId, battlefieldCards) {
    const container = document.getElementById(containerId);
    container.innerHTML = '';
    
    battlefieldCards.forEach(battlefieldCard => {
        const card = document.createElement('div');
        card.className = 'card';
        
        // Add exhausted class if card is exhausted (rotated 90 degrees)
        if (battlefieldCard.exhausted) {
            card.classList.add('exhausted');
        }
        
        const cardData = cardDatabase[battlefieldCard.card_id];
        if (cardData && cardData.Image) {
            card.innerHTML = `
                <img src="${cardData.Image}" alt="${cardData.Name}" />
                <div class="tooltip">${cardData.Name}${battlefieldCard.exhausted ? ' (Exhausted)' : ''}<br/>Instance: ${battlefieldCard.instance_id}</div>
            `;
            card.addEventListener('mouseenter', () => showCardDetail(cardData));
            card.addEventListener('mouseleave', hideCardDetail);
        } else {
            card.innerHTML = `
                <img src="/back.png" alt="Card Back" />
                <div class="tooltip">${battlefieldCard.card_id}${battlefieldCard.exhausted ? ' (Exhausted)' : ''}<br/>Instance: ${battlefieldCard.instance_id}</div>
            `;
        }
        
        container.appendChild(card);
    });
}

async function nextStep() {
    if (currentStep < gameSteps.length - 1) {
        currentStep++;
        await renderCurrentStep();
    }
}

async function previousStep() {
    if (currentStep > 0) {
        currentStep--;
        await renderCurrentStep();
    }
}

async function seekToPosition(event) {
    const timeline = event.currentTarget;
    const rect = timeline.getBoundingClientRect();
    const clickX = event.clientX - rect.left;
    const percentage = clickX / rect.width;
    
    currentStep = Math.floor(percentage * gameSteps.length);
    if (currentStep >= gameSteps.length) currentStep = gameSteps.length - 1;
    if (currentStep < 0) currentStep = 0;
    
    await renderCurrentStep();
}

function playPause() {
    const button = document.getElementById('playButton');
    
    if (isPlaying) {
        clearInterval(playInterval);
        isPlaying = false;
        button.textContent = 'Play';
    } else {
        isPlaying = true;
        button.textContent = 'Pause';
        
        playInterval = setInterval(() => {
            if (currentStep < gameSteps.length - 1) {
                nextStep();
            } else {
                // Auto-stop at end
                clearInterval(playInterval);
                isPlaying = false;
                button.textContent = 'Play';
            }
        }, 1000); // 1 second per step
    }
}

function showCardDetail(cardData) {
    const detail = document.getElementById('cardDetail');
    
    document.getElementById('cardDetailImage').src = cardData.Image || '';
    document.getElementById('cardDetailName').textContent = cardData.Name || 'Unknown Card';
    document.getElementById('cardDetailCost').textContent = cardData.Cost || '-';
    document.getElementById('cardDetailStrength').textContent = cardData.Strength || '-';
    document.getElementById('cardDetailWillpower').textContent = cardData.Willpower || '-';
    document.getElementById('cardDetailLore').textContent = cardData.Lore || '-';
    document.getElementById('cardDetailType').textContent = cardData.Type || '-';
    document.getElementById('cardDetailColor').textContent = cardData.Color || '-';
    document.getElementById('cardDetailRarity').textContent = cardData.Rarity || '-';
    document.getElementById('cardDetailBodyText').textContent = cardData.Body_Text || '';
    document.getElementById('cardDetailFlavorText').textContent = cardData.Flavor_Text || '';
    
    detail.classList.add('visible');
}

function hideCardDetail() {
    const detail = document.getElementById('cardDetail');
    detail.classList.remove('visible');
}

async function renderGameHistory() {
    const historyContainer = document.getElementById('game-history');
    if (!historyContainer) {
        console.error('game-history element not found');
        return;
    }
    
    historyContainer.innerHTML = '';
    
    if (!gameSteps || gameSteps.length === 0) {
        console.log('No game steps to render in history');
        historyContainer.innerHTML = '<div>No game history available</div>';
        return;
    }
    
    console.log(`Rendering game history: ${currentStep + 1} of ${gameSteps.length} steps`);
    
    try {
        // Fetch history descriptions from server
        const response = await fetch(`/api/games/${currentGameID}/history`);
        if (!response.ok) {
            throw new Error(`Failed to fetch history: ${response.statusText}`);
        }
        
        const result = await response.json();
        const historyDescriptions = result.data || [];
        
        // Show history up to current step
        for (let i = 0; i <= currentStep; i++) {
            const step = gameSteps[i];
            const historyItem = historyDescriptions[i];
            if (!step || !historyItem) continue;
            
            const historyEntry = document.createElement('div');
            historyEntry.className = 'history-entry';
            
            if (i === currentStep) {
                historyEntry.classList.add('current');
            }
            
            historyEntry.innerHTML = `
                <div style="font-weight: bold; margin-bottom: 5px;">Step ${i + 1}</div>
                <div>${historyItem.description}</div>
            `;
            
            // Click to jump to step
            historyEntry.addEventListener('click', async () => {
                currentStep = i;
                await renderCurrentStep();
            });
            
            historyContainer.appendChild(historyEntry);
        }
        
        // Auto-scroll to current step
        const currentEntry = historyContainer.querySelector('.current');
        if (currentEntry) {
            currentEntry.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    } catch (error) {
        console.error('Error rendering game history:', error);
        historyContainer.innerHTML = '<div>Error loading game history</div>';
    }
}

function showZoneContents(zoneName, cards) {
    const zoneTooltip = document.getElementById('zoneTooltip') || createZoneTooltip();
    
    let content = `<div class="zone-tooltip-title">${zoneName}</div>`;
    content += '<div class="zone-tooltip-cards">';
    
    cards.forEach(cardData => {
        const cardId = cardData.cardId || cardData.card_id;
        const cardInfo = cardDatabase[cardId];
        
        if (cardInfo) {
            content += `
                <div class="zone-tooltip-card">
                    <img src="${cardInfo.Image}" alt="${cardInfo.Name}" />
                    <span>${cardInfo.Name}</span>
                </div>
            `;
        }
    });
    
    content += '</div>';
    zoneTooltip.innerHTML = content;
    zoneTooltip.classList.add('visible');
}

function hideZoneContents() {
    const zoneTooltip = document.getElementById('zoneTooltip');
    if (zoneTooltip) {
        zoneTooltip.classList.remove('visible');
    }
}

function createZoneTooltip() {
    const tooltip = document.createElement('div');
    tooltip.id = 'zoneTooltip';
    tooltip.className = 'zone-tooltip';
    document.body.appendChild(tooltip);
    return tooltip;
}

// Keyboard shortcuts
document.addEventListener('keydown', (event) => {
    switch(event.key) {
        case 'ArrowLeft':
            previousStep();
            break;
        case 'ArrowRight':
            nextStep();
            break;
        case ' ':
            event.preventDefault();
            playPause();
            break;
    }
});