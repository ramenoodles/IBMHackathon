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
