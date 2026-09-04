-- ====================================================================
-- Siamchai AI Decision Platform - PostgreSQL Operational Schema DDL
-- Database Host : 122.155.164.15:5434
-- Database Name : siamchai_decision_db
-- Generated For : Decision Platform & Multi-Dimensional Analytics
-- ====================================================================

-- 0. Enable Required PostgreSQL Extensions
CREATE EXTENSION IF NOT EXISTS vector;      -- สำหรับ RAG / Vector Embedding
CREATE EXTENSION IF NOT EXISTS pg_trgm;     -- สำหรับ Trigram Full-text Search

-- ====================================================================
-- 1. Master Data Tables (ภูมิศาสตร์, สาขา, สินค้า, คู่ค้า)
-- ====================================================================

-- 1.1 ตารางภูมิภาค (Regions)
CREATE TABLE IF NOT EXISTS regions (
    id INT PRIMARY KEY,
    mysql_id INT,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE regions IS 'ตารางข้อมูลภูมิภาคหลัก (เช่น ภาคเหนือ, ภาคกลาง, ภาคอีสาน, ภาคใต้)';
COMMENT ON COLUMN regions.id IS 'รหัสภูมิภาคใน Oracle';
COMMENT ON COLUMN regions.mysql_id IS 'รหัสภูมิภาคเดิมใน MySQL';
COMMENT ON COLUMN regions.name IS 'ชื่อภูมิภาค';

-- 1.2 ตารางจังหวัด (Provinces)
CREATE TABLE IF NOT EXISTS provinces (
    id INT PRIMARY KEY,
    mysql_id INT,
    name VARCHAR(255) NOT NULL,
    region_id INT REFERENCES regions(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE provinces IS 'ตารางข้อมูลจังหวัด 77 จังหวัดทั่วประเทศไทย';
COMMENT ON COLUMN provinces.id IS 'รหัสจังหวัดใน Oracle';
COMMENT ON COLUMN provinces.mysql_id IS 'รหัสจังหวัดเดิมใน MySQL';
COMMENT ON COLUMN provinces.name IS 'ชื่อจังหวัด';
COMMENT ON COLUMN provinces.region_id IS 'รหัสภูมิภาคที่จังหวัดนี้สังกัด (FK -> regions.id)';

-- 1.3 ตารางสาขา (Branches / Shops)
CREATE TABLE IF NOT EXISTS branches (
    id INT PRIMARY KEY,
    code VARCHAR(50),
    name VARCHAR(255) NOT NULL,
    province_id INT REFERENCES provinces(id),
    mysql_province_id INT,
    region_id INT REFERENCES regions(id),
    mysql_region_id INT,
    location TEXT,
    tel VARCHAR(100),
    lat NUMERIC(10, 8),
    lng NUMERIC(11, 8),
    is_head_office BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE branches IS 'ตารางข้อมูลสาขา (Shop / Branch) พร้อมมิติเชิงภูมิศาสตร์และพิกัด GPS';
COMMENT ON COLUMN branches.id IS 'รหัสสาขา (Shop ID)';
COMMENT ON COLUMN branches.code IS 'รหัสย่อสาขา';
COMMENT ON COLUMN branches.name IS 'ชื่อสาขา';
COMMENT ON COLUMN branches.province_id IS 'รหัสจังหวัด (FK -> provinces.id)';
COMMENT ON COLUMN branches.mysql_province_id IS 'รหัสจังหวัดเดิม MYSQLPROVINCEID จากตาราง SHOP';
COMMENT ON COLUMN branches.region_id IS 'รหัสภูมิภาค (FK -> regions.id)';
COMMENT ON COLUMN branches.mysql_region_id IS 'รหัสภูมิภาคเดิม MYSQLREGIONID จากตาราง SHOP';
COMMENT ON COLUMN branches.location IS 'ที่อยู่ / ที่ตั้งสาขา';
COMMENT ON COLUMN branches.tel IS 'เบอร์โทรศัพท์ติดต่อสาขา';
COMMENT ON COLUMN branches.lat IS 'พิกัด GPS ละติจูด (Latitude)';
COMMENT ON COLUMN branches.lng IS 'พิกัด GPS ลองจิจูด (Longitude)';
COMMENT ON COLUMN branches.is_head_office IS 'แฟล็กระบุว่าเป็นสำนักงานใหญ่หรือไม่';
COMMENT ON COLUMN branches.is_active IS 'สถานะการเปิดใช้งานสาขา';

-- 1.4 ตารางแบรนด์สินค้า (Product Brands)
CREATE TABLE IF NOT EXISTS product_brands (
    id INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE product_brands IS 'ตารางแบรนด์สินค้า (Master Product Brand)';

-- 1.5 ตารางหมวดหมู่สินค้า (Product Categories)
CREATE TABLE IF NOT EXISTS product_categories (
    id INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE product_categories IS 'ตารางหมวดหมู่สินค้าหลัก (Master Product Category)';

-- 1.6 ตารางกลุ่มสินค้า (Product Groups)
CREATE TABLE IF NOT EXISTS product_groups (
    id INT PRIMARY KEY,
    description VARCHAR(255),
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE product_groups IS 'ตารางกลุ่มสินค้า (Product Group)';

-- 1.7 ตารางประเภทสินค้า (Product Types)
CREATE TABLE IF NOT EXISTS product_types (
    id INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE product_types IS 'ตารางประเภทสินค้า (Master Product Type)';

-- 1.8 ตารางข้อมูลสินค้า (Products)
CREATE TABLE IF NOT EXISTS products (
    id INT PRIMARY KEY,
    code VARCHAR(255),
    pu_product_id INT,
    brand_id INT REFERENCES product_brands(id),
    category_id INT REFERENCES product_categories(id),
    group_id INT REFERENCES product_groups(id),
    type_id INT REFERENCES product_types(id),
    description TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE products IS 'ตารางข้อมูลสินค้าทั้งหมด (Master Products)';
COMMENT ON COLUMN products.code IS 'รหัสสินค้า (Product Code)';
COMMENT ON COLUMN products.pu_product_id IS 'รหัส PU Product (เชื่อมโยงระบบสต็อก)';

-- 1.9 ตารางคู่ค้า / ซัพพลายเออร์ (Suppliers)
CREATE TABLE IF NOT EXISTS suppliers (
    id INT PRIMARY KEY,
    code VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50),
    address VARCHAR(255),
    tel VARCHAR(100),
    fax VARCHAR(100),
    contact_name VARCHAR(100),
    email VARCHAR(100),
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE suppliers IS 'ตารางข้อมูลผู้จัดจำหน่าย/คู่ค้า (Company Type = S)';

-- 1.10 ตารางลูกค้า (Customers)
CREATE TABLE IF NOT EXISTS customers (
    id INT PRIMARY KEY,
    code VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50),
    nation_id VARCHAR(20),
    address VARCHAR(255),
    tel VARCHAR(100),
    email VARCHAR(100),
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE customers IS 'ตารางข้อมูลลูกค้า';

-- ====================================================================
-- 2. Partitioned Transaction Tables (sell & sell_detail)
-- ====================================================================

-- 2.1 ตารางหัวบิลขาย (Sell Header - Partitioned by sell_date)
CREATE TABLE IF NOT EXISTS sell (
    id BIGINT NOT NULL,
    shop_id INT REFERENCES branches(id),
    sell_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, sell_date)
) PARTITION BY RANGE (sell_date);
COMMENT ON TABLE sell IS 'ตารางหัวบิลขาย (Partitioned by Month ตาม sell_date)';

-- 2.2 ตารางรายการขายย่อย (Sell Detail - Partitioned by sell_date)
CREATE TABLE IF NOT EXISTS sell_detail (
    id BIGINT NOT NULL,
    sell_id BIGINT NOT NULL,
    company_id INT REFERENCES suppliers(id),
    product_id INT REFERENCES products(id),
    shop_id INT REFERENCES branches(id),
    qty INT NOT NULL DEFAULT 0,
    sell_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, sell_date)
) PARTITION BY RANGE (sell_date);
COMMENT ON TABLE sell_detail IS 'ตารางรายการสินค้าในบิลขาย (Partitioned by Month ตาม sell_date)';

-- Partitions เริ่มต้น และ Default Overflow Partitions
CREATE TABLE IF NOT EXISTS sell_default PARTITION OF sell DEFAULT;
CREATE TABLE IF NOT EXISTS sell_detail_default PARTITION OF sell_detail DEFAULT;

-- ตัวอย่าง Monthly Partition สำหรับปี 2026
CREATE TABLE IF NOT EXISTS sell_y2026m08 PARTITION OF sell
    FOR VALUES FROM ('2026-08-01 00:00:00+07') TO ('2026-09-01 00:00:00+07');

CREATE TABLE IF NOT EXISTS sell_detail_y2026m08 PARTITION OF sell_detail
    FOR VALUES FROM ('2026-08-01 00:00:00+07') TO ('2026-09-01 00:00:00+07');

CREATE TABLE IF NOT EXISTS sell_y2026m09 PARTITION OF sell
    FOR VALUES FROM ('2026-09-01 00:00:00+07') TO ('2026-10-01 00:00:00+07');

CREATE TABLE IF NOT EXISTS sell_detail_y2026m09 PARTITION OF sell_detail
    FOR VALUES FROM ('2026-09-01 00:00:00+07') TO ('2026-10-01 00:00:00+07');

-- ====================================================================
-- 3. Stock & Inventory Tables (ยอดคงเหลือและเป้าสต็อก)
-- ====================================================================

-- 3.1 ตารางยอดคงเหลือสต็อก (Stock Balance / Snapshot)
CREATE TABLE IF NOT EXISTS stock_balance (
    id SERIAL PRIMARY KEY,
    shop_id INT REFERENCES branches(id),
    product_id INT REFERENCES products(id),
    pu_product_id INT,
    qty INT NOT NULL DEFAULT 0,
    current_qty NUMERIC(12, 2) DEFAULT 0,
    truesell_qty_7_day NUMERIC(12, 2) DEFAULT 0,
    target_qty_7_day_original NUMERIC(12, 2) DEFAULT 0,
    target_qty_7_day NUMERIC(12, 2) DEFAULT 0,
    current_qty_by_type NUMERIC(12, 2) DEFAULT 0,
    target_qty_by_type_7_day NUMERIC(12, 2) DEFAULT 0,
    min_qty NUMERIC(12, 2) DEFAULT 0,
    max_qty NUMERIC(12, 2) DEFAULT 0,
    po_qty NUMERIC(12, 2) DEFAULT 0,
    trans_in_qty NUMERIC(12, 2) DEFAULT 0,
    trans_out_qty NUMERIC(12, 2) DEFAULT 0,
    po_status VARCHAR(50) DEFAULT 'รอตัดสินใจ',
    truesell_qty_30_day NUMERIC(12, 2) DEFAULT 0,
    target_qty_30_day_original NUMERIC(12, 2) DEFAULT 0,
    target_qty_30_day NUMERIC(12, 2) DEFAULT 0,
    target_qty_by_type_30_day NUMERIC(12, 2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_stock_balance_shop_pu UNIQUE (shop_id, pu_product_id)
);
COMMENT ON TABLE stock_balance IS 'ตารางยอดคงเหลือสต็อกสินค้าจริงและยอดขายย้อนหลัง (bs_balance_stock)';

-- 3.2 ตารางการตั้งค่าเป้าสต็อก (Stock Targets / Settings)
CREATE TABLE IF NOT EXISTS stock_targets (
    id SERIAL PRIMARY KEY,
    shop_id INT REFERENCES branches(id),
    product_group_id INT REFERENCES product_groups(id),
    category_id INT REFERENCES product_categories(id),
    type_id INT REFERENCES product_types(id),
    brand_id INT REFERENCES product_brands(id),
    pu_product_id INT,
    min_qty NUMERIC(12, 2) DEFAULT 0,
    max_qty NUMERIC(12, 2) DEFAULT 0,
    sell_multiply NUMERIC(12, 4) DEFAULT 0,
    show_qty NUMERIC(12, 2) DEFAULT 0,
    created_by INT,
    updated_by INT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_stock_targets_shop_group UNIQUE (shop_id, product_group_id)
);
COMMENT ON TABLE stock_targets IS 'ตารางเกณฑ์เป้าหมายสต็อกขั้นต่ำ-สูงสุดแยกตามสาขาและกลุ่มสินค้า (bs_pg_setting)';

-- ====================================================================
-- 4. AI Semantic & Vector Search Table (PGVector)
-- ====================================================================

-- 4.1 ตารางเวกเตอร์ค้นหาสินค้าเชิงความหมาย (Product Embeddings)
CREATE TABLE IF NOT EXISTS product_embeddings (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id),
    content TEXT NOT NULL,
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE product_embeddings IS 'ตารางเวกเตอร์ Embedding สำหรับ AI RAG Semantic Search (1536 dimensions)';

-- ====================================================================
-- 5. Performance Indexes
-- ====================================================================

-- Indexes สำหรับมิติภูมิศาสตร์และการค้นหาสาขา
CREATE INDEX IF NOT EXISTS idx_branches_province ON branches (province_id);
CREATE INDEX IF NOT EXISTS idx_branches_region ON branches (region_id);
CREATE INDEX IF NOT EXISTS idx_branches_mysql_prov ON branches (mysql_province_id);
CREATE INDEX IF NOT EXISTS idx_branches_mysql_reg ON branches (mysql_region_id);
CREATE INDEX IF NOT EXISTS idx_provinces_region ON provinces (region_id);

-- Indexes สำหรับรายการขายและการวิเคราะห์ Transaction
CREATE INDEX IF NOT EXISTS idx_sell_shop_date ON sell (shop_id, sell_date);
CREATE INDEX IF NOT EXISTS idx_sell_detail_shop_prod_date ON sell_detail (shop_id, product_id, sell_date);
CREATE INDEX IF NOT EXISTS idx_sell_detail_company_date ON sell_detail (company_id, sell_date);
CREATE INDEX IF NOT EXISTS idx_sell_detail_date_brin ON sell_detail USING BRIN (sell_date);

-- Indexes สำหรับสต็อกและการค้นหาสินค้า
CREATE INDEX IF NOT EXISTS idx_products_code ON products (code);
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category_id);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products (brand_id);
CREATE INDEX IF NOT EXISTS idx_products_group ON products (group_id);
CREATE INDEX IF NOT EXISTS idx_products_type ON products (type_id);
CREATE INDEX IF NOT EXISTS idx_stock_balance_shop ON stock_balance (shop_id);
CREATE INDEX IF NOT EXISTS idx_stock_balance_prod ON stock_balance (product_id);
CREATE INDEX IF NOT EXISTS idx_stock_targets_shop ON stock_targets (shop_id);

-- ====================================================================
-- 6. Analytical Views (มุมมองเชิงวิเคราะห์ยอดขายและสต็อกรายภูมิภาค/จังหวัด)
-- ====================================================================

-- View สรุปข้อมูลสาขาพร้อมชื่อจังหวัดและภูมิภาค
CREATE OR REPLACE VIEW v_branch_geographic_summary AS
SELECT 
    b.id AS branch_id,
    b.code AS branch_code,
    b.name AS branch_name,
    p.id AS province_id,
    p.name AS province_name,
    r.id AS region_id,
    r.name AS region_name,
    b.location,
    b.tel,
    b.lat,
    b.lng,
    b.is_head_office,
    b.is_active
FROM branches b
LEFT JOIN provinces p ON b.province_id = p.id
LEFT JOIN regions r ON b.region_id = r.id;

COMMENT ON VIEW v_branch_geographic_summary IS 'มุมมองสรุปข้อมูลสาขาพร้อมมิติชื่อจังหวัด ภูมิภาค และพิกัดแผนที่';
