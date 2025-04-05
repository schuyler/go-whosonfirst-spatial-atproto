package lexicon

import (
	"encoding/json"
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

func TestPlace(t *testing.T) {
	t.Run("Marshal and Unmarshal Place with Location", func(t *testing.T) {
		// Create a place with a Location
		original := Place{
			Type: PlaceTypeID,
			URI:  "at://gazetteer.social/social.gazetteer.place/123",
			Location: Location{
				Type:      LocationTypeID,
				Latitude:  "37.7749",
				Longitude: "-122.4194",
			},
			Name: "San Francisco",
		}

		// Marshal to JSON
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal Place: %v", err)
		}

		// Unmarshal back
		var decoded Place
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("Failed to unmarshal Place: %v", err)
		}

		// Verify fields
		if decoded.Type != PlaceTypeID {
			t.Errorf("Expected Type %s, got %s", PlaceTypeID, decoded.Type)
		}
		if decoded.URI != original.URI {
			t.Errorf("Expected URI %s, got %s", original.URI, decoded.URI)
		}
		if decoded.Name != original.Name {
			t.Errorf("Expected Name %s, got %s", original.Name, decoded.Name)
		}

		// Check location
		loc, ok := decoded.AsLocation()
		if !ok {
			t.Fatalf("Expected Location, but IsLocation() returned false")
		}
		origLoc, _ := original.AsLocation()
		if loc.Latitude != origLoc.Latitude {
			t.Errorf("Expected Latitude %s, got %s", origLoc.Latitude, loc.Latitude)
		}
		if loc.Longitude != origLoc.Longitude {
			t.Errorf("Expected Longitude %s, got %s", origLoc.Longitude, loc.Longitude)
		}
	})

	t.Run("Marshal and Unmarshal Place with Shape", func(t *testing.T) {
		// Create a GeoJSON polygon
		polygon := orb.Polygon{
			{{-122.5, 37.7}, {-122.4, 37.7}, {-122.4, 37.8}, {-122.5, 37.8}, {-122.5, 37.7}},
		}
		geometry := geojson.NewGeometry(polygon)
		geomBytes, _ := json.Marshal(geometry)

		// Create a place with a Shape
		original := Place{
			Type: PlaceTypeID,
			URI:  "at://gazetteer.social/social.gazetteer.place/456",
			Location: Shape{
				Type:     ShapeTypeID,
				Geometry: geomBytes,
				MimeType: "application/geo+json",
			},
			Name: "San Francisco Bay Area",
		}

		// Marshal to JSON
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal Place with Shape: %v", err)
		}

		// Unmarshal back
		var decoded Place
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("Failed to unmarshal Place with Shape: %v", err)
		}

		// Verify shape
		shape, ok := decoded.AsShape()
		if !ok {
			t.Fatalf("Expected Shape, but IsShape() returned false")
		}
		if shape.MimeType != "application/geo+json" {
			t.Errorf("Expected MimeType application/geo+json, got %s", shape.MimeType)
		}
	})

	t.Run("Missing Location", func(t *testing.T) {
		// Place without location
		original := Place{
			Type: PlaceTypeID,
			URI:  "at://gazetteer.social/social.gazetteer.place/789",
			Name: "Somewhere",
		}

		// Marshal and unmarshal
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal Place without Location: %v", err)
		}

		var decoded Place
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("Failed to unmarshal Place without Location: %v", err)
		}

		// Location should be nil
		if decoded.Location != nil {
			t.Errorf("Expected nil Location, got %T", decoded.Location)
		}
	})
}

func TestLocation(t *testing.T) {
	t.Run("AsPoint", func(t *testing.T) {
		loc := Location{
			Type:      LocationTypeID,
			Latitude:  "37.7749",
			Longitude: "-122.4194",
		}

		point, err := loc.AsPoint()
		if err != nil {
			t.Fatalf("Failed to convert Location to Point: %v", err)
		}

		expectedLon := -122.4194
		expectedLat := 37.7749
		if point[0] != expectedLon || point[1] != expectedLat {
			t.Errorf("Expected point [%f, %f], got [%f, %f]", expectedLon, expectedLat, point[0], point[1])
		}
	})

	t.Run("LocationFromPoint", func(t *testing.T) {
		point := orb.Point{-122.4194, 37.7749}
		loc := LocationFromPoint(point)

		if loc.Type != LocationTypeID {
			t.Errorf("Expected Type %s, got %s", LocationTypeID, loc.Type)
		}
		if loc.Latitude != "37.774900" {
			t.Errorf("Expected Latitude 37.774900, got %s", loc.Latitude)
		}
		if loc.Longitude != "-122.419400" {
			t.Errorf("Expected Longitude -122.419400, got %s", loc.Longitude)
		}
	})

	t.Run("Invalid Location", func(t *testing.T) {
		// Test with invalid latitude
		loc := Location{
			Type:      LocationTypeID,
			Latitude:  "invalid",
			Longitude: "-122.4194",
		}

		_, err := loc.AsPoint()
		if err == nil {
			t.Error("Expected error for invalid latitude, got nil")
		}

		// Test with invalid longitude
		loc = Location{
			Type:      LocationTypeID,
			Latitude:  "37.7749",
			Longitude: "invalid",
		}

		_, err = loc.AsPoint()
		if err == nil {
			t.Error("Expected error for invalid longitude, got nil")
		}
	})
}

func TestShape(t *testing.T) {
	t.Run("GeoJSON Shape", func(t *testing.T) {
		// Create a GeoJSON polygon
		polygon := orb.Polygon{
			{{-122.5, 37.7}, {-122.4, 37.7}, {-122.4, 37.8}, {-122.5, 37.8}, {-122.5, 37.7}},
		}

		shape, err := ShapeFromGeometry(polygon)
		if err != nil {
			t.Fatalf("Failed to create Shape from geometry: %v", err)
		}

		if shape.Type != ShapeTypeID {
			t.Errorf("Expected Type %s, got %s", ShapeTypeID, shape.Type)
		}
		if shape.MimeType != "application/geo+json" {
			t.Errorf("Expected MimeType application/geo+json, got %s", shape.MimeType)
		}

		// Convert back to geometry
		geom, err := shape.AsGeometry()
		if err != nil {
			t.Fatalf("Failed to convert Shape to geometry: %v", err)
		}

		// Verify it's a polygon
		poly, ok := geom.(orb.Polygon)
		if !ok {
			t.Fatalf("Expected Polygon geometry, got %T", geom)
		}

		// Basic validation of the polygon
		if len(poly) != 1 || len(poly[0]) != 5 {
			t.Errorf("Polygon has unexpected structure: %v", poly)
		}
	})

	t.Run("WKT Shape", func(t *testing.T) {
		wktString := "POLYGON((-122.5 37.7, -122.4 37.7, -122.4 37.8, -122.5 37.8, -122.5 37.7))"
		shape, err := ShapeFromWKT(wktString)
		if err != nil {
			t.Fatalf("Failed to create Shape from WKT: %v", err)
		}

		if shape.Type != ShapeTypeID {
			t.Errorf("Expected Type %s, got %s", ShapeTypeID, shape.Type)
		}
		if shape.MimeType != "application/vnd.geo+wkt" {
			t.Errorf("Expected MimeType application/vnd.geo+wkt, got %s", shape.MimeType)
		}

		// Convert back to geometry
		geom, err := shape.AsGeometry()
		if err != nil {
			t.Fatalf("Failed to convert WKT Shape to geometry: %v", err)
		}

		// Verify it's a polygon
		_, ok := geom.(orb.Polygon)
		if !ok {
			t.Fatalf("Expected Polygon geometry from WKT, got %T", geom)
		}
	})

	t.Run("Invalid Shapes", func(t *testing.T) {
		// Test with invalid GeoJSON
		_, err := ShapeFromGeoJSON([]byte(`{invalid json}`))
		if err == nil {
			t.Error("Expected error for invalid GeoJSON, got nil")
		}

		// Test with invalid WKT
		_, err = ShapeFromWKT("NOT A WKT STRING")
		if err == nil {
			t.Error("Expected error for invalid WKT, got nil")
		}

		// Test with unsupported mime type
		shape := Shape{
			Type:     ShapeTypeID,
			Geometry: []byte(`{}`),
			MimeType: "application/unsupported",
		}
		_, err = shape.AsGeometry()
		if err == nil {
			t.Error("Expected error for unsupported mime type, got nil")
		}
	})
}
