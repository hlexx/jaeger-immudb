FROM ubuntu
ARG APP_DIR=temp-data
ARG QUERY_BIN=${APP_DIR}/query
ENV SPAN_STORAGE_TYPE=grpc-plugin
COPY ui.json ui.json
COPY ${QUERY_BIN} query
COPY plugin-config.yaml cfg/plugin-config.yaml
COPY ${APP_DIR}/jaeger-query jaeger-query
CMD ./jaeger-query --query.ui-config=ui.json --span-storage.type=grpc-plugin --grpc-storage-plugin.binary=query --grpc-storage-plugin.configuration-file=plugin-config.yaml --metrics-backend=none --log-level=debug
