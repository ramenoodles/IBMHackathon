# Deploy to Fly.io

Two Fly apps: **onbober-api** (Go) and **onbober-web** (Vue + nginx). The web app proxies `/api` to the API over Fly's private network.

## One-time setup

### 1. Install Fly CLI

**Windows (PowerShell):**

```powershell
powershell -Command "iwr https://fly.io/install.ps1 -useb | iex"
```

Restart the terminal, then:

```powershell
fly version
```

### 2. Log in

```powershell
fly auth login
```

### 3. Deploy (creates apps on first run)

From the repo root:

```powershell
.\scripts\deploy-fly.ps1
```

### 4. Set Watsonx secrets (required for AI features)

Run **after** the first deploy (the API app must exist):

```powershell
fly secrets set `
  WATSONX_API_KEY="your-key" `
  WATSONX_PROJECT_ID="your-project-id" `
  WATSONX_MODEL="ibm/granite-4-h-small" `
  -a onbober-api
```

Then redeploy the API so machines pick up secrets: `cd backend; fly deploy`

Optional limits (defaults are fine for demos):

```powershell
fly secrets set REQUEST_TIMEOUT_SECONDS=120 CLONE_TIMEOUT_SECONDS=120 -a onbober-api
```

### 4. App names already taken?

If `onbober-api` or `onbober-web` is unavailable, pick new names and update:

- `backend/fly.toml` → `app = 'your-api-name'`
- `frontend/fly.toml` → `app = 'your-web-name'` and `API_UPSTREAM = 'your-api-name.internal:8080'`
- `scripts/deploy-fly.ps1` → `$ApiApp` / `$WebApp` variables

**Demo URL:** https://onbober-web.fly.dev

## Useful commands

```powershell
fly logs -a onbober-api
fly logs -a onbober-web
fly status -a onbober-web
fly secrets list -a onbober-api
fly open -a onbober-web
```

## Local Docker (unchanged)

```powershell
docker compose up --build
```

Uses `frontend/Dockerfile` + `nginx.conf` (Compose service name `backend`). Fly uses `Dockerfile.fly` separately.
