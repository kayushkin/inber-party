# About inber-party

## Status and scope

inber-party is a retired operator-facing RPG dashboard. It still runs at `party.kayushkin.com`, but only maintenance and shutdown work belong here. Do not add features, new integrations, or new consumers; put new work in the active replacement service.

# How it works

## Runtime and data

The Go server provides the REST and WebSocket APIs and serves the React build from `frontend/dist`. The live unit reads Inber sessions from `~/inber-workspace/.inber/sessions.db` in read-only mode. PostgreSQL is optional and is not configured on this host, so the service runs in inber-only mode and database-backed write features stay unavailable. Do not invent a database fallback or add `DATABASE_URL` unless someone has first chosen and provisioned PostgreSQL.

# Access and operations

## Live service

`inber-party.service` is an enabled user unit. It runs `~/bin/inber-party` on `:8180`; nginx exposes it as `party.kayushkin.com`. The unit template is `deploy/inber-party.service`, and `/health` is the readiness endpoint.

## The working directory is part of production

The unit uses `WorkingDirectory=~/repos/inber-party`, and the server opens `frontend/dist` by relative path. Keep the main clone on the default branch. Moving the working directory can leave `/health` green while every frontend asset returns 404.

# Working in this repo

## Build, test and deploy

Run `go test ./...` for the backend. In `frontend/`, use `npm ci`, `npm run build`, and `npm run test:e2e` as needed. Deploy only from a clean, pushed default branch with `./deploy.sh`; it runs the shared deploy gate, builds both halves, installs the unit and binary, restarts the user service, and checks readiness. Do not deploy this retired service as part of unrelated work.
