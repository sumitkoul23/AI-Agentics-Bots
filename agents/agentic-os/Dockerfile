# Agentic OS — Python FastAPI gateway
FROM python:3.12-slim

WORKDIR /app

# System deps for Playwright
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl wget gnupg ca-certificates \
    libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 \
    libdrm2 libxkbcommon0 libxcomposite1 libxdamage1 libxrandr2 \
    libgbm1 libpango-1.0-0 libpangocairo-1.0-0 libasound2 \
    && rm -rf /var/lib/apt/lists/*

# Install Python deps
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Install Playwright Chromium browser
RUN playwright install chromium --with-deps || true

# Copy source
COPY . /app/agents/agentic-os/

ENV AGENTIC_DATA_DIR=/data
ENV PYTHONPATH=/app
VOLUME ["/data"]
EXPOSE 8001

WORKDIR /app/agents/agentic-os
CMD ["uvicorn", "gateway.main:app", "--host", "0.0.0.0", "--port", "8001", "--workers", "1"]
