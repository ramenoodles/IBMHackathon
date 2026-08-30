/** Rotating phrases while a flow graph is loading or expanding. */
export const FLOW_LOADING_PHRASES = [
  'Bob is tracing the execution path…',
  'Gnawing through the call graph…',
  'Building a dam from your control flow…',
  'Mapping branches in the lodge…',
  'Chewing through another function…',
  'Bob is following the CFG upstream…',
  'Stacking steps in the workspace…',
  'Beavering away at your flow map…',
] as const

export const FLOW_INITIAL_STAGES = [
  { threshold: 0, text: 'TRACING EXECUTION PATH...', badge: 'STAGE 1: CFG SCAN' },
  { threshold: 20, text: 'FOLLOWING BRANCHES...', badge: 'STAGE 2: CALL GRAPH' },
  { threshold: 50, text: 'BUILDING FLOW MAP...', badge: 'STAGE 3: MERGING NODES' },
  { threshold: 75, text: 'POSITIONING STEPS...', badge: 'STAGE 4: LAYOUT' },
  { threshold: 100, text: 'RENDERING DIAGRAM...', badge: 'FLOW READY' },
] as const

export const FLOW_FULL_STAGES = [
  { threshold: 0, text: 'EXPANDING CALL GRAPH...', badge: 'STAGE 1: EXPAND' },
  { threshold: 25, text: 'FOLLOWING BRANCHES...', badge: 'STAGE 2: BRANCHES' },
  { threshold: 50, text: 'MERGING SUBFLOWS...', badge: 'STAGE 3: MERGE' },
  { threshold: 75, text: 'FINALIZING NODES...', badge: 'STAGE 4: POLISH' },
  { threshold: 100, text: 'FULL FLOW READY!', badge: 'COMPLETE' },
] as const
