import { useEffect } from 'react';
import { useSetAtom } from 'jotai';
import { EventsOff, EventsOn } from '../../../wailsjs/runtime/runtime';
import { getAppLaunchSettings, getTraceSettings } from '@/wails/trace-backend';
import {
    accumulateTracesAtom,
    busyAtom,
    cleanupMessageAtom,
    elevatedAtom,
    enableOtsTraceAtom,
    errorAtom,
    failedFilesAtom,
    gatherDoneAtom,
    type GatherDoneEvent,
    gatheringAtom,
    isTracingAtom,
    progressAtom,
    requireElevationAtLaunchAtom,
    settingsAtom,
    verbosityAtom,
} from './a-trace-tool-atoms';

export function useSettingsInit() {
    const setSettings = useSetAtom(settingsAtom);
    const setIsTracing = useSetAtom(isTracingAtom);
    const setAccumulateTraces = useSetAtom(accumulateTracesAtom);
    const setEnableOtsTrace = useSetAtom(enableOtsTraceAtom);
    const setVerbosity = useSetAtom(verbosityAtom);
    const setElevated = useSetAtom(elevatedAtom);
    const setRequireElevationAtLaunch = useSetAtom(requireElevationAtLaunchAtom);
    const setError = useSetAtom(errorAtom);
    const setProgress = useSetAtom(progressAtom);
    const setGathering = useSetAtom(gatheringAtom);
    const setGatherDone = useSetAtom(gatherDoneAtom);
    const setFailedFiles = useSetAtom(failedFilesAtom);
    const setCleanupMessage = useSetAtom(cleanupMessageAtom);
    const setBusy = useSetAtom(busyAtom);

    useEffect(
        () => {
            async function loadSettings() {
                try {
                    const [s, launch] = await Promise.all([getTraceSettings(), getAppLaunchSettings()]);
                    setSettings(s);
                    setIsTracing(s.isTracing);
                    setAccumulateTraces(s.accumulateTraces);
                    setEnableOtsTrace(s.enableOtsTrace);
                    setVerbosity(s.verbosity || 4);
                    setElevated(launch.isElevated);
                    setRequireElevationAtLaunch(launch.requireElevationAtLaunch);
                } catch (e) {
                    setError(String(e));
                }
            }

            loadSettings();

            EventsOn('trace:state', (data: { isTracing: boolean; }) => {
                setIsTracing(data.isTracing);
            });
            EventsOn('trace:progress', (data: { collected: number; total: number; }) => {
                setProgress(data);
            });
            EventsOn('trace:gather-done', (data: GatherDoneEvent) => {
                setGathering(false);
                setGatherDone(true);
                setFailedFiles(data.failedFiles ?? []);
                setCleanupMessage(data.cleanupMessage ?? null);
                setBusy(false);
            });
            EventsOn('trace:gather-error', (data: { message: string; }) => {
                setGathering(false);
                setError(data.message);
                setBusy(false);
            });

            return () => {
                EventsOff('trace:state');
                EventsOff('trace:progress');
                EventsOff('trace:gather-done');
                EventsOff('trace:gather-error');
            };
        },
        []);
}
