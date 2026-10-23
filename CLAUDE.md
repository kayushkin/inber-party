# About inber-party

## Status and scope

inber-party is retired and stopped. On 2026-10-23 `repo:agent-party` replaced it at `party.kayushkin.com`; its unit is disabled and healthcheck no longer watches it. Its 25 unmerged test and fix branches are pushed to GitHub and kept, unmerged. Do not restart it or add to it; new work goes in agent-party.

# How it works

## Runtime and data

The Go server provides the REST and WebSocket APIs and serves the React build from `frontend/dist`. The live unit reads Inber sessions from `~/inber-workspace/.inber/sessions.db` in read-only mode. PostgreSQL is optional and is not configured on this host, so the service runs in inber-only mode and database-backed write features stay unavailable. Do not invent a database fallback or add `DATABASE_URL` unless someone has first chosen and provisioned PostgreSQL.

# Access and operations

## Live service

`inber-party.service` is disabled and stopped. `:8180` and `party.kayushkin.com` now belong to `agent-party.service`. The unit template `deploy/inber-party.service` remains only as a record.

## The working directory is part of production

Moot since the service stopped: it opened `frontend/dist` by relative path from `WorkingDirectory=~/repos/inber-party`. agent-party embeds its page instead.

# Working in this repo

## Build, test and deploy

Run `go test ./...` for the backend. In `frontend/`, use `npm ci`, `npm run build`, and `npm run test:e2e` as needed. Deploy only from a clean, pushed default branch with `./deploy.sh`; it runs the shared deploy gate, builds both halves, installs the unit and binary, restarts the user service, and checks readiness. Do not deploy this retired service as part of unrelated work.
