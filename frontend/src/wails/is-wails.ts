import type {
    AppLaunchSettings,
    PendingAction,
    TraceSettings,
    TraceStartOptions,
} from './backend-types';

type BackendApp = {
    CancelGather: () => Promise<void>;
    GetAppLaunchSettings: () => Promise<AppLaunchSettings>;
    GetTraceSettings: () => Promise<TraceSettings>;
    IsTracingActive: () => Promise<boolean>;
    PickGatherDestination: () => Promise<string>;
    RequestElevationForAction: (action: PendingAction) => Promise<void>;
    SetRequireElevationAtLaunch: (checked: boolean) => Promise<void>;
    StartTracing: (opts: TraceStartOptions) => Promise<void>;
    StopTracingAndGather: (dest: string) => Promise<void>;
    ToggleDevTools: () => Promise<void>;
};

type WailsWindow = Window & {
    go?: {
        backend?: {
            App?: BackendApp;
        };
    };
    runtime?: unknown;
};

export function isBackendAvailable(): boolean {
    if (typeof window === 'undefined') {
        return false;
    }

    return Boolean((window as WailsWindow).go?.backend?.App);
}

export function getBackendApp(): BackendApp | undefined {
    if (!isBackendAvailable()) {
        return undefined;
    }

    return (window as WailsWindow).go!.backend!.App!;
}
