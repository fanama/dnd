import { writable } from 'svelte/store';

interface PlayerStats {
    nom: string;
    lieu: string;
    pv: number;
    inventaire: any[];
}

interface Location {
    nom: string;
    background: string;
    objects: any[];
}

interface GameState {
    me: string | null;
    location: string;
    players: Record<string, PlayerStats>;
    logs: string[];
    currentLocationObjects: any[];
    locations: Location[];
}

interface SyncData {
    type: 'sync';
    liste: Record<string, PlayerStats>;
    locations?: { nom: string; objects: any[] }[];
}

interface ChatData {
    type: 'chat';
    msg: string;
}

export const gameState = writable<GameState>({
    me: null,
    location: 'En Voyage...',
    players: {},
    logs: [],
    currentLocationObjects: [],
    locations: []
});

let socket: WebSocket | undefined;

export function connect(pseudo: string, charName: string, charClass: string): void {
    if (socket) {
        socket.close();
    }

    socket = new WebSocket(`ws://${window.location.hostname}:8000/ws/${pseudo}`);

    socket.onopen = () => {
        gameState.update(s => ({ ...s, me: pseudo }));
        socket!.send(JSON.stringify({
            type: 'init',
            nom_personnage: charName,
            classe: charClass
        }));
    };

    socket.onmessage = (event: MessageEvent) => {
        const data: SyncData | ChatData = JSON.parse(event.data);
        if (data.type === 'chat') {
            gameState.update(s => ({ ...s, logs: [...s.logs, data.msg] }));
        } else if (data.type === 'sync') {
            gameState.update(s => {
                const myStats = data.liste[pseudo];

                let currentLocationObjects: any[] = [];
                if (data.locations) {
                    const loc = data.locations.find(l => l.nom === (myStats ? myStats.lieu : s.location));
                    if (loc) currentLocationObjects = loc.objects;
                }

                return {
                    ...s,
                    players: data.liste,
                    location: myStats ? myStats.lieu : s.location,
                    currentLocationObjects: currentLocationObjects,
                    locations: data.locations || s.locations
                };
            });
        }
    };

    socket.onclose = () => {
        gameState.update(s => ({ ...s, me: null }));
    };
}

export function sendAction(action: Record<string, unknown>): void {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(action));
    }
}
