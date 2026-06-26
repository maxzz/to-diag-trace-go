import { useEffect } from 'react';
import { toggleDevTools } from '@/wails/trace-backend';
import { ElevationStatusMarker } from './ElevationStatusMarker';
import { TraceTool } from './TraceTool';

export function App() {

    useEffect(
        () => {
            function handleKeyDown(e: KeyboardEvent) {
                const isDevToolsShortcut = (e.ctrlKey && e.shiftKey && e.code === 'F12') || (e.ctrlKey && e.shiftKey && e.code === 'KeyI');
                if (isDevToolsShortcut) {
                    toggleDevTools().catch(console.error);
                }
            }
            
            const controller = new AbortController();
            window.addEventListener('keydown', handleKeyDown, { signal: controller.signal });
            return () => controller.abort();
        }, []
    );

    return (
        <div className="min-h-screen text-sm bg-white grid grid-rows-[auto_1fr_auto]">

            <header className="p-3 text-center text-white bg-linear-to-r from-blue-500 to-blue-700 border-b border-blue-900 shadow">
                <div className="flex flex-col items-center gap-2">
                    <span>DigitalPersona Diagnostic Tool</span>
                    <ElevationStatusMarker />
                </div>
            </header>

            <main className="self-center justify-self-center p-4 w-full flex justify-center">
                <TraceTool />
            </main>

            <footer className="p-3 text-center text-white bg-linear-to-r from-blue-500 to-blue-700 border-t border-blue-900">
                <p>&copy; 2026 HID Global</p>
            </footer>

        </div>
    );
}
