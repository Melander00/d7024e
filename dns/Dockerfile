FROM node:24-alpine AS builder

WORKDIR /app

COPY package*.json .

RUN npm ci && npm cache clean --force

COPY . .

CMD ["npm", "run", "start"]