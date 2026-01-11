#!/usr/bin/env python3
"""
Bangladesh Stations Seed Data Generator

Generates SQL INSERT statements for all Bangladesh district headquarters
as stations in the Travio catalog database.

Usage:
    python3 seed-bd-stations.py

Output:
    server/scripts/seed-bd-stations.sql
"""

import json
import uuid
import hashlib
from pathlib import Path

# Division ID to Name mapping
DIVISION_MAP = {
    "1": "Chattogram",
    "2": "Rajshahi", 
    "3": "Khulna",
    "4": "Rangpur",
    "5": "Barishal",
    "6": "Dhaka",
    "7": "Sylhet",
    "8": "Mymensingh"
}

def generate_station_code(name: str, existing_codes: set) -> str:
    """
    Generate a 3-letter IATA-style station code.
    
    Args:
        name: District/City name
        existing_codes: Set of already used codes
        
    Returns:
        Unique 3-letter uppercase code
    """
    # Remove common suffixes and clean name
    clean_name = name.replace(" Sadar", "").replace(" District", "").strip()
    
    # Try first 3 letters
    base_code = clean_name[:3].upper()
    
    if base_code not in existing_codes:
        return base_code
    
    # Try first letter + next 2 consonants
    consonants = [c for c in clean_name.upper() if c.isalpha() and c not in 'AEIOU']
    if len(consonants) >= 3:
        alt_code = consonants[0] + consonants[1] + consonants[2]
        if alt_code not in existing_codes:
            return alt_code
    
    # Fallback: add numeric suffix
    counter = 2
    while f"{base_code}{counter}" in existing_codes:
        counter += 1
    
    return f"{base_code}{counter}"[:3]

def generate_uuid_from_district_id(district_id: str) -> str:
    """
    Generate deterministic UUID from district ID.
    
    Args:
        district_id: District ID from BD data
        
    Returns:
        UUID string
    """
    # Create deterministic UUID using namespace UUID5
    namespace = uuid.UUID('6ba7b810-9dad-11d1-80b4-00c04fd430c8')  # DNS namespace
    seed = f"bd-district-station-{district_id}"
    return str(uuid.uuid5(namespace, seed))

def main():
    # Load Bangladesh administrative divisions data
    data_path = Path(__file__).parent.parent / "data" / "bangladesh_administrative_divisions.json"
    
    with open(data_path, 'r', encoding='utf-8') as f:
        bd_data = json.load(f)
    
    districts = bd_data['districts']
    
    # Track used codes to avoid conflicts
    existing_codes = set()
    
    # Generate station entries
    stations = []
    
    for district in districts:
        district_id = district['id']
        division_id = district['division_id']
        name = district['name']
        bn_name = district['bn_name']
        lat = district['lat']
        lon = district['lon']
        division_name = DIVISION_MAP.get(division_id, "Unknown")
        
        # Generate unique code
        code = generate_station_code(name, existing_codes)
        existing_codes.add(code)
        
        # Generate deterministic UUID
        station_uuid = generate_uuid_from_district_id(district_id)
        
        # Create station entry
        station = {
            'id': station_uuid,
            'code': code,
            'name': name,
            'bn_name': bn_name,
            'city': name,
            'state': division_name,
            'country': 'Bangladesh',
            'latitude': lat,
            'longitude': lon,
            'status': 'active'
        }
        
        stations.append(station)
    
    # Generate SQL file
    output_path = Path(__file__).parent / "seed-bd-stations.sql"
    
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write("-- Bangladesh Stations Seed Data\n")
        f.write("-- Generated from bangladesh_administrative_divisions.json\n")
        f.write(f"-- Total Stations: {len(stations)}\n")
        f.write("-- ============================================================================\n\n")
        
        f.write("\\c travio_catalog\n\n")
        
        f.write("-- Insert all Bangladesh district headquarters as stations\n")
        f.write("INSERT INTO stations (id, code, name, city, state, country, latitude, longitude, status) VALUES\n")
        
        # Write INSERT values
        for i, station in enumerate(stations):
            comma = "," if i < len(stations) - 1 else ";"
            
            f.write(f"  ('{station['id']}', ")
            f.write(f"'{station['code']}', ")
            f.write(f"'{station['name']}', ")
            f.write(f"'{station['city']}', ")
            f.write(f"'{station['state']}', ")
            f.write(f"'{station['country']}', ")
            f.write(f"{station['latitude']}, ")
            f.write(f"{station['longitude']}, ")
            f.write(f"'{station['status']}'){comma}")
            
            # Add comment with Bangla name
            f.write(f"  -- {station['bn_name']}\n")
        
        f.write("\nON CONFLICT (id) DO UPDATE SET\n")
        f.write("  name = EXCLUDED.name,\n")
        f.write("  latitude = EXCLUDED.latitude,\n")
        f.write("  longitude = EXCLUDED.longitude,\n")
        f.write("  updated_at = NOW();\n")
        
        f.write("\n-- Verification query\n")
        f.write("-- SELECT state, COUNT(*) as station_count FROM stations WHERE country = 'Bangladesh' GROUP BY state ORDER BY state;\n")
    
    # Print summary
    print(f"✅ Generated SQL seed file: {output_path}")
    print(f"📊 Total Stations: {len(stations)}")
    print(f"\\n📍 Stations by Division:")
    
    division_counts = {}
    for station in stations:
        div = station['state']
        division_counts[div] = division_counts.get(div, 0) + 1
    
    for div in sorted(division_counts.keys()):
        print(f"   {div}: {division_counts[div]} stations")
    
    print(f"\\n✨ Sample Station Codes:")
    for station in stations[:10]:
        print(f"   {station['code']} - {station['name']} ({station['bn_name']})")

if __name__ == "__main__":
    main()
