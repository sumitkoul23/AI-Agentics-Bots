# Chain Deployment Studio — Node backend
FROM node:22-alpine

WORKDIR /app

# Install deps first (better layer caching)
COPY server/package.json server/package-lock.json* ./server/
RUN cd server && npm install --omit=dev

# App: the Node server plus the web/ front-end it serves
COPY server ./server
COPY web ./web

ENV PORT=8000
EXPOSE 8000

# Persist deployments outside the image when a volume is mounted at /data
ENV DATA_FILE=/data/deployments.json
VOLUME ["/data"]

WORKDIR /app/server
CMD ["node", "index.js"]
