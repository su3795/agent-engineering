# ============================================================
#  Agent Engineering Platform V2 — 多阶段构建
#  用法:
#    docker build -t agent-engineering:v2 .
#    docker run -p 80:80 -p 8000:8000 agent-engineering:v2
# ============================================================
FROM node:20-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM python:3.11-slim AS backend
WORKDIR /app
COPY backend/requirements-min.txt ./
RUN pip install --no-cache-dir -r requirements-min.txt
COPY backend/ ./
# 将前端构建产物放到 backend/static/, 由 main_v2.py 挂载为 /static
COPY --from=frontend-build /app/frontend/dist ./static

EXPOSE 8000
CMD ["uvicorn", "main_v2:app", "--host", "0.0.0.0", "--port", "8000"]
