-- Add outbox table for reliable event publishing
CREATE TABLE IF NOT EXISTS event_outbox (
    id UUID PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    topic VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    retries INT DEFAULT 0
);

-- Index for efficient polling of unprocessed events
CREATE INDEX IF NOT EXISTS idx_outbox_unprocessed 
ON event_outbox (created_at) 
WHERE processed_at IS NULL;

-- Index for cleanup of processed events
CREATE INDEX IF NOT EXISTS idx_outbox_processed 
ON event_outbox (processed_at) 
WHERE processed_at IS NOT NULL;
