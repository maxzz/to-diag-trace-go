import { useEffect, useState } from 'react';
import { getAppLaunchSettings } from '@/wails/trace-backend';

export function ElevationStatusMarker() {
    const [elevated, setElevated] = useState<boolean | null>(null);

    useEffect(() => {
        getAppLaunchSettings()
            .then((settings) => setElevated(settings.isElevated))
            .catch(() => setElevated(false));
    }, []);

    if (elevated === null) {
        return null;
    }

    return (
        <span
            className="inline-flex items-center gap-1.5 rounded-full border border-white/30 bg-white/15 px-2.5 py-0.5 text-[11px] font-medium tracking-wide"
            title={
                elevated
                    ? 'Application is running with administrator privileges'
                    : 'Application is running without administrator privileges'
            }
        >
            <span
                className={`h-2 w-2 shrink-0 rounded-full ${elevated ? 'bg-emerald-300' : 'bg-amber-300'}`}
                aria-hidden="true"
            />
            {elevated ? 'Elevated' : 'Not elevated'}
        </span>
    );
}
