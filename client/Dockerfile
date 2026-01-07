FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install pnpm and dependencies
RUN npm install -g pnpm && pnpm install --frozen-lockfile

# Copy source
COPY . .

# Build arg for API URL
ARG VITE_API_URL
ENV VITE_API_URL=${VITE_API_URL}

# Build the app
RUN pnpm build

# Production stage
FROM node:20-alpine

WORKDIR /app

# Install pnpm
RUN npm install -g pnpm

# Copy the prepared app with build output and dependencies
COPY --from=builder /app ./

EXPOSE 5173

# Run in preview mode for development
CMD ["pnpm", "preview", "--host", "0.0.0.0", "--port", "5173"]
