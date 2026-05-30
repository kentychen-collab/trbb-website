-- ═══════════════════════════════════════════════════════════
-- Migration: 04_user_profile_fields.sql
-- 擴充 users 資料表，加入完整會員資料欄位
-- 使用 stored procedure 避免 MySQL 8.0 IF NOT EXISTS bug
-- ═══════════════════════════════════════════════════════════

DROP PROCEDURE IF EXISTS add_col_if_not_exists;

DELIMITER $$
CREATE PROCEDURE add_col_if_not_exists(
  IN tbl  VARCHAR(64),
  IN col  VARCHAR(64),
  IN def  TEXT
)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME   = tbl
      AND COLUMN_NAME  = col
  ) THEN
    SET @sql = CONCAT('ALTER TABLE `', tbl, '` ADD COLUMN `', col, '` ', def);
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END$$
DELIMITER ;

-- ── 逐欄位加入 ──────────────────────────────────────────────
CALL add_col_if_not_exists('users', 'name_zh',
  "VARCHAR(50) DEFAULT NULL COMMENT '中文姓名' AFTER `display_name`");

CALL add_col_if_not_exists('users', 'name_en',
  "VARCHAR(100) DEFAULT NULL COMMENT '英文姓名' AFTER `name_zh`");

CALL add_col_if_not_exists('users', 'id_number',
  "VARCHAR(20) DEFAULT NULL COMMENT '身份證字號' AFTER `name_en`");

CALL add_col_if_not_exists('users', 'passport_number',
  "VARCHAR(20) DEFAULT NULL COMMENT '護照號碼' AFTER `id_number`");

CALL add_col_if_not_exists('users', 'gender',
  "TINYINT(1) DEFAULT NULL COMMENT '1=男,2=女,3=其他' AFTER `passport_number`");

CALL add_col_if_not_exists('users', 'birthday',
  "DATE DEFAULT NULL COMMENT '出生年月日' AFTER `gender`");

CALL add_col_if_not_exists('users', 'shirt_size',
  "VARCHAR(10) DEFAULT NULL COMMENT 'XS/S/M/L/XL/2XL/3XL' AFTER `birthday`");

CALL add_col_if_not_exists('users', 'food_type',
  "TINYINT(1) DEFAULT NULL COMMENT '1=葷,2=素,3=全素' AFTER `shirt_size`");

CALL add_col_if_not_exists('users', 'address',
  "VARCHAR(300) DEFAULT NULL COMMENT '通訊地址' AFTER `food_type`");

CALL add_col_if_not_exists('users', 'emergency_relation',
  "VARCHAR(30) DEFAULT NULL COMMENT '緊急聯絡人關係' AFTER `emergency_phone`");

-- ── 修改既有欄位 comment（不影響資料）──────────────────────
ALTER TABLE `users`
  MODIFY COLUMN `emergency_contact` VARCHAR(50)  DEFAULT NULL COMMENT '緊急聯絡人姓名',
  MODIFY COLUMN `emergency_phone`   VARCHAR(20)  DEFAULT NULL COMMENT '緊急聯絡人手機';

-- ── 清理 ────────────────────────────────────────────────────
DROP PROCEDURE IF EXISTS add_col_if_not_exists;
