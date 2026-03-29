# Override via build-arg when Docker Hub is unreachable (e.g. docker.m.daocloud.io/library/node:20-alpine).
ARG NODE_IMAGE=node:20-alpine
FROM ${NODE_IMAGE}

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

EXPOSE 4173

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "4173"]
