-- ═══════════════════════════════════════════════════════════
-- Migration: 06_site_settings.sql
-- 網站環境設定表
-- ═══════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS `site_settings` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `key`        VARCHAR(100)    NOT NULL UNIQUE COMMENT '設定鍵名',
  `value`      MEDIUMTEXT      DEFAULT NULL COMMENT '設定值（JSON 或純字串）',
  `type`       VARCHAR(20)     NOT NULL DEFAULT 'string'
               COMMENT 'string|json|color|gradient|image|number|boolean',
  `group`      VARCHAR(50)     NOT NULL DEFAULT 'general' COMMENT '設定群組',
  `label`      VARCHAR(100)    DEFAULT NULL COMMENT '顯示名稱',
  `updated_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP
               ON UPDATE CURRENT_TIMESTAMP,
  `updated_by` BIGINT UNSIGNED DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ── 預設值 ───────────────────────────────────────────────────
INSERT INTO `site_settings` (`key`, `value`, `type`, `group`, `label`) VALUES
-- Logo
('logo_image',    NULL,                     'image',    'brand',  'Logo 圖片'),
('logo_text',     'TRBB 鐵人拔巴',           'string',   'brand',  'Logo 說明文字'),
('logo_text_size','1rem',                   'string',   'brand',  'Logo 文字大小'),
-- Banner
('banner_image',  NULL,                     'image',    'banner', '橫幅圖片'),
('banner_image_2',NULL,                     'image',    'banner', '橫幅圖片 2（備用）'),
('banner_text',   NULL,                     'string',   'banner', '橫幅說明文字'),
('banner_link',   NULL,                     'string',   'banner', '橫幅連結 URL'),
('banner_visible','1',                      'boolean',  'banner', '顯示橫幅'),
-- Background
('bg_color',  '{"type":"solid","colors":["#FFFFF3"]}', 'gradient','theme','底色'),
('bg2_color', '{"type":"solid","colors":["#F5F5E8"]}', 'gradient','theme','次要底色'),
('card_color','{"type":"solid","colors":["#FFFFFF"]}',  'gradient','theme','卡片底色'),
('border_color','#E0E3DA',                  'color',    'theme',  '邊框顏色'),
-- Brand colors
('primary_color','#CF2027',                 'color',    'theme',  '主色（品牌紅）'),
('navy_color',   '#1A3A7A',                 'color',    'theme',  '海軍藍'),
('accent_color', '#A593E0',                 'color',    'theme',  '強調色（紫）'),
-- Typography — Body
('font_body',       'Barlow',               'string',   'typography','內文字體'),
('font_body_size',  '16px',                 'string',   'typography','內文字體大小'),
('font_body_color', '#566270',              'color',    'typography','內文顏色'),
('font_body_weight','400',                  'string',   'typography','內文字重'),
-- Typography — Heading
('font_heading',        'Barlow Condensed', 'string',   'typography','標題字體'),
('font_heading_size',   '1.6rem',           'string',   'typography','標題字體大小'),
('font_heading_color',  '#1A3A7A',          'color',    'typography','標題顏色'),
('font_heading_weight', '700',              'string',   'typography','標題字重'),
-- Typography — Display (hero)
('font_display',        'Bebas Neue',       'string',   'typography','展示字體（大標）'),
('font_display_color',  '#1A3A7A',          'color',    'typography','展示字顏色'),
-- Navbar
('navbar_bg',   '{"type":"solid","colors":["rgba(255,255,243,0.95)"]}', 'gradient','navbar','導覽列背景'),
('navbar_text', '#1A3A7A',                  'color',    'navbar', '導覽列文字顏色')
ON DUPLICATE KEY UPDATE `label`=VALUES(`label`);

-- Icon / Favicon
INSERT INTO `site_settings` (`key`, `value`, `type`, `group`, `label`) VALUES
('site_icon',    NULL, 'image', 'brand', '網站 Icon（Favicon）'),
('site_icon_lg', NULL, 'image', 'brand', '應用 Icon（192px）')
ON DUPLICATE KEY UPDATE `label`=VALUES(`label`);


-- ── Strava tokens ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS `strava_tokens` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`          BIGINT UNSIGNED NOT NULL UNIQUE,
  `access_token`     VARCHAR(512)    NOT NULL,
  `refresh_token`    VARCHAR(512)    NOT NULL,
  `token_type`       VARCHAR(32)     DEFAULT 'Bearer',
  `expires_at`       DATETIME        DEFAULT NULL,
  `strava_athlete_id` BIGINT         DEFAULT NULL COMMENT 'Strava athlete ID',
  `athlete_name`     VARCHAR(100)    DEFAULT NULL,
  `last_sync_at`     DATETIME        DEFAULT NULL,
  `created_at`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
