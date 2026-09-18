import { writable } from 'svelte/store';

export type ToastKind = 'success' | 'error' | 'info' | 'warn' | 'damage' | 'heal' | 'crit';

export interface Toast {
    id: number;
    kind: ToastKind;
    title: string;
    msg: string;
}

export const toasts = writable<Toast[]>([]);

let toastId = 0;

export function pushToast(kind: ToastKind, title: string, msg = '', duration = 3800): void {
    const id = ++toastId;
    toasts.update((list) => [...list.slice(-4), { id, kind, title, msg }]);
    setTimeout(() => dismissToast(id), duration);
}

export function dismissToast(id: number): void {
    toasts.update((list) => list.filter((t) => t.id !== id));
}