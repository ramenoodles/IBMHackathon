# OnBober

OnBober is a repository onboarding UI backed by one Go API server.

## Development

Requirements: Go 1.22, Node.js, npm, Git for GitHub sources, and ripgrep.

```sh
cd backend
go run ./cmd/api

cd frontend
npm install
npm run dev
```

The frontend proxies `/api` to `http://localhost:8080`.

Configure optional Watsonx explanations with `WATSONX_API_KEY`,
`WATSONX_PROJECT_ID`, and `WATSONX_MODEL`. Server settings are `HOST`, `PORT`,
and `RG_BINARY`.
