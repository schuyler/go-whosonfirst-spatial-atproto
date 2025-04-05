package lexicon

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/paulmach/orb/geojson"
)

const (
	PlaceTypeID    = "social.gazetteer.place"
	LocationTypeID = PlaceTypeID + "#location"
	ShapeTypeID    = PlaceTypeID + "#shape"
)

// Place represents the main place object in the social.gazetteer.place lexicon
type Place struct {
	Type       string         `json:"$type"`
	URI        string         `json:"uri,omitempty"`
	Location   any            `json:"location,omitempty"`
	Name       string         `json:"name,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Location represents a location with latitude and longitude
type Location struct {
	Type      string `json:"$type"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Altitude  string `json:"altitude,omitempty"`
}

// Shape represents a geometry object with coordinates
type Shape struct {
	Type     string `json:"$type"`
	Geometry string `json:"geometry"` // Changed to string to store serialized geometry
	MimeType string `json:"mimeType,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Place
func (p *Place) UnmarshalJSON(data []byte) error {
	// Create a temporary type to avoid recursion
	type PlaceTemp Place

	// First unmarshal everything except the union type
	var temp PlaceTemp
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Unmarshal the full object to examine the location field
	var objMap map[string]interface{}
	if err := json.Unmarshal(data, &objMap); err != nil {
		return err
	}

	// Get the location field
	locField, ok := objMap["location"]
	if !ok {
		// Location is optional, so we can just return
		*p = Place(temp)
		return nil
	}

	// Determine the location type
	locMap, ok := locField.(map[string]interface{})
	if !ok {
		return fmt.Errorf("location field is not an object")
	}

	// Convert location to the appropriate type
	locData, err := json.Marshal(locField)
	if err != nil {
		return err
	}

	if _, hasLat := locMap["latitude"]; hasLat {
		var loc Location
		if err := json.Unmarshal(locData, &loc); err != nil {
			return err
		}
		// Set the correct type ID
		loc.Type = LocationTypeID
		temp.Location = loc
	} else if _, hasGeometry := locMap["geometry"]; hasGeometry {
		var shape Shape
		if err := json.Unmarshal(locData, &shape); err != nil {
			return err
		}
		// Set the correct type ID
		shape.Type = ShapeTypeID
		temp.Location = shape
	} else {
		return fmt.Errorf("unable to determine location type")
	}

	*p = Place(temp)
	return nil
}

// MarshalJSON implements custom JSON marshaling for Place
func (p Place) MarshalJSON() ([]byte, error) {
	// Create a temporary type to avoid recursion
	type PlaceTemp Place

	// First marshal everything except the union type
	temp := PlaceTemp(p)

	// Make sure $type is set for the Place
	if temp.Type == "" {
		temp.Type = PlaceTypeID
	}

	// Marshal based on the actual type
	switch loc := p.Location.(type) {
	case Location:
		// Make sure $type is set correctly
		if loc.Type == "" {
			loc.Type = LocationTypeID
			temp.Location = loc
		}
	case Shape:
		// Make sure $type is set correctly
		if loc.Type == "" {
			loc.Type = ShapeTypeID
			temp.Location = loc
		}
	case nil:
		// Location is optional
	default:
		return nil, fmt.Errorf("unknown location type: %T", p.Location)
	}

	return json.Marshal(temp)
}

// AsLocation returns the Location as a Location if applicable
func (p *Place) AsLocation() (Location, bool) {
	if loc, ok := p.Location.(Location); ok {
		return loc, true
	}
	return Location{}, false
}

// AsShape returns the Location as a Shape if applicable
func (p *Place) AsShape() (Shape, bool) {
	if shape, ok := p.Location.(Shape); ok {
		return shape, true
	}
	return Shape{}, false
}

// IsLocation determines if the Location is a point
func (p *Place) IsLocation() bool {
	_, ok := p.Location.(Location)
	return ok
}

// IsShape determines if the Location is a shape
func (p *Place) IsShape() bool {
	_, ok := p.Location.(Shape)
	return ok
}

// AsPoint converts the location to an orb.Point
func (l Location) AsPoint() (orb.Point, error) {
	lat, err := strconv.ParseFloat(l.Latitude, 64)
	if err != nil {
		return orb.Point{}, fmt.Errorf("invalid latitude: %s", err)
	}

	lon, err := strconv.ParseFloat(l.Longitude, 64)
	if err != nil {
		return orb.Point{}, fmt.Errorf("invalid longitude: %s", err)
	}

	return orb.Point{lon, lat}, nil
}

// LocationFromPoint creates a Location from an orb.Point
func LocationFromPoint(p orb.Point) Location {
	// Format with 6 decimal places
	lat := fmt.Sprintf("%.6f", p[1])
	lon := fmt.Sprintf("%.6f", p[0])

	return Location{
		Type:      LocationTypeID,
		Latitude:  lat,
		Longitude: lon,
	}
}

// ShapeFromGeoJSON creates a Shape from GeoJSON bytes
func ShapeFromGeoJSON(geojsonBytes []byte) (Shape, error) {
	// Validate that it's valid GeoJSON
	_, err := geojson.UnmarshalGeometry(geojsonBytes)
	if err != nil {
		return Shape{}, fmt.Errorf("invalid GeoJSON: %s", err)
	}

	return Shape{
		Type:     ShapeTypeID,
		Geometry: string(geojsonBytes), // Store as string
		MimeType: "application/geo+json",
	}, nil
}

// ShapeFromWKT creates a Shape from a WKT string
func ShapeFromWKT(wktString string) (Shape, error) {
	// Validate and parse the WKT
	_, err := wkt.Unmarshal(wktString)
	if err != nil {
		return Shape{}, fmt.Errorf("invalid WKT: %s", err)
	}

	return Shape{
		Type:     ShapeTypeID,
		Geometry: wktString, // Store the original WKT string
		MimeType: "application/vnd.geo+wkt",
	}, nil
}

// ShapeFromGeometry creates a Shape from an orb.Geometry
func ShapeFromGeometry(geom orb.Geometry) (Shape, error) {
	// Convert to GeoJSON
	geometry := geojson.NewGeometry(geom)
	bytes, err := json.Marshal(geometry)
	if err != nil {
		return Shape{}, fmt.Errorf("failed to marshal geometry to GeoJSON: %s", err)
	}

	return Shape{
		Type:     ShapeTypeID,
		Geometry: string(bytes), // Store as string
		MimeType: "application/geo+json",
	}, nil
}

// AsGeometry parses the Shape's geometry as GeoJSON or WKT and returns an orb.Geometry
func (s Shape) AsGeometry() (orb.Geometry, error) {
	if s.MimeType == "" {
		// Try to guess the format based on the first character
		if len(s.Geometry) > 0 && (s.Geometry[0] == '{' || s.Geometry[0] == '[') {
			// Looks like JSON, assume GeoJSON
			geom, err := geojson.UnmarshalGeometry([]byte(s.Geometry))
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal GeoJSON geometry: %s", err)
			}
			return geom.Geometry(), nil
		} else {
			// Try as WKT
			geom, err := wkt.Unmarshal(s.Geometry)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal WKT geometry: %s", err)
			}
			return geom, nil
		}
	} else if s.MimeType == "application/geo+json" {
		// GeoJSON format
		geom, err := geojson.UnmarshalGeometry([]byte(s.Geometry))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal GeoJSON geometry: %s", err)
		}
		return geom.Geometry(), nil
	} else if s.MimeType == "application/vnd.geo+wkt" {
		// WKT format
		geom, err := wkt.Unmarshal(s.Geometry)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal WKT geometry: %s", err)
		}
		return geom, nil
	}

	return nil, fmt.Errorf("unsupported mime type: %s", s.MimeType)
}

// PlaceFromGeoJSON creates a Place from a GeoJSON Feature
func PlaceFromGeoJSON(feature *geojson.Feature, uri string) (Place, error) {
	place := Place{
		Type:       PlaceTypeID,
		URI:        uri,
		Attributes: make(map[string]any),
	}

	// Extract the name from properties
	if name, ok := feature.Properties["wof:name"].(string); ok {
		place.Name = name
	} else if name, ok := feature.Properties["name"].(string); ok {
		place.Name = name
	}

	// Extract attributes from properties
	for k, v := range feature.Properties {
		place.Attributes[k] = v
	}

	// Set location based on geometry type
	geom := feature.Geometry
	if geom != nil {
		if geom.GeoJSONType() == "Point" {
			// For point geometry, create a Location
			point := geom.(orb.Point)
			place.Location = LocationFromPoint(point)
		} else {
			// For non-point geometry, create a Shape
			shape, err := ShapeFromGeometry(geom)
			if err != nil {
				return Place{}, fmt.Errorf("failed to create shape: %s", err)
			}
			place.Location = shape
		}
	}

	return place, nil
}
