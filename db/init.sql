SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE members (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid          VARCHAR(36) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    phone         VARCHAR(20),
    password_hash VARCHAR(255),
    name          VARCHAR(100) NOT NULL,
    avatar_url    VARCHAR(500),
    gender        TINYINT DEFAULT 0 COMMENT '0=未填 1=男 2=女',
    birthday      DATE,
    role          TINYINT DEFAULT 0 COMMENT '0=一般 1=管理員 2=超管',
    status        TINYINT DEFAULT 1 COMMENT '0=停用 1=啟用',
    oauth_provider VARCHAR(20),
    oauth_id      VARCHAR(100),
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE member_addresses (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    member_id  BIGINT UNSIGNED NOT NULL,
    label      VARCHAR(50),
    recipient  VARCHAR(100) NOT NULL,
    phone      VARCHAR(20) NOT NULL,
    zip        VARCHAR(10),
    city       VARCHAR(50),
    district   VARCHAR(50),
    address    VARCHAR(255) NOT NULL,
    is_default TINYINT DEFAULT 0,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE login_logs (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    member_id  BIGINT UNSIGNED NOT NULL,
    ip         VARCHAR(45),
    user_agent VARCHAR(500),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_member (member_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE categories (
    id        INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    parent_id INT UNSIGNED DEFAULT 0,
    name      VARCHAR(100) NOT NULL,
    slug      VARCHAR(100) NOT NULL UNIQUE,
    sort      INT DEFAULT 0,
    status    TINYINT DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE products (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category_id  INT UNSIGNED NOT NULL,
    name         VARCHAR(255) NOT NULL,
    slug         VARCHAR(255) NOT NULL UNIQUE,
    description  TEXT,
    cover_image  VARCHAR(500),
    price        DECIMAL(10,2) NOT NULL,
    sale_price   DECIMAL(10,2),
    stock        INT DEFAULT 0,
    status       TINYINT DEFAULT 1,
    is_featured  TINYINT DEFAULT 0,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_images (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT UNSIGNED NOT NULL,
    url        VARCHAR(500) NOT NULL,
    sort       INT DEFAULT 0,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_variants (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT UNSIGNED NOT NULL,
    sku        VARCHAR(100) NOT NULL UNIQUE,
    spec_name  VARCHAR(100),
    price      DECIMAL(10,2),
    stock      INT DEFAULT 0,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE carts (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    member_id  BIGINT UNSIGNED NOT NULL,
    variant_id BIGINT UNSIGNED NOT NULL,
    quantity   INT NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_member_variant (member_id, variant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE orders (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_no        VARCHAR(50) NOT NULL UNIQUE,
    member_id       BIGINT UNSIGNED NOT NULL,
    type            TINYINT DEFAULT 1 COMMENT '1=購物 2=團報 3=二手',
    status          TINYINT DEFAULT 0 COMMENT '0=待付款 1=已付款 2=備貨中 3=已出貨 4=完成 5=取消',
    payment_method  VARCHAR(50),
    payment_status  TINYINT DEFAULT 0,
    payment_at      DATETIME,
    shipping_method VARCHAR(50),
    shipping_fee    DECIMAL(10,2) DEFAULT 0,
    total_amount    DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    final_amount    DECIMAL(10,2) NOT NULL,
    recipient_name  VARCHAR(100),
    recipient_phone VARCHAR(20),
    recipient_addr  VARCHAR(300),
    note            TEXT,
    ecpay_trade_no  VARCHAR(100),
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_member (member_id),
    INDEX idx_status (status),
    INDEX idx_order_no (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE order_items (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id    BIGINT UNSIGNED NOT NULL,
    product_id  BIGINT UNSIGNED,
    variant_id  BIGINT UNSIGNED,
    name        VARCHAR(255) NOT NULL,
    spec        VARCHAR(100),
    price       DECIMAL(10,2) NOT NULL,
    quantity    INT NOT NULL,
    subtotal    DECIMAL(10,2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE coupons (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code         VARCHAR(50) NOT NULL UNIQUE,
    type         TINYINT,
    value        DECIMAL(10,2) NOT NULL,
    min_amount   DECIMAL(10,2) DEFAULT 0,
    max_discount DECIMAL(10,2),
    total_quota  INT,
    used_count   INT DEFAULT 0,
    start_at     DATETIME,
    end_at       DATETIME,
    status       TINYINT DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE coupon_usage (
    id        BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    coupon_id BIGINT UNSIGNED NOT NULL,
    member_id BIGINT UNSIGNED NOT NULL,
    order_id  BIGINT UNSIGNED NOT NULL,
    used_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_member_coupon (member_id, coupon_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE events (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title        VARCHAR(255) NOT NULL,
    slug         VARCHAR(255) UNIQUE,
    category     VARCHAR(50),
    cover_image  VARCHAR(500),
    description  TEXT,
    event_date   DATE NOT NULL,
    location     VARCHAR(300),
    reg_start_at DATETIME,
    reg_end_at   DATETIME,
    max_quota    INT,
    status       TINYINT DEFAULT 1,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE event_tickets (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_id    BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(100) NOT NULL,
    price       DECIMAL(10,2) NOT NULL,
    quota       INT,
    sold_count  INT DEFAULT 0,
    reg_start   DATETIME,
    reg_end     DATETIME,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE event_registrations (
    id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id          BIGINT UNSIGNED NOT NULL,
    event_id          BIGINT UNSIGNED NOT NULL,
    ticket_id         BIGINT UNSIGNED NOT NULL,
    member_id         BIGINT UNSIGNED NOT NULL,
    participant       VARCHAR(100) NOT NULL,
    id_number         VARCHAR(255) COMMENT 'AES encrypted',
    emergency_contact VARCHAR(100),
    emergency_phone   VARCHAR(20),
    tshirt_size       VARCHAR(10),
    status            TINYINT DEFAULT 0,
    check_in_at       DATETIME,
    FOREIGN KEY (event_id) REFERENCES events(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE race_calendar (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    category    VARCHAR(50),
    event_date  DATE NOT NULL,
    location    VARCHAR(300),
    organizer   VARCHAR(200),
    reg_url     VARCHAR(500),
    description TEXT,
    tags        JSON,
    status      TINYINT DEFAULT 1,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_date (event_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE member_saved_races (
    member_id  BIGINT UNSIGNED,
    race_id    BIGINT UNSIGNED,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (member_id, race_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE second_hand_items (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    seller_id   BIGINT UNSIGNED NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    category_id INT UNSIGNED,
    `condition` TINYINT COMMENT '1=全新 2=近全新 3=良好 4=普通',
    price       DECIMAL(10,2) NOT NULL,
    images      JSON,
    status      TINYINT DEFAULT 0 COMMENT '0=審核中 1=上架 2=已售出 3=下架',
    views       INT DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_seller (seller_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE second_hand_inquiries (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    item_id    BIGINT UNSIGNED NOT NULL,
    buyer_id   BIGINT UNSIGNED NOT NULL,
    message    TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE garmin_tokens (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    member_id     BIGINT UNSIGNED NOT NULL UNIQUE,
    access_token  TEXT NOT NULL,
    token_secret  TEXT NOT NULL,
    scope         VARCHAR(255),
    connected_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE garmin_activities (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    member_id     BIGINT UNSIGNED NOT NULL,
    activity_id   VARCHAR(100) NOT NULL UNIQUE,
    activity_type VARCHAR(50),
    start_time    DATETIME,
    duration_sec  INT,
    distance_m    FLOAT,
    avg_hr        INT,
    calories      INT,
    raw_data      JSON,
    synced_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_member (member_id),
    INDEX idx_start (start_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;

-- =====================
-- 會員編號 & 審核狀態
-- member status: 0=待審核 1=啟用 2=停用
-- =====================
ALTER TABLE members ADD COLUMN member_no VARCHAR(50) NULL UNIQUE AFTER uuid;
ALTER TABLE members MODIFY COLUMN status TINYINT DEFAULT 0 COMMENT '0=待審核 1=啟用 2=停用';

-- =====================
-- 後台管理員（獨立於 members）
-- =====================
CREATE TABLE admin_users (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username      VARCHAR(50) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(100) NOT NULL,
    role          TINYINT DEFAULT 1 COMMENT '1=一般管理員 2=超級管理員',
    status        TINYINT DEFAULT 1 COMMENT '0=停用 1=啟用',
    created_by    BIGINT UNSIGNED COMMENT '建立者 admin_user id',
    last_login_at DATETIME,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 管理員權限（一般管理員才需要，超管預設全權限）
CREATE TABLE admin_permissions (
    admin_id          BIGINT UNSIGNED NOT NULL,
    manage_members    TINYINT DEFAULT 0,
    manage_events     TINYINT DEFAULT 0,
    manage_products   TINYINT DEFAULT 0,
    manage_orders     TINYINT DEFAULT 0,
    manage_second_hand TINYINT DEFAULT 0,
    updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (admin_id),
    FOREIGN KEY (admin_id) REFERENCES admin_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 商品分類 seed
INSERT IGNORE INTO categories (name, slug, sort, status) VALUES
('合作商品', 'partner', 1, 1),
('社團商品', 'club', 2, 1),
('二手福利', 'second-hand-welfare', 3, 1);
