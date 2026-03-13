FROM node:20-alpine AS builder

WORKDIR /app

# Copy root configurations that might be needed
# Copy the SDK first because dashboard depends on it locally
COPY sdk/js ./sdk/js

# Copy dashboard code
COPY dashboard ./dashboard

# Build the SDK first
WORKDIR /app/sdk/js
RUN npm install
RUN npm run build

# Build the dashboard
WORKDIR /app/dashboard
# Install dependencies, ensuring it uses the local SDK reference
RUN npm install
RUN npm run build

# Production stage
FROM node:20-alpine

WORKDIR /app

# Copy built dashboard from the builder
COPY --from=builder /app/dashboard/build ./build
COPY --from=builder /app/dashboard/package.json ./package.json

# Copy the node_modules from the builder
COPY --from=builder /app/dashboard/node_modules ./node_modules

EXPOSE 3000

# Start the Node server provided by adapter-node
CMD ["node", "build/index.js"]
