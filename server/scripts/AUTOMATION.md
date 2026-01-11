# Stations Database - Automation Guide

## ✅ **YES**, Everything is Now Automated!

All production improvements will automatically apply in new environments when you run `docker compose up`.

## What Happens Automatically

### On Fresh Environment (`docker compose up -d`)

1. **PostgreSQL Container Starts**
2. **Database Initialization** (`/docker-entrypoint-initdb.d/`)
   - ✅ `create-multiple-postgresql-databases.sh` - Creates all databases
   - ✅ `init-db.sql` - Runs automatically, which includes:
     - Base schema creation
     - **Legacy station seeds** (6 stations)
     - **Import `seed-bd-stations.sql`** (64 Bangladesh stations)
     - **Production constraints** (unique codes, validations)
     - **Performance indexes** (8 indexes)
     - **Triggers** (auto-update timestamps)
     - **Views** (active_stations, stations_by_division)
     - **Documentation** (table/column comments)

### What You Get Out of the Box

```bash
docker compose --env-file .env.dev up -d
```

**Result:**
- 70 total stations (6 legacy + 64 Bangladesh districts)
- 6 constraints enforcing data integrity
- 9 indexes for fast queries
- 2 utility views for common queries
- Auto-updating timestamps
- Full-text search capability

## Verification in New Environment

To verify everything works in a fresh setup:

```bash
# 1. Stop and remove everything
docker compose --env-file .env.dev down -v

# 2. Start fresh
docker compose --env-file .env.dev up -d

# 3. Wait for postgres to be healthy
docker compose --env-file .env.dev ps postgres

# 4. Verify stations loaded
docker exec travio-postgres psql -U postgres -d travio_catalog \
  -c "SELECT COUNT(*) FROM stations WHERE country = 'Bangladesh';"
# Expected: 70

# 5. Verify constraints exist
docker exec travio-postgres psql -U postgres -d travio_catalog \
  -c "SELECT COUNT(*) FROM pg_constraint WHERE conrelid = 'stations'::regclass;"
# Expected: 6

# 6. Verify indexes exist  
docker exec travio-postgres psql -U postgres -d travio_catalog \
  -c "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'stations';"
# Expected: 9

# 7. Test views work
docker exec travio-postgres psql -U postgres -d travio_catalog \
  -c "SELECT division, station_count FROM stations_by_division;"
# Expected: 8 divisions with counts
```

## File Structure

```
server/scripts/
├── init-db.sql                          # ← Main initialization (AUTO-RUNS)
│   ├── Base schemas
│   ├── Imports seed-bd-stations.sql
│   └── Production improvements
├── seed-bd-stations.sql                 # ← Bangladesh stations (AUTO-IMPORTED)
├── seed-bd-stations.py                  # ← Generator script (manual, when updating data)
└── migrations/
    ├── V003__*.sql                      # ← Manual migration (already in init-db.sql)
    └── improvements-summary.md          # ← Documentation
```

## Developer Workflow

### For New Team Members

```bash
git clone <repo>
cd Travio
docker compose --env-file .env.dev up -d
# ✅ Everything ready! 70 stations with all improvements
```

### To Regenerate Station Data

If Bangladesh adds new districts or you need to update coordinates:

```bash
# 1. Update bangladesh_administrative_divisions.json
# 2. Regenerate SQL
python3 server/scripts/seed-bd-stations.py

# 3. Rebuild database
docker compose --env-file .env.dev down -v
docker compose --env-file .env.dev up -d
```

## CI/CD Integration

The setup works perfectly for CI/CD:

```yaml
# .github/workflows/test.yml
- name: Start Database
  run: docker compose --env-file .env.dev up -d postgres
  
- name: Wait for DB
  run: |
    until docker exec travio-postgres pg_isready; do sleep 1; done
    
- name: Verify Stations
  run: |
    count=$(docker exec travio-postgres psql -U postgres -d travio_catalog \
      -tAc "SELECT COUNT(*) FROM stations")
    if [ "$count" -lt "70" ]; then exit 1; fi
```

## Production Deployment

### For Existing Databases (Already Have Data)

If you're deploying to a database that already exists:

1. **Run migration manually** (one-time):
```bash
cat server/scripts/migrations/V003__*.sql | \
  psql -h <prod-host> -U <user> -d travio_catalog
```

2. **Seed new stations** (if needed):
```bash
cat server/scripts/seed-bd-stations.sql | \
  psql -h <prod-host> -U <user> -d travio_catalog
```

### For Fresh Production Setup

Just use the same Docker Compose approach:
```bash
docker compose -f docker-compose.prod.yml up -d
# ✅ Everything auto-applies
```

## What's NOT Automated (By Design)

These require manual intervention for data safety:

1. **Adding new stations** - Use Admin UI or manual INSERT
2. **Updating existing stations** - Use UPDATE queries after review
3. **Dropping constraints** - Manual ALTER TABLE (production safety)

## Troubleshooting

### "Constraint already exists" Error
This is normal on restarts. The `IF NOT EXISTS` clause handles it gracefully.

### "Stations count is 6 not 70"
Check if `seed-bd-stations.sql` import failed:
```bash
docker logs travio-postgres | grep "seed-bd-stations"
```

### "No such file" for seed-bd-stations.sql
Verify volume mount in `docker-compose.yml`:
```yaml
volumes:
  - ./server/scripts/seed-bd-stations.sql:/docker-entrypoint-initdb.d/seed-bd-stations.sql:ro
```

## Summary

**Yes, everything is fully automated!** 🎉

Any developer or CI/CD system can run:
```bash
docker compose up
```

And get a production-ready stations database with:
- ✅ 70 stations
- ✅ All constraints
- ✅ All indexes
- ✅ All views
- ✅ All utilities

**Zero manual steps required!**
