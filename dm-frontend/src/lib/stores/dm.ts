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

export interface DerivedCombat {
    ac: number;
    attack_mod: number;
    damage_mod: number;
    damage_dice: string;
    hit_dice: number;
    weapon_name: string;
    ranged: boolean;
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
    combat?: DerivedCombat;
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
    selectedPlayer: string | null;
    latestExport: unknown;
}

export const dmState = writable<DMState>({
    me: null,
    connected: false,
    players: {},
    npcs: {},
    locations: [],
    selectedPlayer: null,
    latestExport: null,
});

// The journal lives in its own capped store: a long session must not grow an
// unbounded array that every message copies and re-renders along with the
// whole DM state.
export const DM_LOG_LIMIT = 300;
export const dmLogs = writable<string[]>([]);
// Counts entries dropped from the front of the capped log so the journal
// keeps numbering from 1 across the whole session.
export const dmLogOffset = writable(0);

function appendLog(msg: string): void {
    let dropped = 0;
    dmLogs.update((logs) => {
        const next = [...logs, msg];
        if (next.length > DM_LOG_LIMIT) {
            dropped = next.length - DM_LOG_LIMIT;
            return next.slice(dropped);
        }
        return next;
    });
    if (dropped > 0) {
        dmLogOffset.update((o) => o + dropped);
    }
}

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

// Syncs are coalesced server-side, but bursts can still arrive in one
// frame: apply only the latest sync per animation frame.
let pendingSync: any = null;
let syncFrame: number | null = null;

function scheduleSync(data: any): void {
    pendingSync = data;
    if (syncFrame !== null) {
        return;
    }
    syncFrame = requestAnimationFrame(() => {
        syncFrame = null;
        const pending = pendingSync;
        pendingSync = null;
        if (pending) applySync(pending);
    });
}

function cancelScheduledSync(): void {
    if (syncFrame !== null) {
        cancelAnimationFrame(syncFrame);
        syncFrame = null;
    }
    pendingSync = null;
}

function applySync(data: any): void {
    dmState.update(s => ({
        ...s,
        players: data.liste || {},
        npcs: data.npcs || {},
        locations: data.locations || s.locations,
    }));
}

function handleMessage(event: MessageEvent): void {
    const data = JSON.parse(event.data);
    if (data.type === 'chat') {
        appendLog(data.msg);
    } else if (data.type === 'sync') {
        scheduleSync(data);
    } else if (data.type === 'locations') {
        // Locations travel separately and only when they changed.
        dmState.update(s => ({ ...s, locations: data.locations || s.locations }));
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
    cancelScheduledSync();
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
