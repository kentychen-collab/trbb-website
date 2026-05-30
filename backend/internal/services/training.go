package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"trbb/internal/models"
	"trbb/pkg/database"
	"trbb/pkg/storage"
)

var (
	ErrTrainingNotFound = errors.New("training log not found")
	ErrNotOwner         = errors.New("not the owner of this log")
)

type TrainingService struct {
	db    *database.DB
	store *storage.Storage
}

func NewTrainingService(db *database.DB, store *storage.Storage) *TrainingService {
	return &TrainingService{db: db, store: store}
}

// ── GPX parser ────────────────────────────────────────────────

type gpxDoc struct {
	XMLName xml.Name  `xml:"gpx"`
	Tracks  []gpxTrk `xml:"trk"`
}
type gpxTrk struct {
	Segs []gpxTrkSeg `xml:"trkseg"`
}
type gpxTrkSeg struct {
	Points []gpxTrkPt `xml:"trkpt"`
}
type gpxTrkPt struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Ele  float64 `xml:"ele"`
	Time string  `xml:"time"`
	Extensions struct {
		HR    int `xml:"TrackPointExtension>hr"`
		Cad   int `xml:"TrackPointExtension>cad"`
		Power int `xml:"TrackPointExtension>power"`
	} `xml:"extensions"`
}

type ParsedGPX struct {
	Points      models.RoutePoints
	StartLat    float64
	StartLng    float64
	StartDate   string  // YYYY-MM-DD extracted from first trackpoint
	DistanceKm  float64
	ElevationM  int
	DurationMin int
	AvgHR       int
	AvgPace     string // mm:ss/km
}

func ParseGPX(data []byte) (*ParsedGPX, error) {
	var doc gpxDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse gpx: %w", err)
	}

	var points []gpxTrkPt
	for _, trk := range doc.Tracks {
		for _, seg := range trk.Segs {
			points = append(points, seg.Points...)
		}
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("no track points found in GPX")
	}

	result := &ParsedGPX{
		StartLat:  points[0].Lat,
		StartLng:  points[0].Lon,
		StartDate: time.Now().Format("2006-01-02"),
	}

	// Extract date from first trackpoint
	if t1 := parseGPXTime(points[0].Time); !t1.IsZero() {
		result.StartDate = t1.Format("2006-01-02")
	}

	// Route points (sample every N points to keep JSON small)
	step := 1
	if len(points) > 500 {
		step = len(points) / 500
	}
	for i := 0; i < len(points); i += step {
		result.Points = append(result.Points, models.LatLng{points[i].Lat, points[i].Lon})
	}
	// Always include last point
	last := points[len(points)-1]
	result.Points = append(result.Points, models.LatLng{last.Lat, last.Lon})

	// Distance (haversine)
	var totalDist float64
	for i := 1; i < len(points); i++ {
		totalDist += haversine(points[i-1].Lat, points[i-1].Lon, points[i].Lat, points[i].Lon)
	}
	result.DistanceKm = math.Round(totalDist*1000) / 1000

	// Elevation gain
	var elevGain float64
	for i := 1; i < len(points); i++ {
		diff := points[i].Ele - points[i-1].Ele
		if diff > 0 {
			elevGain += diff
		}
	}
	result.ElevationM = int(elevGain)

	// Duration
	if len(points) >= 2 {
		t1 := parseGPXTime(points[0].Time)
		t2 := parseGPXTime(points[len(points)-1].Time)
		if !t1.IsZero() && !t2.IsZero() {
			result.DurationMin = int(t2.Sub(t1).Minutes())
		}
	}

	// Avg pace (min/km)
	if result.DistanceKm > 0 && result.DurationMin > 0 {
		secsPerKm := float64(result.DurationMin*60) / result.DistanceKm
		mins := int(secsPerKm) / 60
		secs := int(secsPerKm) % 60
		result.AvgPace = fmt.Sprintf("%d:%02d", mins, secs)
	}

	// Avg HR
	var hrSum, hrCount int
	for _, p := range points {
		if p.Extensions.HR > 0 {
			hrSum += p.Extensions.HR
			hrCount++
		}
	}
	if hrCount > 0 {
		result.AvgHR = hrSum / hrCount
	}

	return result, nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func parseGPXTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// ── Create / Upload ───────────────────────────────────────────

type TrainingInput struct {
	Title        string   `json:"title"       form:"title"     binding:"required"`
	SportType    int      `json:"sport_type"  form:"sport_type"`
	Date         string   `json:"date"        form:"date"      binding:"required"`
	DurationMin  *int     `json:"duration_min"`
	DistanceKm   *float64 `json:"distance_km"`
	AvgHeartRate *int     `json:"avg_heart_rate"`
	MaxHeartRate *int     `json:"max_heart_rate"`
	Calories     *int     `json:"calories"`
	ElevationM   *int     `json:"elevation_m"`
	AvgPace      string   `json:"avg_pace"`
	AvgSpeedKph  *float64 `json:"avg_speed_kph"`
	PowerAvg     *int     `json:"power_avg"`
	Note         string   `json:"note"`
	IsPublic     bool     `json:"is_public"`
	Photos       []string `json:"photos"`
}

func (s *TrainingService) Create(ctx context.Context, userID uint64, in TrainingInput) (*models.TrainingLog, error) {
	uid := uuid.New().String()
	now := time.Now()
	photosJSON, _ := json.Marshal(in.Photos)

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO training_logs
		  (uuid,user_id,title,sport_type,date,
		   duration_min,distance_km,avg_heart_rate,max_heart_rate,
		   calories,elevation_m,avg_pace,avg_speed_kph,power_avg,
		   note,is_public,photos,source,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'manual',?,?)`,
		uid, userID, in.Title, in.SportType, in.Date,
		in.DurationMin, in.DistanceKm, in.AvgHeartRate, in.MaxHeartRate,
		in.Calories, in.ElevationM, nullStr(in.AvgPace), in.AvgSpeedKph, in.PowerAvg,
		nullStr(in.Note), boolInt(in.IsPublic), string(photosJSON),
		now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create training: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetByID(ctx, uint64(id), userID)
}

// UploadGPX uploads GPX file, parses it, and creates/updates a log
func (s *TrainingService) UploadGPX(ctx context.Context, userID, logID uint64, filename string, data []byte) (*models.TrainingLog, error) {
	parsed, err := ParseGPX(data)
	if err != nil {
		return nil, fmt.Errorf("invalid GPX: %w", err)
	}

	// Upload raw GPX to MinIO
	objectPath, err := s.store.UploadImage(ctx,
		"training", filename, "application/gpx+xml",
		bytes.NewReader(data), int64(len(data)),
	)
	if err != nil {
		return nil, fmt.Errorf("store gpx: %w", err)
	}

	routeJSON, _ := json.Marshal(parsed.Points)
	now := time.Now()

	if logID > 0 {
		// Update existing log
		_, err = s.db.ExecContext(ctx, `
			UPDATE training_logs SET
			  gpx_file_path=?,route_points=?,start_lat=?,start_lng=?,
			  source='gpx',updated_at=?
			WHERE id=? AND user_id=?`,
			objectPath, string(routeJSON), parsed.StartLat, parsed.StartLng,
			now, logID, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("update gpx: %w", err)
		}
		return s.GetByID(ctx, logID, userID)
	}

	// Create new log from GPX data
	sportType := 1 // default Run
	uid := uuid.New().String()

	// Use first point's time as date, fallback to today
	gpxDate := parsed.StartDate
	if gpxDate == "" {
		gpxDate = time.Now().Format("2006-01-02")
	}

	var durMin interface{} = nil
	if parsed.DurationMin > 0 { durMin = parsed.DurationMin }
	var distKm interface{} = nil
	if parsed.DistanceKm > 0 { distKm = parsed.DistanceKm }
	var elevM interface{} = nil
	if parsed.ElevationM > 0 { elevM = parsed.ElevationM }
	var avgPace interface{} = nil
	if parsed.AvgPace != "" { avgPace = parsed.AvgPace }
	var avgHR interface{} = nil
	if parsed.AvgHR > 0 { avgHR = parsed.AvgHR }
	var startLatV, startLngV interface{} = nil, nil
	if parsed.StartLat != 0 { startLatV = parsed.StartLat }
	if parsed.StartLng != 0 { startLngV = parsed.StartLng }

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO training_logs
		  (uuid,user_id,title,sport_type,date,
		   duration_min,distance_km,elevation_m,avg_pace,avg_heart_rate,
		   gpx_file_path,route_points,start_lat,start_lng,
		   is_public,photos,source,created_at,updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'[]','gpx',?,?)`,
		uid, userID, filename, sportType, gpxDate,
		durMin, distKm, elevM, avgPace, avgHR,
		objectPath, string(routeJSON), startLatV, startLngV,
		0, // is_public: caller can patch after upload
		now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create from gpx: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetByID(ctx, uint64(id), userID)
}

// UploadFIT stores the FIT file (parsing FIT binary requires external lib, store as-is)
func (s *TrainingService) UploadFIT(ctx context.Context, userID, logID uint64, filename string, data []byte) (string, error) {
	objectPath, err := s.store.UploadImage(ctx,
		"training", filename, "application/octet-stream",
		bytes.NewReader(data), int64(len(data)),
	)
	if err != nil {
		return "", fmt.Errorf("store fit: %w", err)
	}
	if logID > 0 {
		_, _ = s.db.ExecContext(ctx,
			"UPDATE training_logs SET fit_file_path=?,source='fit',updated_at=? WHERE id=? AND user_id=?",
			objectPath, time.Now(), logID, userID,
		)
	}
	return objectPath, nil
}

// ── PatchPublic — only update is_public ──────────────────────

func (s *TrainingService) PatchPublic(ctx context.Context, id, userID uint64, isPublic bool) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE training_logs SET is_public=?,updated_at=NOW() WHERE id=? AND user_id=?",
		isPublic, id, userID,
	)
	return err
}

// ── Update ────────────────────────────────────────────────────

func (s *TrainingService) Update(ctx context.Context, id, userID uint64, in TrainingInput) (*models.TrainingLog, error) {
	photosJSON, _ := json.Marshal(in.Photos)
	_, err := s.db.ExecContext(ctx, `
		UPDATE training_logs SET
		  title=?,sport_type=?,date=?,
		  duration_min=?,distance_km=?,avg_heart_rate=?,max_heart_rate=?,
		  calories=?,elevation_m=?,avg_pace=?,avg_speed_kph=?,power_avg=?,
		  note=?,is_public=?,photos=?,updated_at=?
		WHERE id=? AND user_id=?`,
		in.Title, in.SportType, in.Date,
		in.DurationMin, in.DistanceKm, in.AvgHeartRate, in.MaxHeartRate,
		in.Calories, in.ElevationM, nullStr(in.AvgPace), in.AvgSpeedKph, in.PowerAvg,
		nullStr(in.Note), boolInt(in.IsPublic), string(photosJSON),
		time.Now(), id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("update training: %w", err)
	}
	return s.GetByID(ctx, id, userID)
}

func (s *TrainingService) Delete(ctx context.Context, id, userID uint64) error {
	res, err := s.db.ExecContext(ctx,
		"DELETE FROM training_logs WHERE id=? AND user_id=?", id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrTrainingNotFound
	}
	return nil
}

// ── Query ─────────────────────────────────────────────────────

const trainingSelectSQL = `
	SELECT t.id,t.uuid,t.user_id,t.title,t.sport_type,t.date,
	       t.duration_min,t.distance_km,t.avg_heart_rate,t.max_heart_rate,
	       t.calories,t.elevation_m,COALESCE(t.avg_pace,''),t.avg_speed_kph,t.power_avg,
	       COALESCE(t.note,''),t.is_public,COALESCE(t.photos,'[]'),
	       COALESCE(t.gpx_file_path,''),COALESCE(t.fit_file_path,''),
	       t.route_points,t.start_lat,t.start_lng,
	       COALESCE(t.source,'manual'),COALESCE(t.garmin_activity_id,''),
	       t.created_at,t.updated_at,
	       COALESCE(u.username,''),COALESCE(u.display_name,''),COALESCE(u.avatar_url,'')
	FROM training_logs t
	JOIN users u ON u.id=t.user_id `

func (s *TrainingService) GetByID(ctx context.Context, id, viewerID uint64) (*models.TrainingLog, error) {
	row := s.db.QueryRowContext(ctx, trainingSelectSQL+"WHERE t.id=?", id)
	log, err := scanTraining(row)
	if err != nil {
		return nil, err
	}
	// access control: private logs only visible to owner
	if !log.IsPublic && log.UserID != viewerID {
		return nil, ErrTrainingNotFound
	}
	return log, nil
}

func (s *TrainingService) GetByUUID(ctx context.Context, uid string, viewerID uint64) (*models.TrainingLog, error) {
	row := s.db.QueryRowContext(ctx, trainingSelectSQL+"WHERE t.uuid=?", uid)
	log, err := scanTraining(row)
	if err != nil {
		return nil, err
	}
	if !log.IsPublic && log.UserID != viewerID {
		return nil, ErrTrainingNotFound
	}
	return log, nil
}

type ListTrainingInput struct {
	UserID    *uint64 `form:"-"`
	Public    bool    `form:"public"`
	SportType *int    `form:"sport_type"`
	DateFrom  string  `form:"date_from"` // YYYY-MM-DD
	DateTo    string  `form:"date_to"`   // YYYY-MM-DD
	Page      int     `form:"page"`
	PageSize  int     `form:"page_size"`
}

type ListTrainingResult struct {
	Logs  []*models.TrainingLog `json:"logs"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Pages int                   `json:"pages"`
}

func (s *TrainingService) List(ctx context.Context, in ListTrainingInput, viewerID uint64) (*ListTrainingResult, error) {
	if in.Page < 1 {
		in.Page = 1
	}
	if in.PageSize < 1 || in.PageSize > 50 {
		in.PageSize = 20
	}

	where := "WHERE 1=1"
	args := []any{}
	if in.UserID != nil {
		where += " AND t.user_id=?"
		args = append(args, *in.UserID)
		// only show private logs to owner
		if *in.UserID != viewerID {
			where += " AND t.is_public=1"
		}
	} else if in.Public {
		where += " AND t.is_public=1"
	}
	if in.SportType != nil {
		where += " AND t.sport_type=?"
		args = append(args, *in.SportType)
	}
	if in.DateFrom != "" {
		where += " AND t.date >= ?"
		args = append(args, in.DateFrom)
	}
	if in.DateTo != "" {
		where += " AND t.date <= ?"
		args = append(args, in.DateTo)
	}

	var total int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM training_logs t JOIN users u ON u.id=t.user_id "+where, args...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}

	offset := (in.Page - 1) * in.PageSize
	query := trainingSelectSQL + where + " ORDER BY t.date DESC, t.created_at DESC LIMIT ? OFFSET ?"
	rows, err := s.db.QueryContext(ctx, query, append(args, in.PageSize, offset)...)
	if err != nil {
		return nil, fmt.Errorf("list training: %w", err)
	}
	defer rows.Close()

	var logs []*models.TrainingLog
	for rows.Next() {
		log, err := scanTrainingRow(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	pages := (total + in.PageSize - 1) / in.PageSize
	if pages < 1 {
		pages = 1
	}
	return &ListTrainingResult{Logs: logs, Total: total, Page: in.Page, Pages: pages}, nil
}

// ── Garmin OAuth ──────────────────────────────────────────────

func (s *TrainingService) SaveGarminToken(ctx context.Context, userID uint64, accessToken, tokenSecret, garminUserID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO garmin_tokens (user_id,access_token,token_secret,garmin_user_id,created_at,updated_at)
		VALUES (?,?,?,?,NOW(),NOW())
		ON DUPLICATE KEY UPDATE
		  access_token=VALUES(access_token),
		  token_secret=VALUES(token_secret),
		  garmin_user_id=VALUES(garmin_user_id),
		  updated_at=NOW()`,
		userID, accessToken, tokenSecret, garminUserID,
	)
	return err
}

func (s *TrainingService) GetGarminToken(ctx context.Context, userID uint64) (*models.GarminToken, error) {
	t := &models.GarminToken{}
	var lastSync sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT id,user_id,access_token,token_secret,
		       COALESCE(garmin_user_id,''),COALESCE(scope,''),
		       last_sync_at,created_at,updated_at
		FROM garmin_tokens WHERE user_id=?`, userID,
	).Scan(&t.ID, &t.UserID, &t.AccessToken, &t.TokenSecret,
		&t.GarminUserID, &t.Scope, &lastSync, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if lastSync.Valid {
		t.LastSyncAt = &lastSync.Time
	}
	return t, nil
}

func (s *TrainingService) DeleteGarminToken(ctx context.Context, userID uint64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM garmin_tokens WHERE user_id=?", userID)
	return err
}

// ── Scan helpers ──────────────────────────────────────────────

func scanTraining(row *sql.Row) (*models.TrainingLog, error) {
	t := &models.TrainingLog{}
	var photosJSON sql.NullString
	var routeJSON  sql.NullString
	var avgPace    sql.NullString
	var garminID   sql.NullString
	var startLat, startLng sql.NullFloat64
	var dur, avgHR, maxHR, cal, elev, powerAvg sql.NullInt64
	var distKm, avgSpeedKph sql.NullFloat64

	err := row.Scan(
		&t.ID, &t.UUID, &t.UserID, &t.Title, &t.SportType, &t.Date,
		&dur, &distKm, &avgHR, &maxHR,
		&cal, &elev, &avgPace, &avgSpeedKph, &powerAvg,
		&t.Note, &t.IsPublic, &photosJSON,
		&t.GpxFilePath, &t.FitFilePath,
		&routeJSON, &startLat, &startLng,
		&t.Source, &garminID,
		&t.CreatedAt, &t.UpdatedAt,
		&t.Username, &t.DisplayName, &t.AvatarURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTrainingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan training: %w", err)
	}
	applyNullInts(t, dur, avgHR, maxHR, cal, elev, powerAvg)
	if distKm.Valid    { t.DistanceKm = &distKm.Float64 }
	if avgSpeedKph.Valid { t.AvgSpeedKph = &avgSpeedKph.Float64 }
	if startLat.Valid  { t.StartLat = &startLat.Float64 }
	if startLng.Valid  { t.StartLng = &startLng.Float64 }
	t.AvgPace           = avgPace.String
	t.GarminActivityID  = garminID.String
	if photosJSON.Valid && photosJSON.String != "" {
		_ = json.Unmarshal([]byte(photosJSON.String), &t.Photos)
	}
	if routeJSON.Valid && routeJSON.String != "" && routeJSON.String != "null" {
		_ = json.Unmarshal([]byte(routeJSON.String), &t.RoutePoints)
	}
	return t, nil
}

func scanTrainingRow(rows *sql.Rows) (*models.TrainingLog, error) {
	t := &models.TrainingLog{}
	var photosJSON sql.NullString
	var routeJSON  sql.NullString
	var avgPace    sql.NullString
	var garminID   sql.NullString
	var startLat, startLng sql.NullFloat64
	var dur, avgHR, maxHR, cal, elev, powerAvg sql.NullInt64
	var distKm, avgSpeedKph sql.NullFloat64

	err := rows.Scan(
		&t.ID, &t.UUID, &t.UserID, &t.Title, &t.SportType, &t.Date,
		&dur, &distKm, &avgHR, &maxHR,
		&cal, &elev, &avgPace, &avgSpeedKph, &powerAvg,
		&t.Note, &t.IsPublic, &photosJSON,
		&t.GpxFilePath, &t.FitFilePath,
		&routeJSON, &startLat, &startLng,
		&t.Source, &garminID,
		&t.CreatedAt, &t.UpdatedAt,
		&t.Username, &t.DisplayName, &t.AvatarURL,
	)
	if err != nil {
		return nil, fmt.Errorf("scan training row: %w", err)
	}
	applyNullInts(t, dur, avgHR, maxHR, cal, elev, powerAvg)
	if distKm.Valid      { t.DistanceKm = &distKm.Float64 }
	if avgSpeedKph.Valid { t.AvgSpeedKph = &avgSpeedKph.Float64 }
	if startLat.Valid    { t.StartLat = &startLat.Float64 }
	if startLng.Valid    { t.StartLng = &startLng.Float64 }
	t.AvgPace          = avgPace.String
	t.GarminActivityID = garminID.String
	if photosJSON.Valid && photosJSON.String != "" {
		_ = json.Unmarshal([]byte(photosJSON.String), &t.Photos)
	}
	if routeJSON.Valid && routeJSON.String != "" && routeJSON.String != "null" {
		_ = json.Unmarshal([]byte(routeJSON.String), &t.RoutePoints)
	}
	return t, nil
}

func applyNullInts(t *models.TrainingLog, dur, avgHR, maxHR, cal, elev, power sql.NullInt64) {
	if dur.Valid   { v := int(dur.Int64);   t.DurationMin = &v }
	if avgHR.Valid { v := int(avgHR.Int64); t.AvgHeartRate = &v }
	if maxHR.Valid { v := int(maxHR.Int64); t.MaxHeartRate = &v }
	if cal.Valid   { v := int(cal.Int64);   t.Calories = &v }
	if elev.Valid  { v := int(elev.Int64);  t.ElevationM = &v }
	if power.Valid { v := int(power.Int64); t.PowerAvg = &v }
}

func boolInt(b bool) int {
	if b { return 1 }
	return 0
}

// ── Garmin Activity sync (framework) ─────────────────────────

// SyncGarminActivities pulls latest activities from Garmin Health API.
// Requires valid OAuth1 token. Currently a placeholder — fill in when
// GARMIN_CLIENT_ID / GARMIN_CLIENT_SECRET are available.
func (s *TrainingService) SyncGarminActivities(ctx context.Context, userID uint64) (int, error) {
	token, err := s.GetGarminToken(ctx, userID)
	if err != nil || token == nil {
		return 0, fmt.Errorf("no garmin token — please connect your Garmin account first")
	}
	// TODO: implement Garmin Health API call using OAuth1
	// Endpoint: GET {GARMIN_API_BASE}/activities/activityFiles
	// Sign request with token.AccessToken + token.TokenSecret
	// Parse response and call s.CreateFromGarminActivity(...)
	return 0, fmt.Errorf("garmin sync not yet configured — waiting for API credentials")
}

// FitFileSummary extracts basic info from FIT file name for display
// Full FIT binary parsing requires github.com/tormoder/fit or similar
func FitFileSummary(filename string) string {
	name := strings.TrimSuffix(filename, ".fit")
	parts := strings.Split(name, "_")
	if len(parts) >= 2 {
		return strings.Join(parts[:2], " ")
	}
	return filename
}

// Unused helpers to keep strconv import
var _ = strconv.Itoa

// ── Admin: list all logs (bypass is_public) ───────────────────

type AdminListTrainingInput struct {
	UserID    *uint64 `form:"user_id"`
	SportType *int    `form:"sport_type"`
	Keyword   string  `form:"keyword"`
	DateFrom  string  `form:"date_from"` // YYYY-MM-DD
	DateTo    string  `form:"date_to"`   // YYYY-MM-DD
	Page      int     `form:"page"`
	PageSize  int     `form:"page_size"`
}

func (s *TrainingService) AdminList(ctx context.Context, in AdminListTrainingInput) (*ListTrainingResult, error) {
	if in.Page < 1 { in.Page = 1 }
	if in.PageSize < 1 || in.PageSize > 100 { in.PageSize = 20 }

	where := "WHERE 1=1"
	args := []any{}
	if in.UserID != nil {
		where += " AND t.user_id=?"
		args = append(args, *in.UserID)
	}
	if in.SportType != nil {
		where += " AND t.sport_type=?"
		args = append(args, *in.SportType)
	}
	if in.Keyword != "" {
		where += " AND (t.title LIKE ? OR u.username LIKE ? OR u.display_name LIKE ? OR u.name_zh LIKE ?)"
		kw := "%" + in.Keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}
	if in.DateFrom != "" {
		where += " AND t.date >= ?"
		args = append(args, in.DateFrom)
	}
	if in.DateTo != "" {
		where += " AND t.date <= ?"
		args = append(args, in.DateTo)
	}

	var total int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM training_logs t JOIN users u ON u.id=t.user_id "+where, args...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count admin training: %w", err)
	}

	offset := (in.Page - 1) * in.PageSize
	rows, err := s.db.QueryContext(ctx,
		trainingSelectSQL+where+" ORDER BY t.date DESC, t.created_at DESC LIMIT ? OFFSET ?",
		append(args, in.PageSize, offset)...,
	)
	if err != nil { return nil, fmt.Errorf("admin list training: %w", err) }
	defer rows.Close()

	var logs []*models.TrainingLog
	for rows.Next() {
		lg, err := scanTrainingRow(rows)
		if err != nil { return nil, err }
		logs = append(logs, lg)
	}
	pages := (total + in.PageSize - 1) / in.PageSize
	if pages < 1 { pages = 1 }
	return &ListTrainingResult{Logs: logs, Total: total, Page: in.Page, Pages: pages}, nil
}

// ── Admin: statistics per user ────────────────────────────────

type UserTrainingStat struct {
	UserID      uint64  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	TotalLogs   int     `json:"total_logs"`
	TotalKm     float64 `json:"total_km"`
	TotalMin    int     `json:"total_min"`
	LastDate    string  `json:"last_date"`
}

func (s *TrainingService) AdminStats(ctx context.Context) ([]*UserTrainingStat, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.username, COALESCE(u.display_name,''),
		       COUNT(t.id),
		       COALESCE(SUM(t.distance_km),0),
		       COALESCE(SUM(t.duration_min),0),
		       COALESCE(MAX(t.date),'')
		FROM users u
		JOIN training_logs t ON t.user_id=u.id
		WHERE u.deleted_at IS NULL
		GROUP BY u.id
		ORDER BY COUNT(t.id) DESC
		LIMIT 100`)
	if err != nil { return nil, fmt.Errorf("admin stats: %w", err) }
	defer rows.Close()

	var stats []*UserTrainingStat
	for rows.Next() {
		s2 := &UserTrainingStat{}
		if err := rows.Scan(
			&s2.UserID, &s2.Username, &s2.DisplayName,
			&s2.TotalLogs, &s2.TotalKm, &s2.TotalMin, &s2.LastDate,
		); err != nil { return nil, err }
		stats = append(stats, s2)
	}
	return stats, nil
}

// ── Strava OAuth2 ─────────────────────────────────────────────

func (s *TrainingService) SaveStravaToken(ctx context.Context,
	userID uint64, accessToken, refreshToken, tokenType string,
	expiresAt *time.Time, athleteID int64, athleteName string,
) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO strava_tokens
		  (user_id, access_token, refresh_token, token_type, expires_at,
		   strava_athlete_id, athlete_name, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
		  access_token=VALUES(access_token),
		  refresh_token=VALUES(refresh_token),
		  token_type=VALUES(token_type),
		  expires_at=VALUES(expires_at),
		  strava_athlete_id=VALUES(strava_athlete_id),
		  athlete_name=VALUES(athlete_name),
		  updated_at=VALUES(updated_at)`,
		userID, accessToken, refreshToken, tokenType, expiresAt,
		athleteID, athleteName, now, now,
	)
	return err
}

func (s *TrainingService) GetStravaToken(ctx context.Context, userID uint64) (*models.StravaToken, error) {
	t := &models.StravaToken{}
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, access_token, refresh_token, token_type,
		       expires_at, strava_athlete_id, COALESCE(athlete_name,''),
		       last_sync_at, created_at, updated_at
		FROM strava_tokens WHERE user_id=?`, userID,
	).Scan(
		&t.ID, &t.UserID, &t.AccessToken, &t.RefreshToken, &t.TokenType,
		&t.ExpiresAt, &t.StravaAthleteID, &t.AthleteName,
		&t.LastSyncAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func (s *TrainingService) DeleteStravaToken(ctx context.Context, userID uint64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM strava_tokens WHERE user_id=?", userID)
	return err
}

// SyncStravaActivities pulls activities from the last 30 days from Strava API.
// Returns number of new activities synced.
func (s *TrainingService) SyncStravaActivities(ctx context.Context, userID uint64) (int, error) {
	token, err := s.GetStravaToken(ctx, userID)
	if err != nil || token == nil {
		return 0, fmt.Errorf("Strava 尚未連結，請先授權")
	}

	// Refresh token if expired
	accessToken := token.AccessToken
	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		newToken, refreshErr := s.refreshStravaToken(ctx, userID, token.RefreshToken)
		if refreshErr != nil {
			return 0, fmt.Errorf("Strava token 已過期，請重新連結")
		}
		accessToken = newToken
	}

	// Fetch activities from last 30 days
	after := time.Now().AddDate(0, 0, -30).Unix()
	apiURL := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?after=%d&per_page=50", after)

	req, _ := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Strava API 請求失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return 0, fmt.Errorf("Strava 授權已失效，請重新連結")
	}
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("Strava API 回傳錯誤: %d", resp.StatusCode)
	}

	var activities []struct {
		ID              int64   `json:"id"`
		Name            string  `json:"name"`
		Type            string  `json:"type"`
		SportType       string  `json:"sport_type"`
		StartDateLocal  string  `json:"start_date_local"`
		MovingTime      int     `json:"moving_time"`
		Distance        float64 `json:"distance"`
		AverageHeartrate float64 `json:"average_heartrate"`
		MaxHeartrate     float64 `json:"max_heartrate"`
		TotalElevGain    float64 `json:"total_elevation_gain"`
		Calories         float64 `json:"calories"`
		Map struct {
			SummaryPolyline string `json:"summary_polyline"`
		} `json:"map"`
	}

	body := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	for {
		n, readErr := resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		if readErr != nil {
			break
		}
	}
	if err := json.Unmarshal(body, &activities); err != nil {
		return 0, fmt.Errorf("解析 Strava 資料失敗")
	}

	// Map Strava sport types to our sport_type integers
	sportMap := map[string]int{
		"Run": 1, "TrailRun": 1, "VirtualRun": 1,
		"Swim": 2, "OpenWaterSwim": 2,
		"Ride": 3, "MountainBikeRide": 3, "VirtualRide": 3, "EBikeRide": 3,
		"Triathlon": 4,
		"WeightTraining": 5, "Crossfit": 5, "Workout": 5,
	}

	count := 0
	for _, act := range activities {
		// Check if already imported (by strava activity ID stored in garmin_activity_id field)
		stravaID := fmt.Sprintf("strava_%d", act.ID)
		var existing int
		s.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM training_logs WHERE user_id=? AND garmin_activity_id=?",
			userID, stravaID,
		).Scan(&existing)
		if existing > 0 {
			continue
		}

		sportType := 6 // other
		if t, ok := sportMap[act.SportType]; ok {
			sportType = t
		} else if t, ok := sportMap[act.Type]; ok {
			sportType = t
		}

		date := ""
		if len(act.StartDateLocal) >= 10 {
			date = act.StartDateLocal[:10]
		}

		durationMin := act.MovingTime / 60
		distKm := act.Distance / 1000.0
		avgHR := int(act.AverageHeartrate)
		maxHR := int(act.MaxHeartrate)
		elevM := int(act.TotalElevGain)
		cal := int(act.Calories)

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO training_logs
			  (uuid, user_id, title, sport_type, date, duration_min, distance_km,
			   avg_heart_rate, max_heart_rate, elevation_m, calories,
			   source, garmin_activity_id, is_public, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,'strava',?,0,NOW(),NOW())`,
			uuid.New().String(),
			userID, act.Name, sportType, date, nullInt(durationMin),
			nullFloat(distKm), nullInt(avgHR), nullInt(maxHR),
			nullInt(elevM), nullInt(cal), stravaID,
		)
		if err == nil {
			count++
		}
	}

	// Update last_sync_at
	s.db.ExecContext(ctx,
		"UPDATE strava_tokens SET last_sync_at=? WHERE user_id=?",
		time.Now(), userID,
	)

	return count, nil
}

// refreshStravaToken exchanges refresh token for new access token
func (s *TrainingService) refreshStravaToken(ctx context.Context, userID uint64, refreshToken string) (string, error) {
	// Requires config — passed via env; use hardcoded env read as fallback
	clientID := getEnvVal("STRAVA_CLIENT_ID")
	clientSecret := getEnvVal("STRAVA_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("Strava credentials not configured")
	}

	formData := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}
	resp, err := http.PostForm("https://www.strava.com/oauth/token", formData)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresAt    int64  `json:"expires_at"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.AccessToken == "" {
		return "", fmt.Errorf("refresh failed")
	}

	expiresAt := time.Unix(result.ExpiresAt, 0)
	token, _ := s.GetStravaToken(ctx, userID)
	athleteName := ""
	athletreID := int64(0)
	if token != nil {
		athleteName = token.AthleteName
		athletreID = token.StravaAthleteID
	}
	s.SaveStravaToken(ctx, userID, result.AccessToken, result.RefreshToken, "Bearer", &expiresAt, athletreID, athleteName)
	return result.AccessToken, nil
}

func getEnvVal(key string) string { return os.Getenv(key) }
func nullInt(v int) interface{}       { if v <= 0 { return nil }; return v }
func nullFloat(v float64) interface{} { if v <= 0 { return nil }; return v }
