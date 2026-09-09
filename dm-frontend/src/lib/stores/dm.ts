import { writable } from 'svelte/store';

export interface Stats {
    nom: string;
    background: string;
    force: number;
    constitution: number;
    vitesse: number;
    charisme: number;
    savoir: number;
    instinct: number;
}

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

export interface Equipment {
    arme: Item | null;
    armure: Item | null;
}

export interface PlayerData {
    nom: string;
    lieu: string;
    pv: number;
    max_pv: number;
    classe: string;
    alignement: string;
    stats: Stats;
    inventaire: Item[];
    sorts: Sort[];
    equipement: Equipment;
    role: boolean;
    is_npc?: boolean;
}

export interface Location {
    nom: string;
    background: string;
    objects: Item[];
    position?: { x: number; y: number };
}

interface DMState {
    me: string | null;
    connected: boolean;
    players: Record<string, PlayerData>;
    npcs: Record<string, PlayerData>;
    locations: Location[];
    logs: string[];
    selectedPlayer: string | null;
}

export const dmState = writable<DMState>({
    me: null,
    connected: false,
    players: {},
    npcs: {},
    locations: [],
    logs: [],
    selectedPlayer: null,
});

let socket: WebSocket | undefined;

export function dmConnect(pseudo: string): void {
    if (socket) socket.close();

    socket = new WebSocket(`ws://${window.location.hostname}:8000/ws/${pseudo}`);

    socket.onopen = () => {
        dmState.update(s => ({ ...s, me: pseudo, connected: true }));
        socket!.send(JSON.stringify({
            type: 'init',
            nom_personnage: pseudo,
            classe: 'Guerrier',
        }));
    };

    socket.onmessage = (event: MessageEvent) => {
        const data = JSON.parse(event.data);
        if (data.type === 'chat') {
            dmState.update(s => ({ ...s, logs: [...s.logs, data.msg] }));
        } else if (data.type === 'sync') {
            dmState.update(s => ({
                ...s,
                players: data.liste || {},
                npcs: data.npcs || {},
                locations: data.locations || s.locations,
            }));
        }
    };

    socket.onclose = () => {
        dmState.update(s => ({ ...s, me: null, connected: false }));
    };
}

export function dmSend(action: Record<string, unknown>): void {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(action));
    }
}

export function selectPlayer(pseudo: string | null): void {
    dmState.update(s => ({ ...s, selectedPlayer: pseudo }));
}
