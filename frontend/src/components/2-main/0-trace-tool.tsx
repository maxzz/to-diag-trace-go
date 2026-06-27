import { useCallback, useEffect } from 'react';
import { useAtom, useSetAtom } from 'jotai';
import { EventsOff, EventsOn } from '../../../wailsjs/runtime/runtime';
import {
    cancelGather,
    getAppLaunchSettings,
    getTraceSettings,
    isTracingActive,
    pickGatherDestination,
    setRequireElevationAtLaunch as persistRequireElevationAtLaunch,
    startTracing,
    stopTracingAndGather,
} from '@/wails/trace-backend';
import { Button } from '@/ui/shadcn/button';
import {
    accumulateTracesAtom,
    busyAtom,
    cleanupMessageAtom,
    elevatedAtom,
    elevationDialogOpenAtom,
    enableOtsTraceAtom,
    errorAtom,
    failedFilesAtom,
    gatherDoneAtom,
    type GatherDoneEvent,
    gatheringAtom,
    isTracingAtom,
    pendingElevationActionAtom,
    progressAtom,
    requireElevationAtLaunchAtom,
    settingsAtom,
    showMoreAtom,
    verbosityAtom,
} from './a-trace-tool-atoms';
import { ElevationDialog } from './b-elevation-dialog';

export function TraceTool() {
    const [settings, setSettings] = useAtom(settingsAtom);
    const [isTracing, setIsTracing] = useAtom(isTracingAtom);
    const [showMore, setShowMore] = useAtom(showMoreAtom);
    const [accumulateTraces, setAccumulateTraces] = useAtom(accumulateTracesAtom);
    const [enableOtsTrace, setEnableOtsTrace] = useAtom(enableOtsTraceAtom);
    const [verbosity, setVerbosity] = useAtom(verbosityAtom);
    const [elevated, setElevated] = useAtom(elevatedAtom);
    const [requireElevationAtLaunch, setRequireElevationAtLaunch] = useAtom(requireElevationAtLaunchAtom);
    const setElevationDialogOpen = useSetAtom(elevationDialogOpenAtom);
    const setPendingElevationAction = useSetAtom(pendingElevationActionAtom);
    const [gathering, setGathering] = useAtom(gatheringAtom);
    const [progress, setProgress] = useAtom(progressAtom);
    const [failedFiles, setFailedFiles] = useAtom(failedFilesAtom);
    const [gatherDone, setGatherDone] = useAtom(gatherDoneAtom);
    const [error, setError] = useAtom(errorAtom);
    const [cleanupMessage, setCleanupMessage] = useAtom(cleanupMessageAtom);
    const [busy, setBusy] = useAtom(busyAtom);

    const loadSettings = useCallback(
        async () => {
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
        },
        [
            setSettings,
            setIsTracing,
            setAccumulateTraces,
            setEnableOtsTrace,
            setVerbosity,
            setElevated,
            setRequireElevationAtLaunch,
            setError,
        ],
    );

    useEffect(
        () => {
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
        [
            loadSettings,
            setIsTracing,
            setProgress,
            setGathering,
            setGatherDone,
            setFailedFiles,
            setCleanupMessage,
            setBusy,
            setError,
        ],
    );

    async function handleRequireElevationChange(checked: boolean) {
        try {
            await persistRequireElevationAtLaunch(checked);
            setRequireElevationAtLaunch(checked);
        } catch (e) {
            setError(String(e));
        }
    }

    async function runPrivilegedAction() {
        if (isTracing) {
            setBusy(true);
            setGathering(true);
            setGatherDone(false);
            setFailedFiles([]);
            setProgress({ collected: 0, total: 0 });
            try {
                const dest = await pickGatherDestination();
                if (!dest) {
                    setGathering(false);
                    setBusy(false);
                    const active = await isTracingActive();
                    setIsTracing(active);
                    return;
                }
                await stopTracingAndGather(dest);
            } catch (e) {
                setError(String(e));
                setGathering(false);
                setBusy(false);
            }
            return;
        }

        setBusy(true);
        try {
            await startTracing({
                accumulateTraces,
                enableOtsTrace: enableOtsTrace && (settings?.passwordManagerFound ?? false),
                verbosity,
            });
            setIsTracing(true);
        } catch (e) {
            setError(String(e));
        } finally {
            setBusy(false);
        }
    }

    async function handlePrimaryAction() {
        setError(null);
        if (!elevated) {
            setPendingElevationAction(isTracing ? 'gather' : 'startTracing');
            setElevationDialogOpen(true);
            return;
        }

        await runPrivilegedAction();
    }

    function handleCancelGather() {
        cancelGather();
        setGathering(false);
        setBusy(false);
    }

    return (
        <div className="w-full max-w-md space-y-3 text-blue-950">
            <p className="text-xs leading-relaxed">
                The DigitalPersona Diagnostic Tool collects information while DigitalPersona software is running
                and saves it in a zip file.
            </p>
            <p className="text-xs leading-relaxed text-blue-800/80">
                This information is used for diagnostic purposes only. It does not contain passwords or
                information in fields detected as protected.
            </p>
            <p className="text-xs leading-relaxed">
                Start tracing, reproduce the problem, gather trace files and send the resulting archive to
                technical support.
            </p>

            <label className="flex items-center gap-2 text-xs">
                <input
                    type="checkbox"
                    checked={requireElevationAtLaunch}
                    disabled={busy}
                    onChange={(e) => handleRequireElevationChange(e.target.checked)}
                />
                Launch with administrator privileges next time
            </label>

            {!elevated && (
                <p className="text-xs font-medium text-amber-700">
                    Running without administrator privileges. Trace operations will prompt for elevation.
                </p>
            )}

            <ElevationDialog />

            {showMore && (
                <div className="space-y-2 rounded-md border border-blue-200 bg-blue-50/50 p-3 text-xs">
                    <label className="flex items-center gap-2">
                        <input
                            type="checkbox"
                            checked={accumulateTraces}
                            disabled={isTracing || busy}
                            onChange={(e) => setAccumulateTraces(e.target.checked)}
                        />
                        Accumulate traces between multiple runs of applications
                    </label>
                    <label className="flex items-center gap-2">
                        <input
                            type="checkbox"
                            checked={enableOtsTrace}
                            disabled={!settings?.passwordManagerFound || isTracing || busy}
                            onChange={(e) => setEnableOtsTrace(e.target.checked)}
                        />
                        Include Password Manager traces
                    </label>
                    <div className="flex items-center gap-2">
                        <span>Verbosity:</span>
                        <input
                            type="number"
                            min={1}
                            max={9}
                            className="w-14 rounded border border-blue-200 px-2 py-0.5"
                            value={verbosity}
                            disabled={isTracing || busy}
                            onChange={(e) => setVerbosity(Number(e.target.value))}
                        />
                    </div>
                    <div>
                        <div className="mb-1">Folder where collected trace files are placed:</div>
                        <input
                            readOnly
                            className="w-full rounded border border-blue-200 bg-white px-2 py-1 text-[11px]"
                            value={settings?.tracePath ?? ''}
                        />
                    </div>
                </div>
            )}

            <div className="flex justify-end">
                <Button
                    variant="ghost"
                    size="sm"
                    className="text-blue-700"
                    onClick={() => setShowMore((v) => !v)}
                    disabled={busy && !showMore}
                >
                    {showMore ? 'Less' : 'More'}
                </Button>
            </div>

            {gathering && (
                <div className="space-y-1">
                    <div className="text-xs">Collecting files…</div>
                    <div className="h-2 overflow-hidden rounded bg-blue-100">
                        <div
                            className="h-full bg-blue-600 transition-all"
                            style={{
                                width: progress.total
                                    ? `${Math.round((progress.collected / progress.total) * 100)}%`
                                    : '0%',
                            }}
                        />
                    </div>
                    <Button variant="outline" size="sm" onClick={handleCancelGather}>
                        Cancel
                    </Button>
                </div>
            )}

            {failedFiles.length > 0 && (
                <div className="space-y-1">
                    <div className="text-xs font-medium">Failed to collect the following files:</div>
                    <textarea
                        readOnly
                        className="h-20 w-full rounded border border-red-200 bg-red-50/30 p-2 text-[11px]"
                        value={failedFiles.join('\n')}
                    />
                </div>
            )}

            {cleanupMessage && (
                <p className="text-xs text-blue-800">{cleanupMessage}</p>
            )}

            {error && <p className="text-xs text-red-700">{error}</p>}

            {isTracing && !gathering && (
                <p className="text-xs font-medium text-green-700">Tracing is active.</p>
            )}

            {gatherDone && !gathering && (
                <div className="flex gap-2">
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                            setGatherDone(false);
                            setIsTracing(false);
                        }}
                    >
                        Close
                    </Button>
                    <Button
                        size="sm"
                        onClick={() => {
                            setGatherDone(false);
                            setCleanupMessage(null);
                            setError(null);
                        }}
                    >
                        Start Again
                    </Button>
                </div>
            )}

            {!gatherDone && (
                <Button
                    className="h-10 w-full text-base"
                    disabled={busy && !isTracing}
                    onClick={handlePrimaryAction}
                >
                    {isTracing ? 'Stop tracing and Gather trace files' : 'Start tracing'}
                </Button>
            )}
        </div>
    );
}
