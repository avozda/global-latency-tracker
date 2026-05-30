CREATE TABLE probe_results (
    id BIGSERIAL PRIMARY KEY,
    target_url TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    dns_lookup_ms DOUBLE PRECISION NOT NULL,
    tcp_connection_ms DOUBLE PRECISION NOT NULL,
    tls_handshake_ms DOUBLE PRECISION NOT NULL,
    server_processing_ms DOUBLE PRECISION NOT NULL,
    ttfb_ms DOUBLE PRECISION NOT NULL,
    total_roundtrip_ms DOUBLE PRECISION NOT NULL,
    measured_at TIMESTAMPTZ NOT NULL,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_probe_results_target_url_measured_at
    ON probe_results (target_url, measured_at DESC);
