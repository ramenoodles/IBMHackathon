# Logo Home Navigation Plan

## Overview

Convert the static "OnBober" logo text in the workspace header into a clickable button. Clicking it opens a confirmation dialog asking the user if they want to leave the workspace. Confirming navigates to the splash/home route (`/`). Cancelling dismisses the dialog and returns to the workspace.

---

## Sub-Task 1 — Make the logo a clickable button with a confirmation dialog

**Intent:** Wrap the logo `<span>` in a `<button>` element so the user can click it from the workspace. On click, a confirmation dialog appears before any navigation occurs, preventing accidental exits.

**Expected Outcomes:**
- The "OnBober" logo in the workspace header is visually a button (pointer cursor, subtle hover state)
- Clicking it opens a confirmation modal with the message: "Are you sure you would like to leave the workspace?"
- The modal has two buttons: "Yes, go home" (navigates to `/`) and "Return to workspace" (closes modal)
- Pressing Escape or clicking the backdrop also closes the modal (already handled by the existing `Modal` component)

**Todo:**
1. In the `<script setup>` block of `WorkspaceView.vue`, add a `ref` for dialog visibility: `const leaveConfirmOpen = ref(false)`
2. `useRouter` is already imported — instantiate `const router = useRouter()` if not already present
3. Replace the logo `<span>` (line 238) with a `<button>` that sets `leaveConfirmOpen.value = true` on click, with `cursor-pointer` styling and a subtle hover state matching the existing header style
4. Add a `<Modal>` instance (already imported at line 12) bound to `:open="leaveConfirmOpen"` with `title="Leave workspace?"` and `@close="leaveConfirmOpen = false"`
5. Inside the modal slot, add the message text and two buttons:
   - "Yes, go home" → calls `router.push('/')` 
   - "Return to workspace" → sets `leaveConfirmOpen.value = false`

**Relevant Context:**
- `frontend/src/views/WorkspaceView.vue` line 238 — logo `<span>` to convert
- `frontend/src/views/WorkspaceView.vue` line 12 — `Modal` already imported
- `frontend/src/views/WorkspaceView.vue` line 341 — existing `Modal` usage pattern to follow
- `frontend/src/router/index.ts` — `'/'` is the splash/home route
- `frontend/src/components/ui/Modal.vue` — accessible modal with Escape + backdrop close built in

**Status:** `[ ] pending`
