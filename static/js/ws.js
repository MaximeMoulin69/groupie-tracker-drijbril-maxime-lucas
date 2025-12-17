class GameWebSocket {
    constructor(roomCode) {
        this.roomCode = roomCode;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.messageHandlers = {};
        this.isConnected = false;
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?room=${this.roomCode}`;

        console.log('Connexion WebSocket...', wsUrl);
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket connecte');
            this.isConnected = true;
            this.reconnectAttempts = 0;

            this.send('player_connected', {
                message: 'Joueur connecte a la salle'
            });
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                console.log('Message recu:', message);
                this.handleMessage(message);
            } catch (error) {
                console.error('Erreur parsing message:', error);
            }
        };

        this.ws.onerror = (error) => {
            console.error('Erreur WebSocket:', error);
        };

        this.ws.onclose = () => {
            console.log('WebSocket deconnecte');
            this.isConnected = false;
            this.attemptReconnect();
        };
    }

    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 10000);
            
            console.log(`Reconnexion ${this.reconnectAttempts}/${this.maxReconnectAttempts} dans ${delay}ms`);

            setTimeout(() => {
                this.connect();
            }, delay);
        } else {
            console.error('Nombre maximum de tentatives atteint');
        }
    }

    send(type, content) {
        if (!this.isConnected || this.ws.readyState !== WebSocket.OPEN) {
            console.warn('WebSocket non connecte');
            return false;
        }

        const message = {
            type: type,
            content: content,
            timestamp: Date.now()
        };

        this.ws.send(JSON.stringify(message));
        console.log('Message envoye:', message);
        return true;
    }

    handleMessage(message) {
        const { type, from, user_id, content } = message;

        if (this.messageHandlers[type]) {
            this.messageHandlers[type](message);
        }

        switch (type) {
            case 'player_connected':
                this.onPlayerConnected(from, user_id);
                break;
            case 'player_disconnected':
                this.onPlayerDisconnected(from, user_id);
                break;
            case 'chat':
                this.onChatMessage(from, content);
                break;
            case 'game_start':
                this.onGameStart(content);
                break;
            case 'round_start':
                this.onRoundStart(content);
                break;
            case 'answer_submitted':
                this.onAnswerSubmitted(from, content);
                break;
            case 'round_end':
                this.onRoundEnd(content);
                break;
            case 'game_end':
                this.onGameEnd(content);
                break;
            case 'scoreboard_update':
                this.onScoreboardUpdate(content);
                break;
            case 'category_added':
                this.onCategoryAdded(content);
                break;
            case 'category_deleted':
                this.onCategoryDeleted(content);
                break;
        }
    }

    on(messageType, handler) {
        this.messageHandlers[messageType] = handler;
    }

    onPlayerConnected(pseudo, userId) {
        console.log(`${pseudo} a rejoint`);
        this.addNotification(`${pseudo} a rejoint la partie`, 'info');
        
        const playersList = document.querySelector('.players-list');
        if (playersList) {
            const li = document.createElement('li');
            li.textContent = pseudo;
            playersList.appendChild(li);
        }
    }

    onPlayerDisconnected(pseudo, userId) {
        console.log(`${pseudo} a quitte`);
        this.addNotification(`${pseudo} a quitte la partie`, 'warning');
    }

    onChatMessage(from, content) {
        console.log(`${from}: ${content}`);
        this.addChatMessage(from, content);
    }

    onGameStart(content) {
        console.log('Partie commence !');
        this.addNotification('La partie commence !', 'success');
        this.hideWaitingRoom();
        this.showGameInterface();
    }

    onRoundStart(content) {
        console.log('Nouveau tour:', content);
        this.addNotification(`Tour ${content.roundNumber}/${content.totalRounds}`, 'info');
        
        if (content.letter) {
            const letterDisplay = document.getElementById('current-letter');
            const gameLetter = document.getElementById('game-letter');
            if (letterDisplay) letterDisplay.textContent = `Lettre : ${content.letter}`;
            if (gameLetter) gameLetter.textContent = content.letter;
        }
    }

    onAnswerSubmitted(from, content) {
        console.log(`${from} a repondu`);
        this.showPlayerAnswered(from);
    }

    onRoundEnd(content) {
        console.log('Fin du tour');
        this.addNotification('Fin du tour !', 'info');
        this.showRoundResults(content);
    }

    onGameEnd(content) {
        console.log('Fin de la partie');
        this.addNotification('Partie terminee !', 'success');
        this.showFinalScoreboard(content.scoreboard);
    }

    onScoreboardUpdate(content) {
        console.log('Mise a jour scoreboard');
        this.updateScoreboard(content);
    }

    onCategoryAdded(content) {
        console.log('Categorie ajoutee:', content.category);
        this.addCategoryTag(content.category);
    }

    onCategoryDeleted(content) {
        console.log('Categorie supprimee:', content.category);
    }

    addNotification(message, type) {
        const notifContainer = document.getElementById('notifications');
        if (!notifContainer) return;

        const notif = document.createElement('div');
        notif.className = `notification notification-${type}`;
        notif.textContent = message;
        notifContainer.appendChild(notif);

        setTimeout(() => notif.classList.add('show'), 10);

        setTimeout(() => {
            notif.classList.remove('show');
            setTimeout(() => notif.remove(), 300);
        }, 3000);
    }

    addChatMessage(from, content) {
        const chatContainer = document.getElementById('chat-messages');
        if (!chatContainer) return;

        const messageDiv = document.createElement('div');
        messageDiv.className = 'chat-message';
        messageDiv.innerHTML = `<strong>${from}:</strong> ${content}`;
        chatContainer.appendChild(messageDiv);

        chatContainer.scrollTop = chatContainer.scrollHeight;
    }

    hideWaitingRoom() {
        const waitingRoom = document.getElementById('waiting-room');
        if (waitingRoom) {
            waitingRoom.style.display = 'none';
        }
    }

    showGameInterface() {
        const gameInterface = document.getElementById('game-interface');
        if (gameInterface) {
            gameInterface.style.display = 'block';
        }
    }

    showPlayerAnswered(pseudo) {
        const playerElement = document.querySelector(`[data-player="${pseudo}"]`);
        if (playerElement) {
            playerElement.classList.add('answered');
        }
    }

    showRoundResults(results) {
        const resultsContainer = document.getElementById('round-results');
        if (resultsContainer) {
            resultsContainer.innerHTML = JSON.stringify(results, null, 2);
            resultsContainer.style.display = 'block';
        }
    }

    updateScoreboard(scoreboardData) {
        const scoreboardElement = document.getElementById('scoreboard');
        if (!scoreboardElement) return;

        let html = '<h3>Scores</h3><div class="scoreboard-list">';
        
        scoreboardData.forEach((entry, index) => {
            html += `
                <div class="scoreboard-entry">
                    <span class="rank">${index + 1}.</span>
                    <span class="pseudo">${entry.Pseudo}</span>
                    <span class="score">${entry.Score} pts</span>
                </div>
            `;
        });

        html += '</div>';
        scoreboardElement.innerHTML = html;
    }

    showFinalScoreboard(scoreboardData) {
        const finalScoreboard = document.getElementById('final-scoreboard');
        if (!finalScoreboard) return;

        let html = '<h2>Scoreboard Final</h2><div class="final-scoreboard-list">';

        scoreboardData.forEach((entry, index) => {
            const medal = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : '';
            html += `
                <div class="final-entry rank-${index + 1}">
                    <span class="medal">${medal}</span>
                    <span class="rank">${index + 1}.</span>
                    <span class="pseudo">${entry.Pseudo}</span>
                    <span class="score">${entry.Score} points</span>
                </div>
            `;
        });

        html += '</div>';
        finalScoreboard.innerHTML = html;
        finalScoreboard.style.display = 'block';
    }

    addCategoryTag(categoryName) {
        const customCategoriesDiv = document.getElementById('custom-categories');
        if (!customCategoriesDiv) return;

        const tag = document.createElement('span');
        tag.className = 'category-tag';
        tag.innerHTML = `
            ${categoryName}
            <button type="button" onclick="removeCategory('${categoryName}')">×</button>
        `;
        customCategoriesDiv.appendChild(tag);
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            console.log('Deconnexion WebSocket');
        }
    }
}

function submitBlindTestAnswer(gameWs, answer) {
    gameWs.send('answer_submitted', {
        answer: answer,
        timestamp: Date.now()
    });
}

function submitPetitBacAnswers(gameWs, answers) {
    gameWs.send('answers_submitted', {
        answers: answers,
        timestamp: Date.now()
    });
}

function votePetitBacAnswer(gameWs, category, answer, isValid) {
    gameWs.send('validation_vote', {
        category: category,
        answer: answer,
        isValid: isValid
    });
}

function startGame(gameWs) {
    gameWs.send('game_start', {
        message: 'L\'hote a demarre la partie'
    });
}

function sendChatMessage(gameWs, message) {
    gameWs.send('chat', message);
}

function addCategory(categoryName) {
    if (!categoryName || !gameWebSocket) return;

    gameWebSocket.send('add_category', {
        category: categoryName
    });
}

function removeCategory(categoryName) {
    if (!gameWebSocket) return;

    gameWebSocket.send('delete_category', {
        category: categoryName
    });

    const tags = document.querySelectorAll('.category-tag');
    tags.forEach(tag => {
        if (tag.textContent.includes(categoryName)) {
            tag.remove();
        }
    });
}

let gameWebSocket = null;

document.addEventListener('DOMContentLoaded', () => {
    const roomCode = document.body.dataset.roomCode || getRoomCodeFromURL();
    
    if (roomCode) {
        console.log('Initialisation WebSocket pour salle:', roomCode);
        gameWebSocket = new GameWebSocket(roomCode);
        gameWebSocket.connect();

        setupChatForm();
        setupStartButton();
        setupAnswerForms();
        setupCategoryCRUD();
    }
});

window.addEventListener('beforeunload', () => {
    if (gameWebSocket) {
        gameWebSocket.disconnect();
    }
});

function getRoomCodeFromURL() {
    const path = window.location.pathname;
    const match = path.match(/\/room\/([a-f0-9]+)/);
    return match ? match[1] : null;
}

function setupChatForm() {
    const chatForm = document.getElementById('chat-form');
    const chatInput = document.getElementById('chat-input');

    if (chatForm && chatInput) {
        chatForm.addEventListener('submit', (e) => {
            e.preventDefault();
            const message = chatInput.value.trim();
            if (message && gameWebSocket) {
                sendChatMessage(gameWebSocket, message);
                chatInput.value = '';
            }
        });
    }
}

function setupStartButton() {
    const startButton = document.getElementById('start-game-btn');
    if (startButton) {
        startButton.addEventListener('click', (e) => {
            e.preventDefault();
            console.log('Clic sur demarrer');
            if (gameWebSocket) {
                startGame(gameWebSocket);
                startButton.disabled = true;
                console.log('Jeu demarre');
            }
        });
    }
}

function setupAnswerForms() {
    const blindtestForm = document.getElementById('blindtest-answer-form');
    if (blindtestForm) {
        blindtestForm.addEventListener('submit', (e) => {
            e.preventDefault();
            const answer = document.getElementById('blindtest-answer').value.trim();
            if (answer && gameWebSocket) {
                submitBlindTestAnswer(gameWebSocket, answer);
                document.getElementById('blindtest-answer').value = '';
                blindtestForm.querySelector('input').disabled = true;
                blindtestForm.querySelector('button').disabled = true;
            }
        });
    }

    const petitbacForm = document.getElementById('petitbac-answer-form');
    if (petitbacForm) {
        petitbacForm.addEventListener('submit', (e) => {
            e.preventDefault();
            
            const answers = {
                artiste: document.getElementById('answer-artiste')?.value.trim() || '',
                album: document.getElementById('answer-album')?.value.trim() || '',
                groupe: document.getElementById('answer-groupe')?.value.trim() || '',
                instrument: document.getElementById('answer-instrument')?.value.trim() || '',
                featuring: document.getElementById('answer-featuring')?.value.trim() || ''
            };

            if (gameWebSocket) {
                submitPetitBacAnswers(gameWebSocket, answers);
                petitbacForm.querySelectorAll('input').forEach(input => input.disabled = true);
                petitbacForm.querySelector('button').disabled = true;
            }
        });
    }
}

function setupCategoryCRUD() {
    const checkboxes = document.querySelectorAll('.category-checkbox input[type="checkbox"]');
    const selectedCountSpan = document.getElementById('selected-count');
    
    if (!checkboxes.length) return;

    function updateSelectedCount() {
        const checked = document.querySelectorAll('.category-checkbox input[type="checkbox"]:checked');
        const count = checked.length;
        
        if (selectedCountSpan) {
            selectedCountSpan.textContent = count;
            
            if (count > 5) {
                selectedCountSpan.style.color = '#FF0000';
            } else if (count < 5) {
                selectedCountSpan.style.color = '#FFAA00';
            } else {
                selectedCountSpan.style.color = '#00D4FF';
            }
        }

        if (count >= 5) {
            checkboxes.forEach(cb => {
                if (!cb.checked) {
                    cb.disabled = true;
                    cb.parentElement.style.opacity = '0.5';
                    cb.parentElement.style.cursor = 'not-allowed';
                }
            });
        } else {
            checkboxes.forEach(cb => {
                cb.disabled = false;
                cb.parentElement.style.opacity = '1';
                cb.parentElement.style.cursor = 'pointer';
            });
        }
    }

    checkboxes.forEach(cb => {
        cb.addEventListener('change', updateSelectedCount);
    });

    updateSelectedCount();

    const configForm = document.getElementById('config-form');
    if (configForm) {
        configForm.addEventListener('submit', (e) => {
            e.preventDefault();
            console.log('Submit config form');

            const selectedCategories = [];
            document.querySelectorAll('.category-checkbox input[type="checkbox"]:checked').forEach(cb => {
                selectedCategories.push(cb.value);
            });

            if (selectedCategories.length !== 5) {
                alert('Tu dois selectionner exactement 5 categories !');
                return;
            }

            if (gameWebSocket) {
                gameWebSocket.send('categories_selected', {
                    categories: selectedCategories
                });
                
                startGame(gameWebSocket);
                console.log('Jeu demarre avec categories:', selectedCategories);
            }
        });
    }
}




