import { useEffect } from 'react';
import { ToggleDevTools } from '../../wailsjs/go/backend/App';
// import wailsLogo from './assets/wails.png';

export function App() {

    useEffect(
        () => {
            function handleKeyDown(e: KeyboardEvent) {
                const isDevToolsShortcut = (e.ctrlKey && e.shiftKey && e.code === 'F12') || (e.ctrlKey && e.shiftKey && e.code === 'KeyI');
                if (isDevToolsShortcut) {
                    ToggleDevTools().catch(console.error);
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
                Go wrapped frontend
            </header>

            <main className="self-center justify-self-center p-4">
                <div className="text-blue-900 font-bold">
                    Go wrapped frontend
                </div>
            </main>

            <footer className="p-3 text-center text-white bg-linear-to-r from-blue-500 to-blue-700 border-t border-blue-900">
                <p>&copy; 2026 No rights reserved.</p>
                {/* <div className="w-fit max-w-md">
                    <a href="https://wails.io" target="_blank">
                        <img src={wailsLogo} className="logo wails" alt="Wails logo" />
                    </a>
                </div> */}
            </footer>

        </div>
    );
}
