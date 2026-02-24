FROM node:24-alpine AS build
WORKDIR /app

COPY package*.json .
RUN npm ci

ARG API_BASE_URL
ENV API_BASE_URL=${API_BASE_URL}

COPY . .
RUN sed -i "s|__API_BASE_URL__|${API_BASE_URL}|g" src/environments/environment.ts \
  && npm run build -- --configuration production

FROM nginxinc/nginx-unprivileged:1.27-alpine AS runtime
COPY --from=build /app/dist/spotlite/browser /usr/share/nginx/html
EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
