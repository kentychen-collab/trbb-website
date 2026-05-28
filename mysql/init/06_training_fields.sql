-- ═══════════════════════════════════════════════════════════
-- Migration: 06_training_fields.sql
-- 擴充 training_logs 加入 GPX/FIT/地圖路徑欄位
-- 新增 garmin_tokens 表儲存 OAuth token
-- ═══════════════════════════════════════════════════════════

DROP PROCEDURE IF EXISTS add_col_if_not_exists;
DELIMITER $$
CREATE PROCEDURE add_col_if_not_exists(IN tbl VARCHAR(64), IN col VARCHAR(64), IN def TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=tbl AND COLUMN_NAME=col
  ) THEN
    SET @sql = CONCAT('ALTER TABLE `', tbl, '` ADD COLUMN `', col, '` ', def);
    PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
  END IF;
END$$
DELIMITER ;

-- GPX 檔案路徑（存在 MinIO）
CALL add_col_if_not_exists('training_logs', 'gpx_file_path',
  "VARCHAR(500) DEFAULT NULL COMMENT 'MinIO object path'");

-- FIT 檔案路徑（存在 MinIO）
CALL add_col_if_not_exists('training_logs', 'fit_file_path',
  "VARCHAR(500) DEFAULT NULL COMMENT 'MinIO object path'");

-- 地圖路線座標（解析自 GPX，存為 JSON array of [lat,lng]）
CALL add_col_if_not_exists('training_logs', 'route_points',
  "JSON DEFAULT NULL COMMENT '[[lat,lng],...]'");

-- 起點座標
CALL add_col_if_not_exists('training_logs', 'start_lat',
  "DECIMAL(10,7) DEFAULT NULL");
CALL add_col_if_not_exists('training_logs', 'start_lng',
  "DECIMAL(10,7) DEFAULT NULL");

-- 資料來源
CALL add_col_if_not_exists('training_logs', 'source',
  "VARCHAR(20) DEFAULT 'manual' COMMENT 'manual|gpx|fit|garmin'");

DROP PROCEDURE IF EXISTS add_col_if_not_exists;

-- ── Garmin OAuth tokens ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS `garmin_tokens` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`       BIGINT UNSIGNED NOT NULL UNIQUE,
  `access_token`  TEXT            NOT NULL,
  `token_secret`  TEXT            NOT NULL  COMMENT 'OAuth1 secret',
  `garmin_user_id` VARCHAR(100)   DEFAULT NULL,
  `scope`         VARCHAR(200)    DEFAULT NULL,
  `last_sync_at`  DATETIME        DEFAULT NULL,
  `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
