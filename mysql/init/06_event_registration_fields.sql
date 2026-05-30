-- ═══════════════════════════════════════════════════════════
-- Migration: 06_event_registration_fields.sql
-- 在 event_registrations 加入報名表單快照欄位
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

CALL add_col_if_not_exists('event_registrations','reg_name_zh',         "VARCHAR(50)  DEFAULT NULL COMMENT '中文姓名'");
CALL add_col_if_not_exists('event_registrations','reg_name_en',         "VARCHAR(100) DEFAULT NULL COMMENT '英文姓名'");
CALL add_col_if_not_exists('event_registrations','reg_id_number',       "VARCHAR(20)  DEFAULT NULL COMMENT '身份證字號'");
CALL add_col_if_not_exists('event_registrations','reg_passport_number', "VARCHAR(20)  DEFAULT NULL COMMENT '護照號碼'");
CALL add_col_if_not_exists('event_registrations','reg_gender',          "TINYINT(1)   DEFAULT NULL COMMENT '1=男,2=女,3=其他'");
CALL add_col_if_not_exists('event_registrations','reg_birthday',        "DATE         DEFAULT NULL COMMENT '出生年月日'");
CALL add_col_if_not_exists('event_registrations','reg_phone',           "VARCHAR(20)  DEFAULT NULL COMMENT '手機'");
CALL add_col_if_not_exists('event_registrations','reg_email',           "VARCHAR(100) DEFAULT NULL COMMENT 'Email'");
CALL add_col_if_not_exists('event_registrations','reg_shirt_size',      "VARCHAR(10)  DEFAULT NULL COMMENT '衣服尺寸'");
CALL add_col_if_not_exists('event_registrations','reg_food_type',       "TINYINT(1)   DEFAULT NULL COMMENT '1=葷,2=素,3=全素'");
CALL add_col_if_not_exists('event_registrations','reg_address',         "VARCHAR(300) DEFAULT NULL COMMENT '通訊地址'");
CALL add_col_if_not_exists('event_registrations','reg_emergency_contact',  "VARCHAR(50) DEFAULT NULL COMMENT '緊急聯絡人姓名'");
CALL add_col_if_not_exists('event_registrations','reg_emergency_phone',    "VARCHAR(20) DEFAULT NULL COMMENT '緊急聯絡人手機'");
CALL add_col_if_not_exists('event_registrations','reg_emergency_relation', "VARCHAR(30) DEFAULT NULL COMMENT '緊急聯絡人關係'");

DROP PROCEDURE IF EXISTS add_col_if_not_exists;
