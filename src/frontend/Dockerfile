# Build stage
FROM node:26-alpine AS builder

WORKDIR /app

COPY package*.json ./

RUN npm ci && npm cache clean --force

COPY . .

# Default to production
ARG BUILD_CONFIGURATION=production

RUN npm run build -- --configuration=${BUILD_CONFIGURATION}

# Production stage
FROM nginx:alpine

COPY --from=builder /app/dist/tenantly-frontend/browser /usr/share/nginx/html

COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]