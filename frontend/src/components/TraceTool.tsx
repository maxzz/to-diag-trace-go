import { useCallback, useEffect, useState } from 'react';
import { EventsOff, EventsOn } from '../../wailsjs/runtime/runtime';
import {
    CancelGather,
    GetAppLaunchSettings,
    GetTraceSettings,
    IsTracingActive,
    PickGatherDestination,
    RequestElevationForAction,
    SetRequireElevationAtLaunch,
    StartTracing,
    StopTracingAndGather,
} from '../../wailsjs/go/backend/App';
import { backend } from '../../wailsjs/go/models';
import { Button } from '@/ui/shadcn/button';

type TraceSettings = {
    isTracing: boolean;
    accumulateTraces: boolean;
    enableOtsTrace: boolean;
    verbosity: number;
    tracePath: string;
    passwordManagerFound: boolean;
};

type GatherDoneEvent = {
    zipPath: string;
    failedFiles: string[];
    deleteOnReboot?: boolean;
    cleanupMessage?: string;
};

type PendingElevationAction = 'startTracing' | 'gather' | null;

export function TraceTool() {
    const [settings, setSettings] = useState<TraceSettings | null>(null);
    const [isTracing, setIsTracing] = useState(false);
    const [showMore, setShowMore] = useState(false);
    const [accumulateTraces, setAccumulateTraces] = useState(true);
    const [enableOtsTrace, setEnableOtsTrace] = useState(false);
    const [verbosity, setVerbosity] = useState(4);
    const [elevated, setElevated] = useState(true);
    const [requireElevationAtLaunch, setRequireElevationAtLaunch] = useState(true);
    const [elevationDialogOpen, setElevationDialogOpen] = useState(false);
    const [pendingElevationAction, setPendingElevationAction] = useState<PendingElevationAction>(null);
    const [gathering, setGathering] = useState(false);
    const [progress, setProgress] = useState({ collected: 0, total: 0 });
    const [failedFiles, setFailedFiles] = useState<string[]>([]);
    const [gatherDone, setGatherDone] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [cleanupMessage, setCleanupMessage] = useState<string | null>(null);
    const [busy, setBusy] = useState(false);

    const loadSettings = useCallback(async () => {
        try {
            const [s, launch] = await Promise.all([GetTraceSettings(), GetAppLaunchSettings()]);
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
    }, []);

    useEffect(() => {
        loadSettings();

        EventsOn('trace:state', (data: { isTracing: boolean }) => {
            setIsTracing(data.isTracing);
        });
        EventsOn('trace:progress', (data: { collected: number; total: number }) => {
            setProgress(data);
        });
        EventsOn('trace:gather-done', (data: GatherDoneEvent) => {
            setGathering(false);
            setGatherDone(true);
            setFailedFiles(data.failedFiles ?? []);
            setCleanupMessage(data.cleanupMessage ?? null);
            setBusy(false);
        });
        EventsOn('trace:gather-error', (data: { message: string }) => {
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
    }, [loadSettings]);

    async function handleRequireElevationChange(checked: boolean) {
        try {
            await SetRequireElevationAtLaunch(checked);
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
                const dest = await PickGatherDestination();
                if (!dest) {
                    setGathering(false);
                    setBusy(false);
                    const active = await IsTracingActive();
                    setIsTracing(active);
                    return;
                }
                await StopTracingAndGather(dest);
            } catch (e) {
                setError(String(e));
                setGathering(false);
                setBusy(false);
            }
            return;
        }

        setBusy(true);
        try {
            await StartTracing({
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

    async function confirmElevation() {
        const action = pendingElevationAction;
        setElevationDialogOpen(false);
        setPendingElevationAction(null);
        if (!action) {
            return;
        }

        try {
            if (action === 'startTracing') {
                await RequestElevationForAction(
                    new backend.PendingAction({
                        type: 'startTracing',
                        startOpts: {
                            accumulateTraces,
                            enableOtsTrace: enableOtsTrace && (settings?.passwordManagerFound ?? false),
                            verbosity,
                        },
                    }),
                );
            } else {
                setBusy(true);
                setGathering(true);
                setGatherDone(false);
                setFailedFiles([]);
                setProgress({ collected: 0, total: 0 });
                await RequestElevationForAction(new backend.PendingAction({ type: 'gather' }));
            }
        } catch (e) {
            setError(String(e));
            setGathering(false);
            setBusy(false);
        }
    }

    function cancelElevation() {
        setElevationDialogOpen(false);
        setPendingElevationAction(null);
    }

    function handleCancelGather() {
        CancelGather();
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

            {elevationDialogOpen && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
                    <div
                        role="dialog"
                        aria-modal="true"
                        aria-labelledby="elevation-dialog-title"
                        className="w-full max-w-sm space-y-4 rounded-md border border-blue-200 bg-white p-4 shadow-lg"
                    >
                        <h2 id="elevation-dialog-title" className="text-sm font-semibold text-blue-950">
                            Administrator privileges required
                        </h2>
                        <p className="text-xs leading-relaxed text-blue-900">
                            This operation requires administrator privileges. Relaunch the application elevated
                            to continue?
                        </p>
                        <div className="flex justify-end gap-2">
                            <Button variant="outline" size="sm" onClick={cancelElevation}>
                                Cancel
                            </Button>
                            <Button size="sm" onClick={confirmElevation}>
                                Elevate and continue
                            </Button>
                        </div>
                    </div>
                </div>
            )}

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
