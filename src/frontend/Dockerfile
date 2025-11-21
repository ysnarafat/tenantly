# Build stage
FROM node:25-alpine AS builder

WORKDIR /app

# Copy package files first for better caching
COPY package*.json ./

# Install dependencies (this layer will be cached if package*.json don't change)
RUN npm ci --silent && \
    npm cache clean --force

# Copy source code
COPY . .

# Build the application with optimizations
RUN npm run build -- --configuration=production

# Production stage
FROM nginx:alpine

# Copy built application
COPY --from=builder /app/dist/tenantly-frontend /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]