import { writable } from 'svelte/store';

export const gameState = writable({
    me: null,
    location: 'En Voyage...',
    players: {},
    logs: [],
    currentLocationObjects: []
});

let socket;

export function connect(pseudo, charName, charClass) {
    // Fermer toute connexion existante pour éviter les doublons
    if (socket) {
        socket.close();
    }

    socket = new WebSocket(`ws://${window.location.hostname}:8000/ws/${pseudo}`);

    socket.onopen = () => {
        gameState.update(s => ({ ...s, me: pseudo }));
        // On envoie les paramètres, mais le serveur doit décider 
        // s'il les utilise (nouveau perso) ou s'il restaure l'existant (reconnaissance du pseudo)
        socket.send(JSON.stringify({
            type: 'init',
            nom_personnage: charName,
            classe: charClass
        }));
    };

    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'chat') {
            gameState.update(s => ({ ...s, logs: [...s.logs, data.msg] }));
        } else if (data.type === 'sync') {
            gameState.update(s => {
                const myStats = data.liste[pseudo];
                
                let currentLocationObjects = [];
                if (data.locations) {
                    const loc = data.locations.find(l => l.nom === (myStats ? myStats.lieu : s.location));
                    if (loc) currentLocationObjects = loc.objects;
                }

                return {
                    ...s,
                    players: data.liste,
                    location: myStats ? myStats.lieu : s.location,
                    currentLocationObjects: currentLocationObjects
                };
            });
        }
    };

    socket.onclose = () => {
        gameState.update(s => ({ ...s, me: null }));
    };
}

export function sendAction(action) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(action));
    }
}
