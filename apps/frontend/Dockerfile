# ─── Stage 1: Build ────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

# Shared-types: install deps + generate TS types from OpenAPI spec
COPY packages/shared-types/package.json packages/shared-types/package-lock.json* ./packages/shared-types/
RUN cd packages/shared-types && npm install

COPY packages/shared-types/openapi.yaml ./packages/shared-types/
RUN cd packages/shared-types && npm run generate:ts

# Frontend: install deps
COPY apps/frontend/package.json apps/frontend/package-lock.json* ./apps/frontend/
RUN cd apps/frontend && npm install

# Frontend: copy source and build
COPY apps/frontend/ ./apps/frontend/

ARG VITE_API_BASE_URL=https://api.masukelas.my.id/api/v1
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN cd apps/frontend && npm run build

# ─── Stage 2: Serve with nginx ─────────────────────────────────────────────
FROM nginx:alpine

# Copy built SPA
COPY --from=builder /app/apps/frontend/dist /usr/share/nginx/html

# nginx config: serve SPA, forward /api to backend
COPY apps/frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
