-- ═══════════════════════════════════════════════════════════
-- Migration: 07_shop_fields.sql
-- 擴充 orders 加入付款方式、取貨方式欄位
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

-- 付款方式: 1=信用卡, 2=轉帳, 3=LINE Pay, 4=現金
CALL add_col_if_not_exists('orders', 'payment_method',
  "TINYINT(1) DEFAULT NULL COMMENT '1=信用卡,2=轉帳,3=LINE Pay,4=現金' AFTER `note`");

-- 取貨方式: 1=宅配, 2=自取
CALL add_col_if_not_exists('orders', 'delivery_method',
  "TINYINT(1) DEFAULT NULL COMMENT '1=宅配,2=自取' AFTER `payment_method`");

-- 自取地點（自取時填寫）
CALL add_col_if_not_exists('orders', 'pickup_location',
  "VARCHAR(200) DEFAULT NULL COMMENT '自取地點' AFTER `delivery_method`");

-- 付款狀態
CALL add_col_if_not_exists('orders', 'payment_status',
  "TINYINT(1) NOT NULL DEFAULT 0 COMMENT '0=未付款,1=已付款,2=已退款' AFTER `pickup_location`");

-- 付款時間
CALL add_col_if_not_exists('orders', 'paid_at',
  "DATETIME DEFAULT NULL AFTER `payment_status`");

-- 出貨時間
CALL add_col_if_not_exists('orders', 'shipped_at',
  "DATETIME DEFAULT NULL AFTER `paid_at`");

-- 物流單號
CALL add_col_if_not_exists('orders', 'tracking_number',
  "VARCHAR(100) DEFAULT NULL COMMENT '物流單號' AFTER `shipped_at`");

DROP PROCEDURE IF EXISTS add_col_if_not_exists;
