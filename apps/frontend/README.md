# TKAPrep Frontend

Vue 3 + TypeScript + Vite. Pinia for state, vue-router for routing.

## Architecture

```
src/
├── views/              # route-level components
├── components/         # shared UI primitives
├── composables/        # useAuth, useApi, etc.
├── stores/             # Pinia state
├── router/             # vue-router setup with role-based guards
├── api/                # API client (will consume @tkaprep/shared-types)
├── App.vue
└── main.ts
```

## Running locally

```bash
cp .env.example .env
npm install
npm run dev
```

Opens at http://localhost:5173.

## Type generation

API types come from the OpenAPI spec at `../../packages/shared-types/openapi.yaml`.

From repo root:
```bash
make generate-ts
```

Output: `packages/shared-types/generated/ts/index.ts`. Import via `@tkaprep/shared-types`.
