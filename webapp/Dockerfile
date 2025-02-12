FROM node:18-alpine AS builder

WORKDIR /app
COPY . .
RUN npm ci
RUN npm run build

FROM nginx:alpine

WORKDIR /app

# Copy built files to Nginx web root
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
