FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps

FROM node:20-alpine AS development
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY package.json package-lock.json ./
# Copy only dashboards files
COPY src/ ./src/
COPY public/ ./public/
COPY config-overrides.js ./
COPY .env* ./
EXPOSE 3000
CMD ["npm", "start"]

FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY package.json package-lock.json ./
# Copy only dashboards files
COPY src/ ./src/
COPY public/ ./public/
COPY config-overrides.js ./
COPY .env* ./

RUN npm run build

FROM nginx:1.27-alpine AS nginx
RUN rm /etc/nginx/conf.d/default.conf && \
    rm -rf /usr/share/nginx/html/*

COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]