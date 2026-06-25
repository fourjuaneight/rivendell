FROM python:3.12-slim

# sqlite3 is needed at runtime for sync.sh's ".backup" snapshot. Install at build
# time (the slim image ships none of sqlite3/wget/curl).
RUN apt-get update \
    && apt-get install -y --no-install-recommends sqlite3 \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir \
    datasette \
    datasette-dashboards \
    datasette-vega \
    datasette-copyable

WORKDIR /app
COPY metadata.yml .
COPY sync.sh .
COPY entrypoint.sh .

RUN mkdir -p /data

EXPOSE 8001
ENTRYPOINT ["/app/entrypoint.sh"]
