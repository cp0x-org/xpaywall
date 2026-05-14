FROM node:22-alpine AS builder
WORKDIR /app

RUN corepack enable && corepack prepare yarn@4.10.3 --activate

COPY package.json yarn.lock .yarnrc.yml ./
COPY .yarn ./.yarn
RUN yarn install --immutable

COPY . .

ARG VITE_APP_BASE_NAME=/
ENV VITE_APP_BASE_NAME=$VITE_APP_BASE_NAME

RUN yarn build

FROM nginx:1.27-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80
ENTRYPOINT ["/docker-entrypoint.sh"]
