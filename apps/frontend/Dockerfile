FROM node:20-alpine

WORKDIR /app

# Copy shared-types first (local file dep)
COPY packages/shared-types/package.json packages/shared-types/package-lock.json* ./packages/shared-types/
RUN cd packages/shared-types && npm install

# Copy and generate TS types from openapi spec
COPY packages/shared-types/openapi.yaml ./packages/shared-types/
RUN cd packages/shared-types && npm run generate:ts

# Install frontend deps
COPY apps/frontend/package.json apps/frontend/package-lock.json* ./apps/frontend/
RUN cd apps/frontend && npm install

# Copy frontend source
COPY apps/frontend/ ./apps/frontend/

WORKDIR /app/apps/frontend

EXPOSE 5173

CMD ["npm", "run", "dev"]
