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

# Run as the image's built-in non-root nginx user. Listens on 8080 (not 80)
# since binding a privileged port requires root.
RUN chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/run /var/log/nginx

USER nginx

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
