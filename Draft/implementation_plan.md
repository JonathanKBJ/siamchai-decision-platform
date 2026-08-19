# แผนการดำเนินงาน Layer 1 (Data Extraction & Ingestion) & Layer 2 (PostgreSQL Operational DB)

แผนการดำเนินงานสำหรับการดึงข้อมูลจาก Oracle Core ERP (Layer 1) และการออกแบบ/จัดสร้างโครงสร้างฐานข้อมูล Operational Database บน PostgreSQL (Layer 2) เพื่อรองรับ **Siamchai AI Decision Platform**

---

## 📸 ภาพรวมระบบและขอบเขตงาน (Overview & Scope)

```text
Oracle Core ERP (10 Tables)
  ├── Master Data (shop, product, brand, category, group, type)
  ├── Transactions (sell, sell_detail)
  └── Inventory (stock_shop, bs_balance_stock)
         │
         ▼
[Layer 1] Data Extraction & Ingestion Pipeline (CDC / Batch Sync)
         │
         ▼
[Layer 2] PostgreSQL Operational Database
  ├── Target Schemas & Normalization
  ├── B-Tree, BRIN & Composite Indexing Strategy
  ├── Table Partitioning (sales, sales_items, inventory_history)
  └── PGVector Extension for Semantic Search Support
```

---

## ⚠️ User Review Required

> [!IMPORTANT]
> **การเลือกเทคโนโลยีสำหรับ Data Pipeline (Layer 1):**
> 1. **Debezium + Kafka / NATS JetStream (CDC)**: เหมาะสำหรับ Real-time Transaction Sync (`sell`, `sell_detail`, `stock_shop`) โดยจับ Change Data Capture จาก Oracle LogMiner/Redo Logs Direct
> 2. **Scheduled Batch ETL (Go Worker / Airflow / Temporal)**: เหมาะสำหรับ Initial Bulk Load และ Master Data Sync รายวัน (เช่น `product`, `shop`)

> [!NOTE]
> **กลยุทธ์ Primary Key ใน PostgreSQL (Layer 2):**
> แนะนำให้ใช้ `UUID` หรือ `BIGINT` เป็น Primary Key ฝั่ง PostgreSQL พร้อมเก็บ `oracle_id` เดิมไว้เป็น Unique Foreign Reference เพื่อป้องกัน Conflict เมื่อย้ายระบบในอนาคต

---

## ❓ Open Questions

> [!IMPORTANT]
> 1. **สิทธิ์การเข้าถึง Oracle Database**: มีการเปิดใช้งาน Oracle LogMiner หรือ Redo Log Archiving แล้วหรือยังเพื่อทำ CDC (Debezium)? หากยังไม่มี สามารถใช้วิธี Polling/Delta Sync ตาม `updated_at` / `created_at` ได้หรือไม่?
> 2. **ย้อนหลังกี่ปีสำหรับการย้ายข้อมูลช่วงแรก (Initial Load)**: ต้องการดึงข้อมูล Transaction การขายย้อนหลังกลับไปทั้งหมดกี่ปี (เช่น 3 ปี หรือ 5 ปี)?

---

## 🛠️ รายละเอียดขั้นตอนการดำเนินงาน (Proposed Plan)

### เฟสที่ 1: การออกแบบ Data Schema & Partitioning บน PostgreSQL (Layer 2)

#### 1.1 ออกแบบ PostgreSQL Table Schema (10 Tables Target)
สร้าง DDL Scripts สำหรับ PostgreSQL:
- **`branches`** (แทน `shop`)
- `id` (BIGINT/UUID PK), `oracle_id` (INT UNIQUE), `code` (VARCHAR), `name` (VARCHAR), `is_active` (BOOLEAN), `created_at`, `updated_at`
- **`product_brands`** (แทน `master_product_brand`)
- `id`, `oracle_id`, `name`, `status`
- **`product_categories`** (แทน `master_product_category`)
- `id`, `oracle_id`, `name`, `status`
- **`product_groups`** (แทน `product_group`)
- `id`, `oracle_id`, `description`, `status`
- **`product_types`** (แทน `master_product_type`)
- `id`, `oracle_id`, `name`, `status`
- **`products`** (แทน `product`)
- `id`, `oracle_id`, `pu_product_id`, `brand_id`, `category_id`, `group_id`, `type_id`, `description`, `status`
- **`sales`** (แทน `sell`) — *Declarative Partitioned Table*
- `id`, `oracle_id`, `sell_date` (TIMESTAMPTZ), `shop_id`, `created_at`
- **`sales_items`** (แทน `sell_detail`) — *Declarative Partitioned Table*
- `id`, `oracle_id`, `sale_id`, `product_id`, `shop_id`, `qty` (NUMERIC/INT), `sell_date` (TIMESTAMPTZ)
- **`stock_balance`** (แทน `stock_shop`)
- `id`, `shop_id`, `product_id`, `qty` (INT), `updated_at`
- **`stock_targets`** (แทน `bs_balance_stock`)
- `id`, `pu_product_id`, `shop_id`, `max_qty` (NUMERIC/INT), `updated_at`

#### 1.2 กำหนดกลยุทธ์ Table Partitioning
สำหรับตารางขนาดใหญ่ `sales` และ `sales_items`:
- ใช้ **Range Partitioning** ตาม `sell_date` แบบรายเดือน (Monthly Partition) เช่น:
  - `sales_y2025m01`, `sales_y2025m02`, ..., `sales_y2026m08`
- กำหนด Auto-Partition Creation Trigger / Script สำหรับสร้าง Partition ล่วงหน้าทุกเดือน

#### 1.3 ออกแบบ ดัชนี (Indexing Strategy)
- **Composite Index**:
  - `sales_items`: `(shop_id, product_id, sell_date)`
  - `sales`: `(shop_id, sell_date)`
  - `stock_balance`: `(shop_id, product_id)`
  - `stock_targets`: `(pu_product_id, shop_id)`
- **BRIN Index**:
  - สร้าง BRIN Index บน `sell_date` สำหรับ `sales` และ `sales_items` เพื่อประหยัดพื้นที่ดิสก์และเร่ง speed การค้นหาตามช่วงเวลา
- **PGVector Extension**:
  - เรียกใช้ `CREATE EXTENSION IF NOT EXISTS vector;`
  - เตรียมตาราง `product_embeddings` สำหรับเก็บ Vector Embedding ข้อมูลสินค้า (รองรับ RAG/Semantic Search ใน Layer 6)

---

### เฟสที่ 2: การพัฒนา Data Ingestion & Extraction Pipeline (Layer 1)

#### 2.1 พัฒนา Initial Bulk Data Loader (Historical Migration)
- พัฒนา Go Command / Script สำหรับดึงข้อมูลย้อนหลัง (Initial Load 3-5 ปี) จาก Oracle Core ERP (`ConnOracle`) เข้าสู่ PostgreSQL (`ConnPostgres`)
- ใช้วิธี Batch Insert ( chunk ละ 5,000 - 10,000 บรรทัด ) และใช้ Transaction เพื่อป้องกันข้อมูลหลุด

#### 2.2 พัฒนา Delta Sync / CDC Service (Real-time Transactions)
- **สตรีมข้อมูล Transaction (`sell`, `sell_detail`, `stock_shop`)**:
  - หากเลือก CDC: ติดตั้ง Debezium Connector ดักจับ Change Event จาก Oracle Redo Logs แล้วส่งผ่าน NATS JetStream / Kafka เข้า PostgreSQL Worker
  - หากเลือก Delta Sync: สร้าง Worker Polling ทุก 1-5 นาที ดึงข้อมูลที่มี `sell_date` หรือ `updated_at` ใหม่ล่าสุดด้วย Upsert (`ON CONFLICT DO UPDATE`)

#### 2.3 พัฒนา Master Data Sync Service (Batch Sync)
- สร้าง Cron Job / Schedule Sync รายวัน (เช่น เวลา 01:00 น.) เพื่อดึง Master Data (`product`, `shop`, `master_product_*`) และ `bs_balance_stock` มาอัปเดตสถานะใน PostgreSQL

---

### เฟสที่ 3: ระบบ Data Validation & Monitoring

- **Data Reconciliation Script**:
  - สร้าง Script ตรวจสอบความถูกต้องของจำนวนแถว (Row Count Check) และผลรวมสต๊อก/ยอดขาย (Sum Check) ระหว่าง Oracle และ PostgreSQL รายวัน
- **Dead Letter Queue (DLQ) & Alerting**:
  - ระบบเก็บ Log เมื่อดึงข้อมูลไม่สำเร็จ และส่งการแจ้งเตือนไปยังทีมผ่าน LINE / Slack / Alerting Dashboard

---

## 🧪 Verification Plan

### Automated Verification
1. **Migration Unit & Integration Tests**:
   - ทดสอบการเชื่อมต่อ Oracle (`OracleConn`) และ PostgreSQL (`PostgresConn`)
   - ทดสอบ Script Initial Bulk Load และ Delta Sync ด้วย Mock Data
2. **Performance Benchmarking**:
   - ทดสอบ Query Speed บน PostgreSQL ที่ทำ Partitioning เทียบกับ Table เดี่ยว ( Target: Response time < 100ms สำหรับยอดขาย 12 เดือน)

### Manual Verification
1. **Data Accuracy Audit**:
   - สุ่มตรวจบิลยอดขายและยอดสต๊อกระหว่าง Oracle ERP กับ PostgreSQL 100 รายการเพื่อยืนยันว่ามูลค่าตรงกัน 100%
2. **Failover & Re-sync Test**:
   - จำลองการปิด Data Pipeline เป็นเวลา 1 ชั่วโมง แล้วเปิดใหม่ เพื่อดูว่าระบบสามารถดึงข้อมูลย้อนหลังกลับมาสมบูรณ์โดยไม่มีข้อมูลตกหล่นหรือซ้ำซ้อน (Idempotency)
