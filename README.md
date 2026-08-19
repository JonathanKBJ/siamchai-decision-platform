# Siamchai AI Decision Platform — Data Ingestion & Database Layer

โปรเจกต์สำหรับระบบ **Siamchai AI Decision Platform** (Layer 1: Data Extraction & Ingestion / Layer 2: PostgreSQL Operational Database) รองรับการนำเข้าข้อมูลแบบ Incremental Sync ประจำวัน, การซ่อมแซมข้อมูลย้อนหลัง (Data Repair/Backfill), และการเชื่อมโยงข้อมูล Oracle ERP แบบ Idempotent

---

## 📌 ข้อมูลการเชื่อมต่อฐานข้อมูล (Database Configuration)

### 1. PostgreSQL Operational Database (เป้าหมาย)
- **Host**: `122.155.164.15`
- **Port**: `5434`
- **User**: `postgres`
- **Password**: `dev!4555@2026`
- **Target Database Name**: `siamchai_decision_db` *(ระบบจะสร้างให้อัตโนมัติในครั้งแรกที่รัน)*

### 2. Oracle Core ERP Database (ต้นทาง)
- **Host**: `10.0.1.32` / `10.0.1.152`
- **Port**: `1521`
- **Service Name**: `ORCL`
- **User**: `siamchai_stock`
- **Access Mode**: Read-Only (ปลอดภัย ไม่กระทบระบบ ERP หลัก)

---

## 🚀 โหมดการใช้งานและการรันระบบ (Execution Modes & CLI Flags)

ระบบรองรับการรันผ่าน **CLI Flags** และ **Environment Variables** เพื่อความยืดหยุ่นในการตั้งค่า Cron Job หรือการสั่งซ่อมข้อมูลย้อนหลัง:

```powershell
# 1. โหมดนำเข้าประจำวัน (Daily Incremental Sync - ค่าเริ่มต้น)
# ดึงข้อมูลย้อนหลัง N วันเพื่อเก็บสะสมยอดขายและอัปเดตบิลที่มีการแก้ไข (ไม่ DROP/TRUNCATE ตาราง)
go run .\cmd\main.go --mode=daily --lookback-days=7

# 2. โหมดซ่อมแซม/เติมข้อมูลย้อนหลังตามช่วงวันที่ (Data Repair / Range Backfill)
# ดึงเฉพาะข้อมูลช่วงวันที่ระบุ เพื่อซ่อมข้อมูลย้อนหลังโดยไม่กระทบช่วงเวลาอื่น
go run .\cmd\main.go --mode=repair --start=2025-01-01 --end=2026-07-30

# 3. โหมด Full Initial Reset (เคลียร์ตารางและรันนำเข้าใหม่ทั้งหมดครั้งแรก)
# บังคับ Drop & Recreate Schema ตารางทั้งหมดใหม่ก่อนนำเข้า
go run .\cmd\main.go --mode=full --full-reset=true
```

### พารามิเตอร์ CLI Flags
| Flag | Description | Default |
| :--- | :--- | :--- |
| `--mode` | โหมดการรัน: `daily`, `repair`, หรือ `full` | `daily` |
| `--start` | วันที่เริ่มต้น (รูปแบบ `YYYY-MM-DD`) สำหรับโหมด repair/full | `SYSDATE - 7` |
| `--end` | วันที่สิ้นสุด (รูปแบบ `YYYY-MM-DD`) สำหรับโหมด repair/full | `SYSDATE` |
| `--lookback-days` | จำนวนวันย้อนหลังสำหรับโหมด daily | `7` |
| `--full-reset` | บังคับ Drop และสร้างตารางทั้งหมดใหม่ก่อนรัน (`true`/`false`) | `false` |

---

## 🗄️ โครงสร้าง 10 ตารางหลักบน PostgreSQL (`siamchai_decision_db`)

1. `branches` — ข้อมูลสาขา (Primary Key ตรงกับ `ID` ใน Oracle `shop`)
2. `product_brands` — แบรนด์สินค้า (ตรงกับ Oracle `master_product_brand`)
3. `product_categories` — หมวดหมู่สินค้า (ตรงกับ Oracle `master_product_category`)
4. `product_groups` — กลุ่มสินค้า (ตรงกับ Oracle `product_group`)
5. `product_types` — ประเภทสินค้า (ตรงกับ Oracle `master_product_type`)
6. `products` — SKU สินค้า (ตรงกับ Oracle `product`)
7. `suppliers` — บริษัทผู้จำหน่าย (ตรงกับ Oracle `company` เฉพาะ `WHERE type = 'S'`)
8. `customers` — โครงสร้างข้อมูลลูกค้า (เตรียมไว้สำหรับรองรับอนาคต)
9. `sell` — หัวข้อบิลขาย *(ทำ Range Partitioning รายเดือน + Index `shop_id, sell_date`)*
10. `sell_detail` — รายการสินค้าที่ขาย *(มี `company_id REFERENCES suppliers(id)` + Range Partitioning รายเดือน + BRIN Index)*
11. `stock_balance` — สต๊อกคงเหลือปัจจุบันตามสาขา (`stock_shop`)
12. `stock_targets` — เป้าหมายสต๊อกสูงสุดตามสาขา (`bs_balance_stock`)
13. `product_embeddings` — ตารางรองรับ PGVector สำหรับ RAG / Semantic Search

---

## 🛡️ มาตรการความปลอดภัยและ Foreign Key Integrity

1. **Idempotent Upserts:** ใช้คำสั่ง `ON CONFLICT (id) DO UPDATE` หากข้อมูลมีอยู่แล้วจะถูก `UPDATE` ล่าสุดจาก Oracle หากเป็นข้อมูลใหม่จะถูก `INSERT` ให้อัตโนมัติ ข้อมูลจะไม่ซ้ำซ้อน
2. **Foreign Key Integrity Validation:** ก่อนนำเข้าตารางที่มีความสัมพันธ์ (`products`, `sell`, `sell_detail`, `stock_balance`) ระบบจะทำการตรวจสอบ ID อ้างอิงกับตารางแม่ผ่าน Memory Lookup Maps หากพบ ID ที่เลิกใช้แล้วใน Oracle ระบบจะแปลงเป็น `NULL` (`nil`) ป้องกันข้อผิดพลาด Foreign Key Violation บน PostgreSQL
3. **Partition Provisioning:** ระบบทำการตรวจสอบและสร้างตาราง Range Partition รายเดือน (`sell_yYYYYmMM`, `sell_detail_yYYYYmMM`) ล่วงหน้าให้อัตโนมัติตามช่วงวันที่รัน

---

## 📁 โครงสร้างโปรเจกต์ (Directory Structure)

```text
siamchai-decision-platform/
├── .env                           # ไฟล์การตั้งค่า Environment & DB Passwords
├── go.mod                         # Go Module Definition
├── README.md                      # เอกสารคู่มือระบบและการใช้งาน
├── cmd/
│   └── main.go                    # Entry point หลัก (รับ CLI flags & สั่งงาน Ingestion Engine)
├── database/
│   ├── database.go                # ฟังก์ชันเชื่อมต่อ Postgres และ Oracle (Auto-create Target DB)
│   └── partitions.go              # ระบบ Auto Provisioning Range Partitions รายเดือน
├── migrations/
│   └── 0001_init_schema.sql       # DDL SQL Script นิยามตาราง, Foreign Keys, Partitions & Indexes
└── pkg/
    ├── models/
    │   └── models.go              # Structs นิยาม GORM Models (autoIncrement:false)
    ├── repository/
    │   ├── repository.go          # Data Repositories & Upsert Operations
    │   └── repository_test.go     # Unit Tests สำหรับ Repository Layer
    └── ingestion/
        ├── extractor.go           # Extractor ทดสอบการเชื่อมต่อ Oracle Read-Only
        └── migrator.go            # Data Migrator Engine (ดึงข้อมูล, กรอง FK และ Sync เข้า Postgres)
```
#   s i a m c h a i - d e c i s i o n - p l a t f o r m  
 