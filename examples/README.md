# Example GeoJSON Files

This directory contains sample GeoJSON files to help you get started with xyzduck.

## Files

### cities.geojson
Point features representing major world cities with properties:
- `name`: City name
- `population`: Population count
- `country`: Country name
- `state`: State/region name

**Example usage:**
```bash
xyzduck init geodata
xyzduck load examples/cities.geojson --db geodata
```

### parks.geojson
Polygon features representing urban parks with properties:
- `name`: Park name
- `area_acres`: Area in acres
- `city`: City where park is located
- `type`: Type of park

**Example usage:**
```bash
xyzduck load examples/parks.geojson --db geodata
```

### routes.geojson
LineString features representing transportation routes with properties:
- `name`: Route description
- `distance_km`: Distance in kilometers
- `mode`: Transportation mode (highway, metro, etc.)
- `route_number`: Route identifier

**Example usage:**
```bash
xyzduck load examples/routes.geojson --db geodata
```

## Sample Queries

After loading the data, you can query it using DuckDB CLI:

```bash
duckdb geodata.duckdb
```

### Query cities by population
```sql
SELECT name, population, country
FROM cities
WHERE population > 5000000
ORDER BY population DESC;
```

### Find parks near cities
```sql
SELECT p.name as park, p.city, c.name as nearest_city,
       ST_Distance(p.geom, c.geom) as distance
FROM parks p
CROSS JOIN cities c
WHERE ST_Distance(p.geom, c.geom) < 1.0
ORDER BY distance;
```

### Calculate route lengths
```sql
SELECT name, distance_km, ST_Length(geom) as calculated_length
FROM routes
ORDER BY distance_km DESC;
```

### Export back to GeoJSON
```sql
COPY (
    SELECT name, population, ST_AsGeoJSON(geom) as geometry
    FROM cities
    WHERE country = 'USA'
) TO 'usa_cities.json';
```

## Creating Your Own GeoJSON Files

GeoJSON files follow this structure:

```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": {
        "type": "Point",
        "coordinates": [longitude, latitude]
      },
      "properties": {
        "key": "value"
      }
    }
  ]
}
```

Supported geometry types:
- Point
- LineString
- Polygon
- MultiPoint
- MultiLineString
- MultiPolygon

## Resources

- [GeoJSON Specification](https://geojson.org/)
- [DuckDB Spatial Documentation](https://duckdb.org/docs/extensions/spatial.html)
- [Find GeoJSON Data](https://github.com/datasets)
