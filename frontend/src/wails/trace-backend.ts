import type { AppLaunchSettings, PendingAction, TraceSettings, TraceStartOptions } from './backend-types';
import { getBackendApp } from './is-wails';

function requireBackendApp() {
    const app = getBackendApp();
    if (!app) {
        throw new Error('Backend is not available.');
    }
    return app;
}

export function cancelGather(): void {
    requireBackendApp().CancelGather();
}

export function getAppLaunchSettings(): Promise<AppLaunchSettings> {
    return requireBackendApp().GetAppLaunchSettings();
}

export function getTraceSettings(): Promise<TraceSettings> {
    return requireBackendApp().GetTraceSettings();
}

export function isTracingActive(): Promise<boolean> {
    return requireBackendApp().IsTracingActive();
}

export function pickGatherDestination(): Promise<string> {
    return requireBackendApp().PickGatherDestination();
}

export function requestElevationForAction(action: PendingAction): Promise<void> {
    return requireBackendApp().RequestElevationForAction(action);
}

export function setRequireElevationAtLaunch(checked: boolean): Promise<void> {
    return requireBackendApp().SetRequireElevationAtLaunch(checked);
}

export function startTracing(opts: TraceStartOptions): Promise<void> {
    return requireBackendApp().StartTracing(opts);
}

export function stopTracingAndGather(dest: string): Promise<void> {
    return requireBackendApp().StopTracingAndGather(dest);
}

export function toggleDevTools(): Promise<void> {
    return requireBackendApp().ToggleDevTools();
}
