# Backend API

`cmd/api` is the only API server. It provides workspace-scoped JSON routes:

- `GET /api/health`
- `POST /api/workspaces`
- `GET /api/workspaces/{id}/tree`
- `GET /api/workspaces/{id}/file`
- `GET /api/workspaces/{id}/symbols`
- `POST /api/workspaces/{id}/graphs`
- `POST /api/workspaces/{id}/graphs/expand`
- `POST /api/workspaces/{id}/explain`

Workspace IDs are opaque and filesystem paths are never returned to clients.
Graph analysis works without Watsonx. Explanations require the Watsonx
environment variables documented in the repository README.

Graph roots are deterministic, function-local control-flow graphs for the
requested file and symbol. Root responses contain at most eight nodes;
resolved call nodes can be expanded into callee fragments containing at most
six nodes, through four call levels.
