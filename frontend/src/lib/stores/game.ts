import { writable, derived } from 'svelte/store';

export interface Item {
    nom: string;
    prix: number;
    encombrement: number;
    isConsumable: boolean;
    bonusDegats: number;
    bonusArmure: number;
}

export interface Sort {
    nom: string;
    niveauSort: number;
    ecoleMagie: string;
    portee: string;
    duree: string;
}

export interface PlayerStats {
    nom: string;
    lieu: string;
    pv: number;
    max_pv: number;
    classe: string;
    alignement: string;
    stats: {
        force: number;
        constitution: number;
        vitesse: number;
        charisme: number;
        savoir: number;
        instinct: number;
    };
    inventaire: Item[];
    sorts: Sort[];
}

export interface Location {
    nom: string;
    background: string;
    objects: Item[];
}

export interface GameState {
    me: string | null;
    location: string;
    players: Record<string, PlayerStats>;
    npcs: any[];
    logs: string[];
    currentLocationObjects: Item[];
    locations: Location[];
}

export interface SyncData {
    type: 'sync';
    liste: Record<string, PlayerStats>;
    npcs?: any;
    locations?: { nom: string; objects: Item[] }[];
}

interface ChatData {
    type: 'chat';
    msg: string;
}

export function getDerivedStats(s: PlayerStats | undefined) {
    if (!s || !s.stats) {
        return { maxPv: 100, armor: 0, damage: 0 };
    }
    const maxPv = Math.floor((s.stats.constitution || 10) * 10);
    const armor = Math.floor((s.stats.vitesse || 10) * 1.5);
    const damage = Math.floor((s.stats.force || 10) * 2);
    return { maxPv, armor, damage };
}

export function getItemCategory(item: Item | string): 'weapon' | 'armor' | 'consumable' | 'misc' {
    if (typeof item === 'string') return 'misc';
    if (item.isConsumable) return 'consumable';
    if (item.bonusDegats > 0) return 'weapon';
    if (item.bonusArmure > 0) return 'armor';
    return 'misc';
}

export const gameState = writable<GameState>({
    me: null,
    location: 'En Voyage...',
    players: {},
    npcs: [],
    logs: [],
    currentLocationObjects: [],
    locations: []
});

export const myStats = derived(gameState, ($gs) => {
    return $gs.me ? $gs.players[$gs.me] : undefined;
});

export const myDerivedStats = derived(myStats, ($ms) => {
    return getDerivedStats($ms);
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

                let currentLocationObjects: Item[] = [];
                if (data.locations) {
                    const loc = data.locations.find(l => l.nom === (myStats ? myStats.lieu : s.location));
                    if (loc) currentLocationObjects = loc.objects;
                }

                const npcList = data.npcs ? Object.values(data.npcs) : [];
                const currentNpcs = npcList.filter((n: any) => n.lieu === (myStats ? myStats.lieu : s.location));

                return {
                    ...s,
                    players: data.liste,
                    npcs: currentNpcs,
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
