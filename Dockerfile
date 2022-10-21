FROM ubuntu
ARG APP_DIR=temp-data
ENV SPAN_STORAGE_TYPE=badger
COPY sampling.json sampling.json
COPY ${APP_DIR}/jaeger-collector jaeger-collector
CMD ./jaeger-collector --collector.num-workers=100 --collector.queue-size=30000  --collector.zipkin.host-port=9411 --sampling.strategies-file=sampling.json --badger.ephemeral=false --badger.directory-value=bdata/value --badger.directory-key=bdata/key --badger.span-store-ttl=24h --metrics-backend=none & wait
