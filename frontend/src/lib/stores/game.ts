import { writable, derived } from 'svelte/store';

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
    equipement: Equipment;
    quests: Quest[];
}

export interface Location {
    nom: string;
    background: string;
    objects: Item[];
    quests?: Quest[];
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
    locations?: { nom: string; objects: Item[]; quests?: Quest[] }[];
}

interface ChatData {
    type: 'chat';
    msg: string;
}

export function abilityModifier(stat: number): number {
    return Math.floor(((stat || 10) - 10) / 2);
}

export function normalizeName(name: string): string {
    return (name || '')
        .toLowerCase()
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '');
}

export interface WeaponInfo {
    sides: number;
    ranged: boolean;
    label: string;
}

function parseDiceSides(s: string): number {
    let t = (s || '').trim().toLowerCase();
    const i = t.indexOf('d');
    if (i >= 0) t = t.slice(i + 1);
    const n = parseInt(t, 10);
    return Number.isFinite(n) && n >= 2 ? n : 0;
}

export function getWeaponInfo(item: Item | null): WeaponInfo {
    if (!item) return { sides: 2, ranged: false, label: '1d2' };
    const n = normalizeName(item.nom);
    let sides: number;
    if (n.includes('dague')) sides = 4;
    else if (n.includes('arbalete') || n.includes('carquois')) sides = 8;
    else if (n.includes('arc')) sides = 8;
    else if (n.includes('epee')) sides = 6;
    else sides = 6;
    const ranged = n.includes('arc') || n.includes('arbalete') || n.includes('carquois');
    const custom = parseDiceSides(item.desDegats);
    if (custom) sides = custom;
    return { sides, ranged, label: `1d${sides}` };
}

function signed(v: number): string {
    return v > 0 ? `+${v}` : `${v}`;
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

function hitDiceSides(classe: string): number {
    switch (classe) {
        case 'Magicien': return 6;
        case 'Voleur':
        case 'Clerc':
        case 'Barde':
        case 'Ranger': return 8;
        default: return 10; // Guerrier
    }
}

export function getDerivedStats(s: PlayerStats | undefined): DerivedStats {
    if (!s || !s.stats) {
        return { maxPv: 10, ac: 10, attackMod: 0, damageMod: 0, damageDice: '1d2', weaponName: 'Mains nues', ranged: false };
    }
    const con = s.stats.constitution || 10;
    const conMod = abilityModifier(con);
    const hd = hitDiceSides(s.classe || 'Guerrier');
    const maxPv = Math.max(1, hd + conMod);
    const weapon = s.equipement?.arme || null;
    const armor = s.equipement?.armure || null;
    const weaponInfo = getWeaponInfo(weapon);
    const attackStat = weaponInfo.ranged ? (s.stats.vitesse || 10) : (s.stats.force || 10);
    const attackMod = abilityModifier(attackStat) + (weapon?.bonusDegats || 0);
    const damageMod = abilityModifier(attackStat);
    const ac = 10 + abilityModifier(s.stats.vitesse || 10) + (armor?.bonusArmure || 0);
    return {
        maxPv,
        ac,
        attackMod,
        damageMod,
        damageDice: `${weaponInfo.label}${signed(damageMod)}`,
        weaponName: weapon?.nom || 'Mains nues',
        ranged: weaponInfo.ranged
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
    return getDerivedStats($ms);
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
    const data: SyncData | ChatData | any = JSON.parse(event.data);
    if (data.type === 'chat') {
        gameState.update(s => ({ ...s, logs: [...s.logs, data.msg] }));
    } else if (data.type === 'init_new_char') {
        gameState.update(s => ({ ...s, newChar: true }));
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

    socket = new WebSocket(wsUrl(pseudo));

    socket.onopen = () => {
        reconnectAttempts = 0;
        connectionStatus.set('connected');
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
            return;
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

export function sendAction(action: Record<string, unknown>): void {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(action));
    }
}
