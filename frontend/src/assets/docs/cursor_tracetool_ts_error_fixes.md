# TraceTool error fixes
_Exported on 6/25/2026 at 20:24:55 PDT from Cursor (3.9.8)_

---

**User**

@frontend/src/components/TraceTool.tsx:158-190 fix these errors

---

**Cursor**

Inspecting `TraceTool.tsx` and linter diagnostics for the elevation confirmation block.


Checking the generated Wails types for `RequestElevationForAction` and `PendingAction`.


Wails expects `backend.PendingAction` class instances, not plain objects — updating the calls to use the generated model.


Fixed. Wails generates `backend.PendingAction` as a class (with a required `convertValues` helper for nested types), so plain object literals were not assignable to that parameter type.

The fix:
1. Import `backend` from `../../wailsjs/go/models`
2. Wrap both `RequestElevationForAction` calls with `new backend.PendingAction({ ... })`

That matches the Wails binding pattern and clears both TypeScript errors.

---

**User**

@frontend/wailsjs/go/models.ts How to fix these errors? I did it in this project C:\y\w\2-web\0-dp\utils\trace-viewer-26-go but it is not enough here. I may missed something. 

---

**Cursor**

Comparing TypeScript config and Wails bindings in this project with the fix in trace-viewer-26-go.


The `exclude` in tsconfig doesn't help when `src` still imports `App.d.ts`, which pulls in `models.ts`. Checking all backend imports to apply the trace-viewer wrapper pattern.


The missing piece: `exclude: ["wailsjs"]` doesn't apply when `src` imports `wailsjs/go/backend/App`, which pulls in `models.ts`. Applying the trace-viewer pattern — route backend calls through `src/wails` wrappers.


The `exclude: ["wailsjs"]` in `tsconfig.app.json` was only half of what trace-viewer needed. Here’s why it still failed here.

## What was missing

`exclude` only affects files TypeScript would include on its own. It does **not** stop following imports from `src`.

In this project, `src` still imported generated bindings:

```
src/components/*.tsx
  → wailsjs/go/backend/App.d.ts
    → wailsjs/go/models.ts   ← `export namespace` + erasableSyntaxOnly = TS1294
```

In trace-viewer-26-go, the real fix was **not importing `wailsjs/go/backend/App` from `src` at all**. Only `wailsjs/runtime/runtime` is imported there, and runtime bindings don’t pull in `models.ts`.

## What I changed

Added a thin wrapper layer under `src/wails/` (same pattern as trace-viewer):

- `backend-types.ts` — hand-written types (plain objects, no namespaces)
- `is-wails.ts` — `getBackendApp()` via `window.go.backend.App`
- `trace-backend.ts` — typed wrappers for all backend calls

Updated `App.tsx`, `ElevationStatusMarker.tsx`, and `TraceTool.tsx` to use those wrappers instead of `wailsjs/go/backend/App` and `wailsjs/go/models`.

`tsc --project tsconfig.app.json` now passes with `erasableSyntaxOnly: true`, and you never need to edit generated `models.ts` — even when Wails regenerates it with `export namespace`.

## Rule of thumb

| Import from | Safe with `erasableSyntaxOnly`? |
|---|---|
| `wailsjs/runtime/runtime` | Yes |
| `wailsjs/go/backend/App` or `wailsjs/go/models` | No — pulls in namespace code |

Keep `exclude: ["wailsjs"]` as documentation, but the important part is: **route Go backend calls through `src/wails/`, not direct generated imports.**

