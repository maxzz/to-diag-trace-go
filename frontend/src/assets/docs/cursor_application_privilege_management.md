# Application privilege management
_Exported on 6/25/2026 at 19:18:48 PDT from Cursor (3.9.8)_

---

**User**

It is necessary to have a way to launch the application without elevated privileges for debugging the user interface, while in normal operation, it should start with elevated privileges. 
Ideally, this should be a saved setting that can be toggled via the UI—for instance, by adding a checkbox to the main screen to choose whether to launch the application with elevated privileges or in normal mode next time.	
Additionally, if an attempt is made to execute a function while the application is running without elevated privileges, a dialog box must be displayed asking whether to elevate the application's privileges. 
If the user chooses to elevate privileges, the process's privileges must be raised and the operation executed; otherwise, no action should be taken.

---

**Cursor**

Exploring the codebase to see how elevation and the UI are implemented today.


Reading the key files to implement this feature.


Implementing saved launch preference, conditional startup elevation, and an elevation confirmation dialog with deferred action execution after UAC relaunch.


Updating the frontend and Wails bindings.


Clearing the pending action if elevation relaunch fails.


Implemented optional elevation at launch, a persisted UI toggle, and on-demand elevation with deferred operation execution.

### Launch preference (saved setting)

- Added `requireElevationAtLaunch` to `%AppData%/to-diag-trace-go/init.json` (defaults to **true** for normal operation).
- **Startup** (`main.go`): UAC relaunch runs only when that setting is enabled and the app is not in `/Auto` mode.
- **Checkbox** on the main screen: *"Launch with administrator privileges next time"* — uncheck it to start in normal mode for UI debugging.

### On-demand elevation dialog

When a privileged action is attempted without admin rights (Start tracing / Stop and Gather):

1. A confirmation dialog appears instead of immediately relaunching.
2. **Cancel** — no action.
3. **Elevate and continue** — saves the pending operation, relaunches via UAC, and quits the current process.
4. After elevation, the elevated instance resumes the operation automatically on startup (trace start, or save-file dialog + gather).

### Files changed

| Area | Files |
|------|-------|
| Settings | `backend/options.go` — `RequireElevationAtLaunch`, `PendingAction`, merge helpers |
| Elevation API | `backend/elevation.go` — `GetAppLaunchSettings`, `SetRequireElevationAtLaunch`, `RequestElevationForAction` |
| Startup | `main.go` — conditional elevation gate |
| Lifecycle | `backend/app.go` — `resumePendingAction()` on `DomReady` |
| Guards | `backend/trace.go` — admin check on `StopTracingAndGather` |
| UI | `frontend/src/components/TraceTool.tsx` — checkbox + dialog |
| Bindings | `frontend/wailsjs/go/backend/App.{js,d.ts}`, `models.ts` |

### Usage

- **UI debugging**: uncheck the checkbox, restart — the app opens without UAC.
- **Normal use**: leave the checkbox checked (default) — UAC prompt on launch as before.
- **Privileged ops while in debug mode**: use Start tracing or Gather — confirm elevation in the dialog when prompted.

Note: Windows cannot elevate an existing process in place; elevation always relaunches the app via UAC, then continues the saved operation in the new elevated instance.

---

**User**

Add marker to the UI Frontend page to show application Elevation current status

---

**Cursor**

Adding a header status marker that fetches elevation state so it's always visible on the page.


Added an elevation status marker in the page header.

### What was added

- New component: `frontend/src/components/ElevationStatusMarker.tsx`
- Placed in the header of `App.tsx`, below the title

### Appearance

- **Elevated** — green dot + “Elevated”
- **Not elevated** — amber dot + “Not elevated”

The badge uses a pill style on the blue header bar and includes a tooltip with the full explanation on hover. It loads status from `GetAppLaunchSettings()` on mount.

