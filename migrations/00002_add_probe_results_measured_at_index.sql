-- +goose Up
CREATE INDEX idx_probe_results_measured_at
    ON probe_results (measured_at);

-- +goose Down
DROP INDEX IF EXISTS idx_probe_results_measured_at;
