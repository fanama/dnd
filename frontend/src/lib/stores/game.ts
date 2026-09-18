import { writable, derived } from 'svelte/store';
import { pushToast } from './toasts';

export type GameEventKind = 'damage' | 'heal' | 'death' | 'loot' | 'quest' | 'buff' | 'xp' | 'level' | 'gold' | 'buy' | 'sell';

export interface GameEvent {
    id: number;
    type: 'event';
    event: GameEventKind;
    source?: string;
    target?: string;
    amount?: number;
    crit?: boolean;
    level?: number;
    lootCount?: number;
    text?: string;
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

export interface SortBuff {
    stat: string;
    valeur: number;
}

export interface Sort {
    nom: string;
    niveauSort: number;
    ecoleMagie: string;
    portee: string;
    duree: string;
    bonus?: number;
    desDegats?: string;
    buff?: SortBuff;
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

export interface PlayerStats {
    nom: string;
    lieu: string;
    pv: number;
    max_pv: number;
    classe: string;
    alignement: string;
    or?: number;
    xp?: number;
    niveau?: number;
    encombrement?: number;
    capacite?: number;
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
    equipement: Equipment;
    quests: Quest[];
    combat: DerivedCombat;
}

export interface Location {
    nom: string;
    background: string;
    objects: Item[];
    quests?: Quest[];
    commerce?: Item[];
    liens?: string[];
}

export interface GameState {
    me: string | null;
    location: string;
    players: Record<string, PlayerStats>;
    npcs: any[];
    logs: string[];
    currentLocationObjects: Item[];
    currentLocationQuests: Quest[];
    locations: Location[];
    newChar: boolean;
}

export interface SyncData {
    type: 'sync';
    liste: Record<string, PlayerStats>;
    npcs?: any;
    locations?: { nom: string; objects: Item[]; quests?: Quest[]; commerce?: Item[]; liens?: string[] }[];
}

interface ChatData {
    type: 'chat';
    msg: string;
}

interface DeathOverlayState {
    active: boolean;
    at?: number;
}

// gameEvents is a rolling feed of structured combat/inventory events consumed
// by the floating-damage layer (and kept small to bound memory).
export const gameEvents = writable<GameEvent[]>([]);

export const deathOverlay = writable<DeathOverlayState>({ active: false });

let eventId = 0;

function pushEvent(ev: Omit<GameEvent, 'id' | 'type'>): void {
    const e: GameEvent = { id: ++eventId, type: 'event', ...ev };
    gameEvents.update((list) => [...list.slice(-24), e]);
}

// handleGameEvent routes structured backend events to the feedback channels:
// floating numbers, toasts, the death overlay and stat-based updates.
function handleGameEvent(me: string, ev: GameEvent): void {
    switch (ev.event) {
        case 'damage':
        case 'heal':
            pushEvent({ event: ev.event, target: ev.target, amount: ev.amount, crit: ev.crit, source: ev.source, text: ev.text });
            break;
        case 'death':
            pushEvent({ event: 'death', target: ev.target, source: ev.source, text: ev.text });
            if (ev.target === me) {
                deathOverlay.set({ active: true, at: Date.now() });
            } else {
                pushToast('warn', `☠️ ${ev.target || 'Un héros'} est tombé !`, ev.text || '');
            }
            break;
        case 'loot':
            pushToast('success', `🎒 ${ev.target} a ramassé : ${ev.text || ''}`);
            break;
        case 'gold':
            pushToast('success', `💰 +${Math.round(ev.amount || 0)} or`, ev.text || '');
            break;
        case 'buy':
            pushToast('info', `🛒 ${ev.text || 'Achat effectué'}`, ev.amount ? `-${Math.round(ev.amount)} or` : '');
            break;
        case 'xp':
            pushToast('info', `✨ +${Math.round(ev.amount || 0)} XP`);
            break;
        case 'level':
            pushToast('success', `🌟 Niveau ${ev.level || '?'} !`, ev.text || '');
            break;
        case 'buff':
            pushToast('info', `✨ ${ev.target || ev.source} : ${ev.text || 'buff'}`);
            break;
        case 'quest':
            pushToast('info', `📜 ${ev.text || 'Quête mise à jour'}`);
            break;
    }
}

export function abilityModifier(stat: number): number {
    return Math.floor(((stat || 10) - 10) / 2);
}

export interface DerivedStats {
    maxPv: number;
    ac: number;
    attackMod: number;
    damageMod: number;
    damageDice: string;
    weaponName: string;
    ranged: boolean;
}

const DEFAULT_DERIVED_STATS: DerivedStats = {
    maxPv: 10,
    ac: 10,
    attackMod: 0,
    damageMod: 0,
    damageDice: '1d2',
    weaponName: 'Mains nues',
    ranged: false,
};

// serverDerivedStats maps the combat statistics precomputed by the backend
// (single source of truth) onto the DerivedStats shape used by the UI.
function serverDerivedStats(s: PlayerStats | undefined): DerivedStats {
    if (!s) return DEFAULT_DERIVED_STATS;
    const c = s.combat;
    return {
        maxPv: s.max_pv || 10,
        ac: c?.ac ?? 10,
        attackMod: c?.attack_mod ?? 0,
        damageMod: c?.damage_mod ?? 0,
        damageDice: c?.damage_dice || '1d2',
        weaponName: c?.weapon_name || 'Mains nues',
        ranged: c?.ranged ?? false,
    };
}

export function getItemCategory(item: Item | string): 'weapon' | 'armor' | 'consumable' | 'misc' {
    if (typeof item === 'string') return 'misc';
    if (item.isConsumable) return 'consumable';
    if (item.bonusDegats > 0 || item.desDegats) return 'weapon';
    if (item.bonusArmure > 0) return 'armor';
    return 'misc';
}

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting';

export const connectionStatus = writable<ConnectionStatus>('disconnected');

// loginError surfaces connection failures on the login screen.
export const loginError = writable('');

export const gameState = writable<GameState>({
    me: null,
    location: 'En Voyage...',
    players: {},
    npcs: [],
    logs: [],
    currentLocationObjects: [],
    currentLocationQuests: [],
    locations: [],
    newChar: false
});

export const myStats = derived(gameState, ($gs) => {
    return $gs.me ? $gs.players[$gs.me] : undefined;
});

export const myDerivedStats = derived(myStats, ($ms) => {
    return serverDerivedStats($ms);
});

let socket: WebSocket | undefined;
let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
let reconnectAttempts = 0;
let intentionalClose = false;

const MAX_RECONNECT_ATTEMPTS = 10;
const BASE_DELAY_MS = 1000;
const MAX_DELAY_MS = 30000;

let savedPseudo = '';
let savedCharName = '';
let savedCharClass = '';
let hasConnected = false;

const SESSION_KEY = 'dnd_session';

function persistSession(): void {
    try {
        sessionStorage.setItem(SESSION_KEY, JSON.stringify({
            pseudo: savedPseudo,
            charName: savedCharName,
            charClass: savedCharClass,
        }));
    } catch {}
}

function restoreSession(): { pseudo: string; charName: string; charClass: string } | null {
    try {
        const raw = sessionStorage.getItem(SESSION_KEY);
        if (!raw) return null;
        const data = JSON.parse(raw);
        if (data?.pseudo && data?.charName && data?.charClass) return data;
    } catch {}
    return null;
}

function clearSession(): void {
    try {
        sessionStorage.removeItem(SESSION_KEY);
    } catch {}
}

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
        loginError.set('Le serveur de jeu est injoignable. Vérifiez votre connexion puis réessayez.');
        return;
    }

    connectionStatus.set('reconnecting');

    const delay = Math.min(BASE_DELAY_MS * Math.pow(2, reconnectAttempts), MAX_DELAY_MS);
    reconnectAttempts++;

    reconnectTimer = setTimeout(() => {
        openSocket(savedPseudo, savedCharName, savedCharClass, true);
    }, delay);
}

function handleMessage(pseudo: string, event: MessageEvent): void {
    const data: SyncData | ChatData | GameEvent | any = JSON.parse(event.data);
    if (data.type === 'chat') {
        gameState.update(s => ({ ...s, logs: [...s.logs, data.msg] }));
    } else if (data.type === 'init_new_char') {
        gameState.update(s => ({ ...s, newChar: true }));
    } else if (data.type === 'event') {
        handleGameEvent(pseudo, data as GameEvent);
    } else if (data.type === 'sync') {
        gameState.update(s => {
            const myStats = data.liste[pseudo];

            let currentLocationObjects: Item[] = [];
            let currentLocationQuests: Quest[] = [];
            if (data.locations) {
                const loc = data.locations.find(l => l.nom === (myStats ? myStats.lieu : s.location));
                if (loc) {
                    currentLocationObjects = loc.objects;
                    currentLocationQuests = loc.quests || [];
                }
            }

            const npcList = data.npcs ? Object.values(data.npcs) : [];
            const currentNpcs = npcList.filter((n: any) => n.lieu === (myStats ? myStats.lieu : s.location));

            return {
                ...s,
                players: data.liste,
                npcs: currentNpcs,
                location: myStats ? myStats.lieu : s.location,
                currentLocationObjects: currentLocationObjects,
                currentLocationQuests: currentLocationQuests,
                locations: data.locations || s.locations
            };
        });
    }
}

function openSocket(pseudo: string, charName: string, charClass: string, isReconnect: boolean): void {
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
    loginError.set('');
    hasConnected = false;

    socket = new WebSocket(wsUrl(pseudo));

    socket.onopen = () => {
        reconnectAttempts = 0;
        hasConnected = true;
        connectionStatus.set('connected');
        deathOverlay.set({ active: false });
        gameState.update(s => ({ ...s, me: pseudo }));
        socket!.send(JSON.stringify({
            type: 'init',
            nom_personnage: charName,
            classe: charClass
        }));
    };

    socket.onmessage = (event: MessageEvent) => {
        handleMessage(pseudo, event);
    };

    socket.onclose = () => {
        if (intentionalClose) {
            connectionStatus.set('disconnected');
            loginError.set('');
            return;
        }
        if (!hasConnected) {
            loginError.set('Impossible de joindre le serveur de jeu. Vérifiez qu\'il est lancé puis réessayez.');
        }
        scheduleReconnect();
    };

    socket.onerror = () => {
        // onclose will fire after onerror, which triggers reconnection
    };
}

export function connect(pseudo: string, charName: string, charClass: string): void {
    intentionalClose = false;
    clearReconnectTimer();
    reconnectAttempts = 0;
    loginError.set('');
    savedPseudo = pseudo;
    savedCharName = charName;
    savedCharClass = charClass;
    persistSession();
    openSocket(pseudo, charName, charClass, false);
}

export function tryRestoreSession(): boolean {
    const session = restoreSession();
    if (!session) return false;
    intentionalClose = false;
    clearReconnectTimer();
    reconnectAttempts = 0;
    savedPseudo = session.pseudo;
    savedCharName = session.charName;
    savedCharClass = session.charClass;
    openSocket(session.pseudo, session.charName, session.charClass, false);
    return true;
}

export function disconnect(): void {
    intentionalClose = true;
    clearReconnectTimer();
    reconnectAttempts = 0;
    clearSession();
    if (socket) {
        socket.onopen = null;
        socket.onmessage = null;
        socket.onclose = null;
        socket.onerror = null;
        socket.close();
        socket = undefined;
    }
    connectionStatus.set('disconnected');
    gameState.update(s => ({ ...s, me: null }));
}

export function finalizeCharacter(stats: Record<string, unknown>): void {
    sendAction({ type: 'create_character', stats });
    gameState.update(s => ({ ...s, newChar: false }));
}

export function sendChat(message: string): void {
    sendAction({ type: 'chat_msg', message });
}

export function sendAction(action: Record<string, unknown>): void {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(action));
    }
}
