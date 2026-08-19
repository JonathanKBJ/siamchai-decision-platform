-- ====================================================================
-- Siamchai AI Decision Platform - PostgreSQL Operational Schema DDL
-- Database Host: 122.155.164.15:5434
-- Database Name: siamchai_decision_db
-- ====================================================================

-- Enable Vector extension for RAG / Semantic Search
CREATE EXTENSION IF NOT EXISTS vector;

-- 1. Master Data Tables
CREATE TABLE IF NOT EXISTS branches (
    id INT PRIMARY KEY,
    code VARCHAR(50),
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_brands (
    id INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_categories (
    id INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_groups (
    id INT PRIMARY KEY,
    description VARCHAR(255),
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_types (
    id INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    id INT PRIMARY KEY,
    pu_product_id VARCHAR(50),
    brand_id INT REFERENCES product_brands(id),
    category_id INT REFERENCES product_categories(id),
    group_id INT REFERENCES product_groups(id),
    type_id INT REFERENCES product_types(id),
    description TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

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

-- 2. Partitioned Transaction Tables (sell & sell_detail)
CREATE TABLE IF NOT EXISTS sell (
    id BIGINT NOT NULL,
    shop_id INT REFERENCES branches(id),
    sell_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, sell_date)
) PARTITION BY RANGE (sell_date);

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

-- Sample Monthly Partitions for 2026
CREATE TABLE IF NOT EXISTS sell_y2026m08 PARTITION OF sell
    FOR VALUES FROM ('2026-08-01 00:00:00+07') TO ('2026-09-01 00:00:00+07');

CREATE TABLE IF NOT EXISTS sell_detail_y2026m08 PARTITION OF sell_detail
    FOR VALUES FROM ('2026-08-01 00:00:00+07') TO ('2026-09-01 00:00:00+07');

-- Default Partition for Future/Historical overflow
CREATE TABLE IF NOT EXISTS sell_default PARTITION OF sell DEFAULT;
CREATE TABLE IF NOT EXISTS sell_detail_default PARTITION OF sell_detail DEFAULT;

-- Performance Indexes
CREATE INDEX IF NOT EXISTS idx_sell_shop_date ON sell (shop_id, sell_date);
CREATE INDEX IF NOT EXISTS idx_sell_detail_shop_prod_date ON sell_detail (shop_id, product_id, sell_date);
CREATE INDEX IF NOT EXISTS idx_sell_detail_company_date ON sell_detail (company_id, sell_date);
CREATE INDEX IF NOT EXISTS idx_sell_detail_date_brin ON sell_detail USING BRIN (sell_date);

-- 3. Stock & Inventory Tables
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
    CONSTRAINT idx_stock_balance_shop_prod UNIQUE (shop_id, product_id)
);

CREATE TABLE IF NOT EXISTS stock_targets (
    id SERIAL PRIMARY KEY,
    shop_id INT REFERENCES branches(id),
    product_group_id INT REFERENCES product_groups(id),
    category_id INT REFERENCES product_categories(id),
    type_id INT REFERENCES product_types(id),
    brand_id INT REFERENCES product_brands(id),
    pu_product_id VARCHAR(50),
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

-- 4. PGVector Semantic Embedding Table
CREATE TABLE IF NOT EXISTS product_embeddings (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id),
    content TEXT NOT NULL,
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
