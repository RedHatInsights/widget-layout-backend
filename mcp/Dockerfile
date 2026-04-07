# Build stage
FROM registry.access.redhat.com/ubi9/nodejs-20:latest AS builder

USER root

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm ci --only=production && \
    npm cache clean --force

# Copy source code
COPY tsconfig.json ./
COPY src ./src

# Install dev dependencies and build
RUN npm install --only=development && \
    npm run build && \
    npm prune --production

# Production stage
FROM registry.access.redhat.com/ubi9/nodejs-20-minimal:latest

USER root

WORKDIR /app

# Copy package.json for dependency metadata
COPY package.json ./

# Copy production dependencies from builder
COPY --from=builder /app/node_modules ./node_modules

# Copy built application
COPY --from=builder /app/dist ./dist

# Switch to non-root user
USER 1001

# Expose port
EXPOSE 8001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD node -e "require('http').get('http://localhost:8001/healthz', (res) => { process.exit(res.statusCode === 200 ? 0 : 1); });"

# Start the application
CMD ["node", "dist/index.js"]
