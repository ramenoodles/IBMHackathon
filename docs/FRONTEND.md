# OnBober — Frontend Reference

The frontend is a Vue 3 + TypeScript single-page application built with Vite. It presents a fully interactive codebase explorer: users load a repository, browse files, step through control-flow graphs, and read AI-generated explanations tailored to their experience level.

## Table of Contents

1. [Tech Stack](#tech-stack)
2. [Source Layout](#source-layout)
3. [User Journey](#user-journey)
4. [State Management](#state-management)
5. [Routing](#routing)
6. [API Client](#api-client)
7. [Views](#views)
8. [Workspace Layout & Panels](#workspace-layout--panels)
9. [File Exploration](#file-exploration)
10. [Flow Graph System](#flow-graph-system)
11. [AI Explanations](#ai-explanations)
12. [Workspace Components](#workspace-components)
13. [UI Components](#ui-components)
14. [Composables Reference](#composables-reference)
15. [Utils & Types](#utils--types)
16. [Constants](#constants)
17. [Personalisation](#personalisation)
18. [Bobbers (Easter Egg)](#bobbers-easter-egg)
19. [Styling](#styling)
20. [Build & Development](#build--development)

---

## Tech Stack

| Library | Version | Purpose |
|---------|---------|---------|
| Vue 3 | 3.x | Composition API, SFCs |
| TypeScript | 5.x | Type safety |
| Vite | 8.x | Dev server, bundler |
| Vue Router | 4.x | Client-side routing |
| Tailwind CSS | 4.x | Utility-first styling (Vite plugin) |
| Mermaid | 11.x | Flowchart DSL → SVG |
| Panzoom | 4.x | Pan & zoom on graph canvas |
| Monaco Editor | CDN | Syntax-aware code viewer |
| Shiki | 3.x | Syntax highlighting |
| Marked | 18.x | Markdown → HTML |
| Vitest | 3.x | Unit test runner |
| ESLint + Oxlint | — | Linting |
| Prettier | — | Formatting |

---

## Source Layout

```
frontend/src/
├── api/
│   └── index.ts              HTTP client wrapping all backend endpoints
├── App.vue                   Root component, router-view wrapper
├── main.ts                   App bootstrap
├── router/
│   └── index.ts              Route definitions with guards
├── store/
│   └── userContext.ts        Reactive session store (sessionStorage-backed)
├── types/
│   └── flowGraph.ts          FlowNode, FlowEdge, FlowGraph, NodeDetail, …
├── components/
│   ├── ui/                   Generic reusable components
│   └── workspace/            Workspace-specific components
├── composables/              Logic hooks (useFlowGraph, useMermaid, …)
├── constants/                App-wide constants (languages, phrases, …)
├── utils/                    Pure utility functions (graph traversal, labels)
└── views/                    Page-level route components
```

---

## User Journey

```
/              SplashView
               ↓ "Initialize Workspace"
/onboarding    OnboardingView  (3 steps)
               Step 1: select familiar programming languages
               Step 2: choose experience level
               Step 3: choose workspace source (demo/local/GitHub/ZIP)
               ↓ workspace created
/workspace     WorkspaceView
               ↓ "New Workspace" or logo click
/onboarding
```

The `/workspace` route is guarded: if `userContext.workspaceId` is empty the user is redirected to `/onboarding`.

---

## State Management

`store/userContext.ts` — a single reactive object persisted to `sessionStorage`.

```typescript
interface UserContext {
  primaryLanguage:     string          // comma-separated backend language keys
  experienceLevel:     ExperienceLevel // 'junior' | 'mid' | 'senior'
  languageComparisons: boolean
  workspaceId:         string
  workspaceName:       string
}
```

**Helper exports:**
| Function | Description |
|----------|-------------|
| `normalizeLanguage(lang)` | Returns first valid backend language from comma-separated string |
| `familiarLanguageNames(primaryLanguage)` | Human-readable label list |
| `hasFamiliarLanguages(primaryLanguage)` | Boolean — at least one valid language |
| `clearWorkspace()` | Resets `workspaceId` and `workspaceName` |

---

## Routing

`router/index.ts`

| Path | Name | Component | Guard |
|------|------|-----------|-------|
| `/` | `splash` | `SplashView` | — |
| `/onboarding` | `onboarding` | `OnboardingView` | — |
| `/workspace` | `workspace` | `WorkspaceView` | Requires `workspaceId` |
| `/about` | `about` | `AboutView` | — |

The workspace guard runs in `router.beforeEach`; missing `workspaceId` redirects to `onboarding`.

---

## API Client

`api/index.ts` — thin wrappers around `fetch`.

| Function | Method | Endpoint | Description |
|----------|--------|----------|-------------|
| `createWorkspace(body)` | POST | `/api/workspaces` | Create from local/GitHub/ZIP |
| `fetchTreeEntries(id, path)` | GET | `/api/workspaces/{id}/tree` | Directory listing |
| `fetchFile(id, path)` | GET | `/api/workspaces/{id}/file` | Read file content |
| `fetchSymbols(id, path)` | GET | `/api/workspaces/{id}/symbols` | Extract symbols |
| `graph(payload)` | POST | `/api/workspaces/{id}/graphs` | Build CFG root |
| `expand(payload)` | POST | `/api/workspaces/{id}/graphs/expand` | Expand call node |
| `enrich(payload)` | POST | `/api/workspaces/{id}/graphs/enrich` | AI-label nodes |
| `explain(payload)` | POST | `/api/workspaces/{id}/explain` | Explain a node |
| `explainStream(payload)` | POST | `/api/workspaces/{id}/explain` | SSE streaming explanation |

All non-2xx responses throw an `ApiError(status, message)`.

---

## Views

### `SplashView.vue`

Landing page. Contains:
- Hero section with Bober mascot and tagline.
- Feature cards (scan-first CFG, progressive reveal, coloured branches, AI labels, experience tuning, language analogies).
- 4-step "How It Works" walkthrough.
- A static sample flow SVG (non-interactive).
- CTA button → `/onboarding`.

### `OnboardingView.vue`

Three-step wizard:

**Step 1 — Programming Languages (`FamiliarLanguagesMenu` logic inline):**
- Multi-select checklist of 23 languages from `PROGRAMMING_LANGUAGES`.
- Filters as the user types in the search field.
- Selection stored in `userContext.primaryLanguage` (comma-separated).
- Automatically enables `languageComparisons` when ≥1 language is selected.

**Step 2 — Experience Level (`ExperienceLevelToggle`):**
- Radio buttons: Junior / Mid / Senior.
- Stored in `userContext.experienceLevel`.

**Step 3 — Workspace Source:**
- Tab bar: **Try Demo**, **Local Path**, **GitHub Repo**, **Upload ZIP**.
- **Demo:** Uses `DEMO_REPO` constant (`IBM/sarama`).
- **Local:** Free-text absolute path; validated on the backend.
- **GitHub:** Accepts `owner/repo` shorthand or full HTTPS URL.
- **ZIP:** File picker, max 200 MB.
- Triggers `useWorkspaceSetup.create()` on submit.
- Shows `BeaverRepoLoader` animation (7.5 s) during setup.
- On success, stores `workspaceId` + `workspaceName` in `userContext`, navigates to `/workspace`.

### `WorkspaceView.vue`

Main interface. Composes the full workspace UI:
- Header bar (logo, file name, toggles).
- Sidebar (file explorer).
- Symbol bar (function chips).
- Flow canvas (graph + AI panel).
- Various modals and drawers.

### `AboutView.vue`

Static page with project and team information, pulling from `TEAM_MEMBERS` constant.

---

## Workspace Layout & Panels

`composables/useWorkspaceLayout.ts` — all values persisted to `sessionStorage`.

| State | Default | Range |
|-------|---------|-------|
| `sidebarOpen` | true | — |
| `explorerWidth` | 260 px | 140–400 px |
| `tracePanelOpen` | false | — |
| `traceWidth` | 320 px | 240–480 px |
| `detailPanelOpen` | false | — |
| `detailWidth` | 360 px | 240–520 px |
| `isMobile` | computed | breakpoint < 768 px |

`composables/usePanelResize.ts` handles pointer-driven horizontal drag-to-resize. `ResizeHandle.vue` is the drag target.

---

## File Exploration

### Sidebar (`components/workspace/Sidebar.vue`)

- Fetches top-level directory entries from the backend; lazily expands sub-directories on click.
- `FileTreeNode.vue` renders each entry recursively, using `useFileIcon` to pick the correct VS Code Material Icon SVG.
- `FileSearchBar.vue` provides fuzzy search across the entire file index (`useFileIndex`).

### File Index (`composables/useFileIndex.ts`)

Builds a flat list of all workspace files (lazy, cached per workspace). Used by `FileSearchBar`.

**Search ranking:**
| Match | Rank |
|-------|------|
| Exact name | 0 (highest) |
| Name starts with query | 1 |
| Name contains query | 2 |
| Path contains query | 3 |

### Symbol Bar (`components/workspace/SymbolBar.vue`)

- Horizontal scrollable row of function-name chips.
- Auto-centers the active chip after switching symbols.
- Shows scroll arrows when content overflows.
- Client-side pagination via `useSymbolBrief` (20 symbols per page for files with >30 symbols).
- Selecting a symbol resets the flow graph and calls `loadRoot()`.

---

## Flow Graph System

### Types (`types/flowGraph.ts`)

```typescript
interface FlowNode {
  id: string
  label: string           // scan-time text (raw call name or code)
  title?: string          // enriched short title
  summary?: string        // enriched one-sentence explanation
  kind: 'normal' | 'branch' | 'call'
  confidence: 'verified' | 'inferred'
  labelSource: 'scan' | 'heuristic' | 'ai'
  expandable: boolean
  childCount: number
  collapsed: boolean      // compact callee preview mode
  line?: number
  code?: string
  file?: string
  calleeFile?: string
  calleeSymbol?: string
}

interface FlowEdge {
  from: string
  to: string
  label?: string          // 'true' | 'false' | edge semantic
}

interface FlowGraph {
  rootId: string
  nodes: FlowNode[]
  edges: FlowEdge[]
  depth: number
  symbol: string
  mock?: boolean
}
```

### Flow Graph Composable (`composables/useFlowGraph.ts`)

Central state machine for a single symbol's graph.

**State:**

| Ref | Type | Description |
|-----|------|-------------|
| `allNodes` / `allEdges` | `FlowNode[]` / `FlowEdge[]` | Full graph from backend |
| `nodes` / `edges` | `FlowNode[]` / `FlowEdge[]` | Currently visible subset |
| `revealedIds` | `Set<string>` | Node IDs visible to the user |
| `enrichedIds` | `Set<string>` | Nodes that have been AI-labelled |
| `rootId` | `string` | Entry node ID |
| `loading` | `boolean` | Initial graph load in progress |
| `enriching` | `boolean` | Background enrichment in progress |
| `expanding` | `boolean` | Node expansion in progress |
| `mappingFullFlow` | `boolean` | Full-flow reveal in progress |
| `fullyExpanded` | `boolean` | All nodes revealed and expanded |

**Key methods:**

`loadRoot(payload)` — Checks cache; fetches if miss; calls `entryOnlyRevealedIds()` to show the entry node only; queues enrichment for the horizon.

`revealFromNode(nodeId)` — BFS expansion of the revealed set from the given node. Queues newly revealed nodes for enrichment. Maintains `SILENT_BUFFER_STEPS` lookahead.

`expandNode(node)` — For compact callee nodes: calls `api.expand()`, merges the returned sub-graph into `allNodes`/`allEdges`, updates visibility and enrichment queue.

`enrichNodes(nodeIds)` — Batches nodes in groups of 8, calls `api.enrich()`, applies title/summary/labelSource patches.

`revealFullFlow()` — Iteratively expands all expandable nodes, updating a progress percentage. Sets `fullyExpanded = true` when done.

### Flow Graph Cache (`composables/useFlowGraphCache.ts`)

Per-file, per-symbol cache stored in memory. Key: `v3::{filePath}::{symbol}`.

Stores the full `SymbolFlowState`: all nodes + edges, root ID, revealed IDs, enriched IDs, fully-expanded flag, parent path. Enables instant switch between previously-visited symbols.

### Graph Warming (`composables/useFileFlowWarm.ts`)

When a file is selected, pre-loads graphs for all its symbols in the background.

- 5 concurrent workers.
- Each worker: fetch root → reveal entry → pre-fetch `SILENT_BUFFER_STEPS` (2) + `ENRICHMENT_HORIZON_DEPTH` (2) ahead → enrich visible + buffer nodes in one batch → cache result.
- Marks the file as "warmed" so `FlowWarmOverlay.vue` can indicate completion.

**Constants:**
| Constant | Value | Description |
|----------|-------|-------------|
| `INITIAL_VISIBLE_COUNT` | 1 | Nodes shown on first load |
| `SILENT_BUFFER_STEPS` | 2 | Nodes pre-loaded ahead of frontier |
| `ENRICHMENT_HORIZON_DEPTH` | 2 | Depth beyond visible to pre-enrich |

### Mermaid Rendering (`composables/useFlowMermaid.ts`)

Converts visible `FlowNode[]` + `FlowEdge[]` into a Mermaid `flowchart TD` string, renders it to SVG, and applies post-processing.

**Node shapes:**
- Branch nodes → `{...}` (diamond)
- All others → `[...]` (rectangle)

**Node CSS classes (set by label source + confidence):**
| Class | Condition | Stroke |
|-------|-----------|--------|
| `verified` | scan-time or verified | green, solid |
| `inferred` | AI-labelled, inferred confidence | amber, dashed |
| `heuristic` | heuristic-labelled | cyan, solid |
| `collapsed` | compact callee preview | magenta, solid |

**Edge colours:**
- `true` → green
- `false` → red
- All others → blue

**Rendering lifecycle:**
1. Structural changes (nodes/edges added) → debounced (150 ms) full re-compile.
2. Label-only changes (title/summary patches) → `applyLabelPatches()` updates DOM directly, skipping Mermaid re-compile.
3. Post-render: `styleEdgeLabels()` colours edge-label pills; click bindings attached; selection highlight applied.

### Pan / Zoom (`composables/useFlowPanZoom.ts`)

Wraps the Panzoom library.

| Parameter | Value |
|-----------|-------|
| Min scale | 5 % |
| Max scale | 200 % |
| Step | 12 % |
| Excluded class | `node` (prevents dragging SVG groups) |

Methods: `zoomIn()`, `zoomOut()`, `reset()`, `centerView()`, `getViewport()`, `setViewport(x, y, scale)`.

Viewport is preserved when the graph is structurally rebuilt.

---

## AI Explanations

### Node Detail Composable (`composables/useNodeDetail.ts`)

Fetches and caches explanations for individual nodes.

**Cache key:** `{file}::{symbol}::{nodeId}::{experience}::{languageComparisons}`

**`loadDetail(params)`** — Calls `api.explain()`, waits for full response.

**`loadDetailStream(params)`** — Calls `api.explainStream()`. Parses SSE:
| Event | Payload | Action |
|-------|---------|--------|
| `token` | `{ content: string }` | Append to explanation text |
| `meta` | Partial `NodeDetail` | Patch metadata fields |
| `done` | Full `NodeDetail` | Finalise |
| `error` | Error string | Surface error to UI |

### `NodeDetail` interface

```typescript
interface NodeDetail {
  id: string
  title: string
  summary: string
  explanation: string
  verifiedExplanation?: string
  inferredExplanation?: string
  evidence?: string[]
  relatedSymbols?: string[]
  confidence: 'verified' | 'inferred'
  mock?: boolean      // true when Watsonx is not configured
}
```

When `mock` is true, the UI shows a "Watsonx is not configured" badge and a placeholder explanation.

---

## Workspace Components

### `FlowCanvas.vue`

The centrepiece of the workspace. Hosts the Mermaid flowchart inside a pan/zoom viewport.

**Props:** `nodes`, `edges`, `rootId`, `loading`, `enriching`, `expanding`, `detail`, `selectedNodeId`

**Emits:** `selectNode`, `expandNode`, `revealNode`, `requestDetail`, `viewSource`, `goToDefinition`, `previewCompacted`, `showFullFlow`

Renders:
- Mermaid SVG inside a scrollable, zoomable container.
- Toolbar: legend overlay, help button, zoom controls.
- Resizable right panel (detail drawer).
- Loading overlays for initial load / enrichment / expansion / full-flow mapping.

### `Sidebar.vue`

Left panel with the file tree and search bar. Collapsible on mobile. Emits file-selection events upstream.

### `SymbolBar.vue`

Below the header. Horizontal scrollable row of symbol chips. Auto-centres the active chip; shows prev/next arrows when overflow exists.

### `AiPanel.vue`

Right-side drawer showing a streaming explanation. Renders markdown, Mermaid diagrams embedded in the explanation, VERIFIED/INFERRED badges, and a mock indicator when Watsonx is unavailable.

### `CodePanel.vue`

Modal that shows the full source file in Monaco Editor, optionally scrolling to and highlighting a target line.

### `NodeDetailDrawer.vue`

Slide-in drawer showing the full `NodeDetail`: title, confidence badge, summary, explanation text, related-symbol pills, and a "Jump to code" button.

### `MonacoEditor.vue`

Wraps Monaco Editor loaded from CDN. Read-only. Custom dark theme matching OnBober colour palette. Accepts `language`, `value`, `highlightLine` props.

### `CompactFlowPreview.vue`

Modal triggered when hovering a compact (collapsed) callee node. Fetches the callee's CFG via `api.graph()`, enriches the nodes, and renders a non-interactive Mermaid preview.

### `ExperienceLevelToggle.vue`

Segmented control (Junior / Mid / Senior) in the workspace header. Writes directly to `userContext.experienceLevel`.

### `LanguageComparisonsToggle.vue`

Toggle switch in the workspace header. Writes to `userContext.languageComparisons`.

### `FamiliarLanguagesMenu.vue`

Dropdown panel in the workspace header for editing the user's familiar languages. Syncs with `userContext.primaryLanguage`.

### `FileFlowBrief.vue`

Short helper text shown above the symbol bar when a file is first selected, explaining how to start exploring.

### `FlowWarmOverlay.vue`

Semi-transparent overlay on the canvas shown while background warming is in progress. Displays progress percentage.

### `WorkspaceBobbers.vue` / `WorkspaceBobbersToggle.vue`

See [Bobbers (Easter Egg)](#bobbers-easter-egg).

---

## UI Components

### `Button.vue`

| Prop | Values | Default |
|------|--------|---------|
| `variant` | `'primary'` \| `'ghost'` \| `'outline'` | `'primary'` |
| `disabled` | boolean | false |
| `block` | boolean | false |
| `type` | `'button'` \| `'submit'` | `'button'` |

Primary uses the `onbober-pink` accent colour (`#ff3366`).

### `BeaverLogBar.vue`

Animated SVG of the Bober mascot chewing through a log. Three modes:

| Mode | Description |
|------|-------------|
| `timed` | Progress = elapsed / duration (default 7500 ms) |
| `progress` | Smoothly eases toward an externally provided `externalProgress` (0–100) |
| `indeterminate` | Sine-wave ramp to 90 %, jumps to 100 % when deactivated |

Physics (via `useBeaverLogAnimation`):
- Beaver moves left-to-right across the log.
- Bobbing and chewing use sine-wave modulation.
- Wood-chip particle system: particles spawn at the jaw, arc outward, fade out.
- Shaving sprites are progressively revealed across the log length.
- Clip rects mask the intact vs. gnawed portions.

### `BeaverRepoLoader.vue`

Timed `BeaverLogBar` (7.5 s) shown during workspace creation. Displays rotating setup phrases (ZIP / GitHub / local variants via `useRotatingPhrase`).

### `BeaverFlowLoader.vue`

Compact progress bar for graph loading operations.

### `LoadingStatus.vue`

Generic inline loading indicator with a message slot.

### `Modal.vue`

Centred dialog with overlay. Slot-based content. Emits `close` on overlay click or Escape key.

### `ResizeHandle.vue`

Invisible drag handle for panel resize. Delegates to `usePanelResize`.

### `TeamCreditsBar.vue`

Footer bar listing team members from the `TEAM_MEMBERS` constant.

### `WorkspaceHelpButton.vue`

"?" button that opens a modal explaining the workspace UI.

---

## Composables Reference

| Composable | File | Purpose |
|-----------|------|---------|
| `useWorkspaceSetup` | `useWorkspaceSetup.ts` | Workspace creation (local / GitHub / ZIP) |
| `useWorkspaceLayout` | `useWorkspaceLayout.ts` | Panel widths, sidebar open/close, mobile detection |
| `useFileIndex` | `useFileIndex.ts` | Full file tree index + fuzzy search |
| `useFileTree` | `useFileTree.ts` | Fetch one directory level |
| `useFlowGraph` | `useFlowGraph.ts` | Graph state machine: load, reveal, expand, enrich |
| `useFlowGraphCache` | `useFlowGraphCache.ts` | Per-symbol graph cache |
| `useFlowMermaid` | `useFlowMermaid.ts` | Compile + render Mermaid; patch labels; handle clicks |
| `useFlowPanZoom` | `useFlowPanZoom.ts` | Pan/zoom viewport control |
| `useMermaid` | `useMermaid.ts` | Mermaid init, `compileMermaidSvg`, `extractMermaidBlock` |
| `useNodeDetail` | `useNodeDetail.ts` | Node explanation fetch, stream, cache |
| `useSymbols` | `useSymbols.ts` | Client-side symbol parsing (unused in production path) |
| `useSymbolBrief` | `useSymbolBrief.ts` | Backend symbol fetch + pagination |
| `usePanelResize` | `usePanelResize.ts` | Pointer-driven horizontal drag resize |
| `useShiki` | `useShiki.ts` | Shiki syntax highlighting (on-demand) |
| `useFileIcon` | `useFileIcon.ts` | Map file extension → VS Code Material Icon slug |
| `useFileFlowWarm` | `useFileFlowWarm.ts` | Concurrent background graph warming |
| `useBeaverLogAnimation` | `useBeaverLogAnimation.ts` | Beaver SVG physics + animation |
| `useRotatingPhrase` | `useRotatingPhrase.ts` | Cycle through an array of phrases with configurable interval |
| `useWorkspaceBobbers` | `useWorkspaceBobbers.ts` | localStorage-persisted bobber visibility |

---

## Utils & Types

### `utils/flowGraphUtils.ts`

| Function | Description |
|----------|-------------|
| `bfsNodeIds(nodes, edges, rootId)` | BFS traversal; returns node IDs in breadth-first order |
| `reachableNodeIds(nodes, edges, rootId)` | All IDs reachable from root |
| `pruneGraphToRoot(nodes, edges, rootId)` | Remove unreachable nodes/edges |
| `silentPrefetchTargets(nodes, edges, revealedIds, steps)` | IDs to pre-load ahead of frontier |
| `enrichmentHorizon(nodes, edges, revealedIds, depth)` | IDs beyond visible to pre-enrich |
| `createSymbolFlowState(graph)` | Initialise a `SymbolFlowState` from a backend graph |
| `cloneSymbolFlowState(state)` | Deep clone state for cache storage |
| `revealedPathLength(nodes, edges, rootId, revealedIds)` | Steps revealed so far |
| `visibleFrontierId(nodes, edges, rootId, revealedIds)` | ID of the last revealed node |

### `utils/flowGraphEnrich.ts`

| Function | Description |
|----------|-------------|
| `fetchGraphRoot(payload)` | Calls `api.graph()`, maps response to `FlowGraph` |
| `enrichSymbolNodes(nodes, payload, nodeIds)` | Calls `api.enrich()`, returns `EnrichPatchResult` |

### `utils/flowGraphLabels.ts`

| Function | Description |
|----------|-------------|
| `labelSourcePill(node)` | Short provenance label: `"Auto"` \| `"Brief"` \| `null` |
| `labelSourceBadge(node)` | Long badge: `"Auto label"` \| `"Brief"` \| `null` |
| `hasEnrichedLabel(node)` | `true` if `labelSource` is `'ai'` or `'heuristic'` |
| `isCompactNode(node)` | `true` if collapsed + expandable + callee info present |
| `canPreviewCalleeFlow(node)` | `true` if callee preview modal should be available |

### `types/flowGraph.ts`

Central type definitions: `FlowNode`, `FlowEdge`, `FlowGraph`, `NodeDetail`, `EnrichPatch`, `GraphRootPayload`, `GraphExpandPayload`, `GraphEnrichPayload`, `ExplainPayload`, `SymbolFlowState`.

---

## Constants

| File | Key exports | Description |
|------|-------------|-------------|
| `experienceLevel.ts` | `EXPERIENCE_LEVELS` | `{ value, label, shortLabel, description }[]` |
| `programmingLanguages.ts` | `PROGRAMMING_LANGUAGES`, `languageLabelsFromStored`, `storedLanguagesFromLabels` | 23 languages |
| `demoRepo.ts` | `DEMO_REPO` | `{ name: 'IBM/sarama', url: 'https://github.com/IBM/sarama' }` |
| `projectRepo.ts` | `PROJECT_REPO` | OnBober's own repository URL |
| `beaverLoader.ts` | `BEAVER_DURATION`, `BEAVER_HOLD_DURATION`, `delay` | Animation timing |
| `aiLoadingPhrases.ts` | `AI_LOADING_PHRASES` | 15 beaver-themed phrases for Watsonx wait |
| `flowLoadingPhrases.ts` | `FLOW_LOADING_PHRASES` | Phrases for graph loading |
| `workspaceSetupPhrases.ts` | `ZIP_PHRASES`, `GITHUB_PHRASES`, `LOCAL_PHRASES` | 6 phrases per setup type |
| `teamMembers.ts` | `TEAM_MEMBERS` | `{ name, role, github }[]` |

---

## Personalisation

### Experience Level

Affects both AI label generation (backend) and explanation verbosity (backend prompt construction):

| Level | Labels | Explanations |
|-------|--------|-------------|
| Junior | Plain language, defines jargon | 2–3 sentences, define non-obvious concepts |
| Mid | Balanced | 2–3 sentences, focus on intent |
| Senior | Concise | 1–2 sentences, intent + non-obvious edge cases |

Sent to backend in every `enrich` and `explain` request as `experience`.

### Language Comparisons

When `userContext.languageComparisons = true` and at least one language is selected:
- Backend appends to the explanation prompt: *"If helpful, end with one brief analogy to {langs} (max one short sentence)."*
- Sent as `languageComparisons: true` and `familiarLanguages: "Python, Rust"` (display names).

---

## Bobbers (Easter Egg)

`components/workspace/WorkspaceBobbers.vue`

Renders 5–10 small Bober mascot SVGs positioned absolutely at low z-index around the workspace. Each mascot drifts and bobs via CSS keyframe animations (drift = vertical sine wave, bob = up/down oscillation, peek = corner entrance).

The toggle button in the workspace header (`WorkspaceBobbersToggle.vue`) controls visibility. State is persisted to `localStorage` under the key `onbober:workspace-bobbers`. Defaults to visible.

---

## Styling

### Design Tokens (Tailwind theme extension)

| Token | Value | Usage |
|-------|-------|-------|
| `onbober-primary` | `#ff3366` | Primary CTA, active state accent |
| `slate-950` | `#0f172a` | Main background |
| `slate-900` | `#1a2a3a` | Panel backgrounds |
| `slate-100` | `#e2e8f0` | Default text |
| `slate-500` | `#64748b` | Muted text |
| Verified green | `#4ade80` | True-branch edges, verified badge |
| False red | `#f87171` | False-branch edges |
| Flow blue | `#60a5fa` | Default edges |
| Inferred amber | `#eab308` | Inferred nodes |

### Typography

Font stack: `ui-monospace, "JetBrains Mono", "Fira Code", "Cascadia Code", Consolas, "Courier New", monospace`

Sizes: `text-xs` (12 px) for labels, `text-sm` (14 px) for body, `text-lg`/`text-xl` for headings.

### Responsive

- **< 768 px (mobile):** Sidebar hidden by default; panels stack; drag resize disabled.
- **≥ 768 px (tablet/desktop):** All panels visible; drag resize active.

---

## Build & Development

### Setup

```sh
cd frontend
pnpm install        # also runs postinstall → copy-icons.mjs
```

### Commands

| Command | Description |
|---------|-------------|
| `pnpm dev` | Vite dev server (default port 5173), proxies `/api` to `localhost:8080` |
| `pnpm build` | Production bundle + type-check; output to `dist/` |
| `pnpm preview` | Serve `dist/` locally |
| `pnpm type-check` | `vue-tsc --noEmit` |
| `pnpm lint` | Oxlint + ESLint |
| `pnpm format` | Prettier (all src files) |
| `pnpm test` | Vitest unit tests |

### Key configuration files

| File | Purpose |
|------|---------|
| `vite.config.ts` | Plugins, `@/` alias, dev proxy |
| `tsconfig.app.json` | App TypeScript config (strict) |
| `tsconfig.node.json` | Vite/tooling TypeScript config |
| `.oxlintrc.json` | Oxlint rules |
| `eslint.config.ts` | ESLint + Vue + TypeScript rules |
| `.prettierrc.json` | Prettier options |
| `vitest.config.ts` | Vitest setup |

### `scripts/copy-icons.mjs`

Post-install script. Copies VS Code Material Icons SVGs from `node_modules/vscode-material-icons/generated/icons` → `public/icons/vscode-material-icons/`. Runs automatically via the `postinstall` hook in `package.json`. Icons are gitignored; they must be regenerated after a clean checkout.
