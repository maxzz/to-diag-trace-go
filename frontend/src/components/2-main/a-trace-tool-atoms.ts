import { atom } from 'jotai';
import { setRequireElevationAtLaunch as persistRequireElevationAtLaunch } from '@/wails/trace-backend';

export type TraceSettings = {
    isTracing: boolean;
    accumulateTraces: boolean;
    enableOtsTrace: boolean;
    verbosity: number;
    tracePath: string;
    passwordManagerFound: boolean;
};

export type GatherDoneEvent = {
    zipPath: string;
    failedFiles: string[];
    deleteOnReboot?: boolean;
    cleanupMessage?: string;
};

export type PendingElevationAction = 'startTracing' | 'gather' | null;

export type GatherProgress = {
    collected: number;
    total: number;
};

export const settingsAtom = atom<TraceSettings | null>(null);

export const isTracingAtom = atom(false);

export const showMoreAtom = atom(false);

export const accumulateTracesAtom = atom(true);
export const enableOtsTraceAtom = atom(false);

export const verbosityAtom = atom(4);

export const elevatedAtom = atom(true);
export const requireElevationAtLaunchBaseAtom = atom(true);
export const elevationDialogOpenAtom = atom(false);
export const pendingElevationActionAtom = atom<PendingElevationAction>(null);

export const gatheringAtom = atom(false);
export const progressAtom = atom<GatherProgress>({ collected: 0, total: 0 });
export const failedFilesAtom = atom<string[]>([]);
export const gatherDoneAtom = atom(false);

export const errorAtom = atom<string | null>(null);
export const cleanupMessageAtom = atom<string | null>(null);

export const requireElevationAtLaunchAtom = atom(
    (get) => get(requireElevationAtLaunchBaseAtom),
    async (_get, set, checked: boolean) => {
        try {
            await persistRequireElevationAtLaunch(checked);
            set(requireElevationAtLaunchBaseAtom, checked);
        } catch (e) {
            set(errorAtom, String(e));
        }
    },
);

export const busyAtom = atom(false);
