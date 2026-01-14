USE travio_inventory;

-- Waitlist table: stores users waiting for seats on a specific trip
-- Partition Key: (organization_id, trip_id) to keep trip queues together
-- Clustering Key: created_at ASC (FIFO queue)
CREATE TABLE IF NOT EXISTS waitlist (
    organization_id text,
    trip_id text,
    user_id text,
    seat_class text, -- Filter by class preference
    requested_seats int,
    created_at timestamp,
    status text, -- 'pending', 'notified', 'expired', 'converted'
    PRIMARY KEY ((organization_id, trip_id), created_at, user_id)
) WITH CLUSTERING ORDER BY (created_at ASC)
  AND compaction = {'class': 'TimeWindowCompactionStrategy', 
                    'compaction_window_unit': 'DAYS', 
                    'compaction_window_size': 1};

-- Materialized View for User lookup (My Waitlist)
-- Allows users to see what waitlists they are on
CREATE MATERIALIZED VIEW IF NOT EXISTS waitlist_by_user AS
    SELECT organization_id, trip_id, user_id, seat_class, requested_seats, created_at, status
    FROM waitlist
    WHERE organization_id IS NOT NULL 
      AND trip_id IS NOT NULL 
      AND created_at IS NOT NULL 
      AND user_id IS NOT NULL
    PRIMARY KEY ((user_id), organization_id, trip_id, created_at);
