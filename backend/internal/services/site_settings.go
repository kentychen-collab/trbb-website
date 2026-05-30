package services

import (
	"context"
	"fmt"
	"time"

	"trbb/pkg/database"
)

type SiteSetting struct {
	Key       string  `json:"key"`
	Value     *string `json:"value"`
	Type      string  `json:"type"`
	Group     string  `json:"group"`
	Label     string  `json:"label"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SiteSettingsService struct {
	db *database.DB
}

func NewSiteSettingsService(db *database.DB) *SiteSettingsService {
	return &SiteSettingsService{db: db}
}

// GetAll — 回傳所有設定（key→value map，供前台直接使用）
func (s *SiteSettingsService) GetAll(ctx context.Context) (map[string]any, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT `key`, `value`, `type` FROM site_settings ORDER BY `group`, `id`")
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}
	defer rows.Close()

	result := map[string]any{}
	for rows.Next() {
		var key, typ string
		var val *string
		if err := rows.Scan(&key, &val, &typ); err != nil {
			return nil, err
		}
		if val == nil {
			result[key] = nil
		} else {
			result[key] = *val
		}
	}
	return result, nil
}

// GetGrouped — 後台管理用，依 group 分類回傳完整設定
func (s *SiteSettingsService) GetGrouped(ctx context.Context) (map[string][]*SiteSetting, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ` + "`key`" + `, COALESCE(` + "`value`" + `,''), ` + "`type`" + `, ` + "`group`" + `,
		       COALESCE(label,''), updated_at
		FROM site_settings ORDER BY ` + "`group`" + `, id`)
	if err != nil {
		return nil, fmt.Errorf("get grouped settings: %w", err)
	}
	defer rows.Close()

	result := map[string][]*SiteSetting{}
	for rows.Next() {
		st := &SiteSetting{}
		var val string
		if err := rows.Scan(&st.Key, &val, &st.Type, &st.Group, &st.Label, &st.UpdatedAt); err != nil {
			return nil, err
		}
		st.Value = &val
		result[st.Group] = append(result[st.Group], st)
	}
	return result, nil
}

// Set — 更新單一設定
func (s *SiteSettingsService) Set(ctx context.Context, key, value string, updaterID uint64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE site_settings SET ` + "`value`" + `=?, updated_by=?, updated_at=?
		WHERE ` + "`key`" + `=?`,
		value, updaterID, time.Now(), key,
	)
	return err
}

// BatchSet — 批次更新多個設定
func (s *SiteSettingsService) BatchSet(ctx context.Context, settings map[string]string, updaterID uint64) error {
	now := time.Now()
	for key, value := range settings {
		_, err := s.db.ExecContext(ctx, `
			UPDATE site_settings SET ` + "`value`" + `=?, updated_by=?, updated_at=?
			WHERE ` + "`key`" + `=?`,
			value, updaterID, now, key,
		)
		if err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
	}
	return nil
}
