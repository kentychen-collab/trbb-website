-- ═══════════════════════════════════════════════════════════
-- Migration: 08_training_extended.sql
-- 擴充訓練記錄欄位 + 自動同步公開設定
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

-- ── training_logs 新增欄位 ────────────────────────────────

-- 總下降高度（m）
CALL add_col_if_not_exists('training_logs', 'descent_m',
  "INT DEFAULT NULL COMMENT '總下降高度 m'");

-- 活動開始時間（含時分秒）
CALL add_col_if_not_exists('training_logs', 'start_time',
  "DATETIME DEFAULT NULL COMMENT '活動開始時間'");

-- GPX 路線縮圖（存在 MinIO，作為預設封面）
CALL add_col_if_not_exists('training_logs', 'map_thumbnail_url',
  "VARCHAR(500) DEFAULT NULL COMMENT 'GPX路線縮圖URL'");

-- 封面圖（優先顯示，預設使用 map_thumbnail_url）
CALL add_col_if_not_exists('training_logs', 'cover_url',
  "VARCHAR(500) DEFAULT NULL COMMENT '封面圖URL'");

-- ── strava_tokens：記憶公開設定 ──────────────────────────
CALL add_col_if_not_exists('strava_tokens', 'sync_public',
  "TINYINT(1) NOT NULL DEFAULT 0 COMMENT '自動同步後是否公開'");
CALL add_col_if_not_exists('strava_tokens', 'auto_sync',
  "TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否啟用 Webhook 自動同步'");

-- ── garmin_tokens：記憶公開設定 ──────────────────────────
CALL add_col_if_not_exists('garmin_tokens', 'sync_public',
  "TINYINT(1) NOT NULL DEFAULT 0 COMMENT '自動同步後是否公開'");
CALL add_col_if_not_exists('garmin_tokens', 'auto_sync',
  "TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否啟用自動同步'");

DROP PROCEDURE IF EXISTS add_col_if_not_exists;

-- ── Strava Webhook subscriptions ─────────────────────────
CREATE TABLE IF NOT EXISTS `strava_webhook_subscriptions` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subscription_id` INT             DEFAULT NULL COMMENT 'Strava subscription_id',
  `callback_url`    VARCHAR(500)    NOT NULL,
  `verify_token`    VARCHAR(100)    NOT NULL,
  `is_active`       TINYINT(1)      NOT NULL DEFAULT 1,
  `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
