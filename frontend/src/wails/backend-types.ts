export type AppLaunchSettings = {
    requireElevationAtLaunch: boolean;
    isElevated: boolean;
};

export type TraceStartOptions = {
    accumulateTraces: boolean;
    enableOtsTrace: boolean;
    verbosity: number;
};

export type PendingAction = {
    type: string;
    startOpts?: TraceStartOptions;
};

export type TraceSettings = {
    isTracing: boolean;
    accumulateTraces: boolean;
    enableOtsTrace: boolean;
    verbosity: number;
    tracePath: string;
    passwordManagerFound: boolean;
};
