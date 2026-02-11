FROM node:20-alpine AS build
WORKDIR /app/freebooru

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY freebooru/package.json freebooru/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY freebooru/ ./
RUN pnpm build

FROM nginx:1.27-alpine


COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=build /app/freebooru/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
