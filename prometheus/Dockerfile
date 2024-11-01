FROM prom/prometheus:latest

USER root

# Add a non-root user
RUN addgroup -S prometheus && adduser -S -G prometheus prometheus

# Make sure Prometheus directories have the correct permissions
RUN mkdir -p /prometheus/data && \
    chown -R prometheus:prometheus /prometheus/data

# Switch to non-root user
USER prometheus

CMD [ "/bin/prometheus", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/prometheus/data" ]

