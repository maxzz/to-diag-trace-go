import { useAtom, useAtomValue, useSetAtom } from 'jotai';
import { requestElevationForAction } from '@/wails/trace-backend';
import { Button } from '@/ui/shadcn/button';
import {
    accumulateTracesAtom,
    busyAtom,
    elevationDialogOpenAtom,
    enableOtsTraceAtom,
    errorAtom,
    failedFilesAtom,
    gatherDoneAtom,
    gatheringAtom,
    pendingElevationActionAtom,
    progressAtom,
    settingsAtom,
    verbosityAtom,
} from './a-trace-tool-atoms';

export function ElevationDialog() {
    const open = useAtomValue(elevationDialogOpenAtom);
    const [pendingAction, setPendingAction] = useAtom(pendingElevationActionAtom);
    const setOpen = useSetAtom(elevationDialogOpenAtom);
    const accumulateTraces = useAtomValue(accumulateTracesAtom);
    const enableOtsTrace = useAtomValue(enableOtsTraceAtom);
    const settings = useAtomValue(settingsAtom);
    const verbosity = useAtomValue(verbosityAtom);
    const setError = useSetAtom(errorAtom);
    const setBusy = useSetAtom(busyAtom);
    const setGathering = useSetAtom(gatheringAtom);
    const setGatherDone = useSetAtom(gatherDoneAtom);
    const setFailedFiles = useSetAtom(failedFilesAtom);
    const setProgress = useSetAtom(progressAtom);

    if (!open) {
        return null;
    }

    function cancelElevation() {
        setOpen(false);
        setPendingAction(null);
    }

    async function confirmElevation() {
        const action = pendingAction;
        setOpen(false);
        setPendingAction(null);
        if (!action) {
            return;
        }

        try {
            if (action === 'startTracing') {
                await requestElevationForAction({
                    type: 'startTracing',
                    startOpts: {
                        accumulateTraces,
                        enableOtsTrace: enableOtsTrace && (settings?.passwordManagerFound ?? false),
                        verbosity,
                    },
                });
            } else {
                setBusy(true);
                setGathering(true);
                setGatherDone(false);
                setFailedFiles([]);
                setProgress({ collected: 0, total: 0 });
                await requestElevationForAction({ type: 'gather' });
            }
        } catch (e) {
            setError(String(e));
            setGathering(false);
            setBusy(false);
        }
    }

    return (
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
    );
}
