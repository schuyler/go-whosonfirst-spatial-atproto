package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	sanitize "github.com/aaronland/go-http-sanitize"
	"github.com/paulmach/orb/geojson"
	reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial-atproto/lexicon"
	spatial_app "github.com/whosonfirst/go-whosonfirst-spatial/application"
)

type GetRecordHandlerOptions struct{}

func GetRecordHandler(app *spatial_app.SpatialApplication, opts *GetRecordHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		logger := slog.Default()
		ctx := req.Context()

		id, err := sanitize.GetInt64(req, P_RECORD_ID)

		if err != nil {
			logger.Error("Failed to derive record ID", "error", err)
			xrpcError(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		body, err := reader.LoadBytes(ctx, app.SpatialDatabase, id)

		if err != nil {
			logger.Error("Failed to load record", "id", id, "error", err)
			xrpcError(rsp, "Not found", http.StatusNotFound)
			return
		}

		// Parse the GeoJSON bytes
		feature, err := geojson.UnmarshalFeature(body)
		if err != nil {
			logger.Error("Failed to parse GeoJSON", "id", id, "error", err)
			xrpcError(rsp, "Server error", http.StatusInternalServerError)
			return
		}

		// Create place URI
		uri := "at://gazetteer.social/org.whosonfirst.place/" + req.URL.Query().Get(P_RECORD_ID)

		// Convert GeoJSON feature to Place
		place, err := lexicon.PlaceFromGeoJSON(feature, uri)
		if err != nil {
			logger.Error("Failed to convert GeoJSON to Place", "id", id, "error", err)
			xrpcError(rsp, "Server error", http.StatusInternalServerError)
			return
		}

		// Convert lexicon.Place to JSON
		placeJSON, err := json.Marshal(place)
		if err != nil {
			logger.Error("Failed to marshal Place object", "id", id, "error", err)
			xrpcError(rsp, "Server error", http.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-type", "application/json")
		rsp.Write(placeJSON)
	}

	record_handler := http.HandlerFunc(fn)
	return record_handler, nil
}
