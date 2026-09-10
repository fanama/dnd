import { writable } from 'svelte/store';

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting';

export const connectionStatus = writable<ConnectionStatus>('disconnected');

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
    desDegats?: string;
}

export interface Sort {
    nom: string;
    niveauSort: number;
    ecoleMagie: string;
    portee: string;
    duree: string;
    bonus?: number;
    desDegats?: string;
    buff?: { stat: string; valeur: number };
}

export interface Equipment {
    arme: Item | null;
    armure: Item | null;
}

export interface Quest {
    nom: string;
    objectif: string;
    obstacle: string;
    recompense: Item[];
    information?: string;
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
    quests?: Quest[];
    role: boolean;
    is_npc?: boolean;
    mobType?: string;
}

export interface Location {
    nom: string;
    background: string;
    objects: Item[];
    quests?: Quest[];
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
    latestExport: unknown;
}

export const dmState = writable<DMState>({
    me: null,
    connected: false,
    players: {},
    npcs: {},
    locations: [],
    logs: [],
    selectedPlayer: null,
    latestExport: null,
});

let socket: WebSocket | undefined;
let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
let reconnectAttempts = 0;
let intentionalClose = false;

const MAX_RECONNECT_ATTEMPTS = 10;
const BASE_DELAY_MS = 1000;
const MAX_DELAY_MS = 30000;

let savedPseudo = '';

function wsUrl(pseudo: string): string {
    const configured = import.meta.env.VITE_WS_URL as string | undefined;
    if (configured) return `${configured.replace(/\/$/, '')}/${pseudo}`;
    if (import.meta.env.PROD) {
        const scheme = window.location.protocol === 'https:' ? 'wss' : 'ws';
        const port = window.location.port ? `:${window.location.port}` : '';
        return `${scheme}://${window.location.hostname}${port}/ws/${pseudo}`;
    }
    return `ws://${window.location.hostname}:8000/ws/${pseudo}`;
}

function clearReconnectTimer(): void {
    if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = undefined;
    }
}

function scheduleReconnect(): void {
    clearReconnectTimer();

    if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
        connectionStatus.set('disconnected');
        return;
    }

    connectionStatus.set('reconnecting');

    const delay = Math.min(BASE_DELAY_MS * Math.pow(2, reconnectAttempts), MAX_DELAY_MS);
    reconnectAttempts++;

    reconnectTimer = setTimeout(() => {
        openSocket(savedPseudo, true);
    }, delay);
}

function handleMessage(event: MessageEvent): void {
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
    } else if (data.type === 'state_export') {
        dmState.update(s => ({ ...s, latestExport: data.payload }));
    }
}

function openSocket(pseudo: string, isReconnect: boolean): void {
    if (socket) {
        socket.onopen = null;
        socket.onmessage = null;
        socket.onclose = null;
        socket.onerror = null;
        if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
            socket.close();
        }
    }

    connectionStatus.set(isReconnect ? 'reconnecting' : 'connecting');

    socket = new WebSocket(wsUrl(pseudo));

    socket.onopen = () => {
        reconnectAttempts = 0;
        connectionStatus.set('connected');
        dmState.update(s => ({ ...s, me: pseudo, connected: true }));
        socket!.send(JSON.stringify({
            type: 'init',
            nom_personnage: pseudo,
            classe: 'Guerrier',
        }));
    };

    socket.onmessage = (event: MessageEvent) => {
        handleMessage(event);
    };

    socket.onclose = () => {
        if (intentionalClose) {
            connectionStatus.set('disconnected');
            return;
        }
        scheduleReconnect();
    };

    socket.onerror = () => {
        // onclose will fire after onerror, which triggers reconnection
    };
}

export function dmConnect(pseudo: string): void {
    intentionalClose = false;
    clearReconnectTimer();
    reconnectAttempts = 0;
    savedPseudo = pseudo;
    openSocket(pseudo, false);
}

export function dmDisconnect(): void {
    intentionalClose = true;
    clearReconnectTimer();
    reconnectAttempts = 0;
    if (socket) {
        socket.onopen = null;
        socket.onmessage = null;
        socket.onclose = null;
        socket.onerror = null;
        socket.close();
        socket = undefined;
    }
    connectionStatus.set('disconnected');
    dmState.update(s => ({ ...s, me: null, connected: false }));
}

export function dmSend(action: Record<string, unknown>): void {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(action));
    }
}

export function selectPlayer(pseudo: string | null): void {
    dmState.update(s => ({ ...s, selectedPlayer: pseudo }));
}

export function dmRequestExport(): void {
    dmSend({ type: 'dm_export_state' });
}

export function dmLoadStateFromJson(json: string): void {
    dmSend({ type: 'dm_load_state', payload: json });
}
