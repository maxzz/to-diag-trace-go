# DigitalPersona Diagnostic Tool (Go)

Go + Wails rewrite of the legacy C++ **DpDiagnosticTool** — a Windows utility for collecting DigitalPersona diagnostic traces.

## Stack

- **Go / Wails v2** — desktop shell, native dialogs, registry/ETW integration
- **React + Vite + Tailwind** — UI
- **`diag/` package** — portable diagnostic logic (Windows-only)
- **`backend/` package** — Wails bindings and window persistence

## Requirements

- Windows 10/11 x64
- Go 1.22+
- [Wails CLI v2.12+](https://wails.io/docs/gettingstarted/installation)
- pnpm (frontend)

## Development

```powershell
pnpm install
pnpm --prefix frontend install
pnpm run dev
```

Frontend dev server: http://localhost:3000 (via `wails dev`).

## Build

```powershell
pnpm run build
# output: build\bin\DpDiagnosticTool.exe
```

## Usage

1. Launch the app (elevates automatically if needed).
2. Optionally expand **More** to configure trace options.
3. Click **Start tracing**, reproduce the issue, then **Stop tracing and Gather trace files**.
4. Choose a ZIP destination; the archive opens in Explorer when complete.

### Command-line modes

| Switch | Behavior |
|--------|----------|
| `/Auto` | Started from Run key; shows UI only if tracing is still active |
| `/DeleteFiles "path64" "path32"` | Headless boot cleanup of trace folders |

## Project layout

```
main.go              # entry, embed, CLI routing
backend/             # Wails app bindings
diag/                # trace registry, ETW, PowerShell, ZIP gather, scheduler
frontend/            # React UI (TraceTool in App shell)
build/               # icons, installer assets
```

## Notes

- Requires administrator privileges to write tracing registry keys and Run key entries.
- PowerShell diagnostics script is embedded verbatim from the original C++ `BINARY` resource (`diag/res/PowerShell.ps1`). At runtime it is written to a temp `.ps1` file and invoked via `powershell.exe -File <script> start|stop "<tracePath>" <verbosity>` with the same flags as `PsTraceUtils::Execute`.
- Single-instance: a second launch activates the existing window.
