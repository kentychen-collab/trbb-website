package models

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	SportRun   = 1
	SportSwim  = 2
	SportBike  = 3
	SportBrick = 4
	SportGym   = 5
	SportOther = 6
)

// LatLng for route points
type LatLng [2]float64

// RoutePoints stores decoded GPX coordinates
type RoutePoints []LatLng

func (r RoutePoints) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}
	return json.Marshal([]LatLng(r))
}
func (r *RoutePoints) Scan(src any) error {
	if src == nil {
		*r = nil
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
	return json.Unmarshal(b, r)
}

type TrainingLog struct {
	ID               uint64      `json:"id"`
	UUID             string      `json:"uuid"`
	UserID           uint64      `json:"user_id"`
	Title            string      `json:"title"`
	SportType        int         `json:"sport_type"`
	Date             string      `json:"date"` // YYYY-MM-DD
	DurationMin      *int        `json:"duration_min"`
	DistanceKm       *float64    `json:"distance_km"`
	AvgHeartRate     *int        `json:"avg_heart_rate"`
	MaxHeartRate     *int        `json:"max_heart_rate"`
	Calories         *int        `json:"calories"`
	ElevationM       *int        `json:"elevation_m"`
	AvgPace          string      `json:"avg_pace"`
	AvgSpeedKph      *float64    `json:"avg_speed_kph"`
	PowerAvg         *int        `json:"power_avg"`
	Note             string      `json:"note"`
	IsPublic         bool        `json:"is_public"`
	Photos           StringSlice `json:"photos"` // reuse from shop model
	GpxFilePath      string      `json:"gpx_file_path,omitempty"`
	FitFilePath      string      `json:"fit_file_path,omitempty"`
	RoutePoints      RoutePoints `json:"route_points,omitempty"`
	StartLat         *float64    `json:"start_lat,omitempty"`
	StartLng         *float64    `json:"start_lng,omitempty"`
	Source           string      `json:"source"` // manual|gpx|fit|garmin
	GarminActivityID string      `json:"garmin_activity_id,omitempty"`
	DescentM         *int        `json:"descent_m,omitempty"`
	StartTime        *time.Time  `json:"start_time,omitempty"`
	MapThumbnailURL  string      `json:"map_thumbnail_url,omitempty"`
	CoverURL         string      `json:"cover_url,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`

	// joined
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// GarminToken stores OAuth 1.0a tokens
type GarminToken struct {
	ID            uint64     `json:"id"`
	UserID        uint64     `json:"user_id"`
	AccessToken   string     `json:"-"`
	TokenSecret   string     `json:"-"`
	GarminUserID  string     `json:"garmin_user_id"`
	Scope         string     `json:"scope"`
	LastSyncAt    *time.Time `json:"last_sync_at"`
	SyncPublic    bool       `json:"sync_public"`
	AutoSync      bool       `json:"auto_sync"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// StravaToken stores OAuth 2.0 tokens for a user's Strava connection
type StravaToken struct {
	ID              uint64     `json:"id"`
	UserID          uint64     `json:"user_id"`
	AccessToken     string     `json:"-"`
	RefreshToken    string     `json:"-"`
	TokenType       string     `json:"token_type"`
	ExpiresAt       *time.Time `json:"expires_at"`
	StravaAthleteID int64      `json:"strava_athlete_id"`
	AthleteName     string     `json:"athlete_name"`
	LastSyncAt      *time.Time `json:"last_sync_at"`
	SyncPublic      bool       `json:"sync_public"`
	AutoSync        bool       `json:"auto_sync"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
