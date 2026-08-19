# Siamchai AI Decision Platform — Planning Work Breakdown Structure (WBS)

เอกสารฉบับนี้สรุปโครงสร้างการแบ่งงาน (Work Breakdown Structure) สำหรับการพัฒนา **Siamchai AI Decision Platform** โดยตั้งต้นจากข้อมูลและตรรกะของระบบ **Planning** (`pkg/planning`) ในโปรเจกต์ `api-purchase` และแบ่งขอบเขตงานตามสถาปัตยกรรมระบบ (Architecture Layers) เพื่อให้แต่ละทีม/ผู้รับผิดชอบสามารถนำไปปฏิบัติงานต่อได้ทันที

---

## 🏗️ Architecture Overview

```text
Oracle / Existing ERP (api-purchase)
          │
          ▼
   [Layer 1] ETL / CDC / Data Sync
          │
          ├──────────────────────────────────┐
          ▼                                  ▼
   [Layer 2] PostgreSQL             [Layer 3] MinIO
   (Operational DB)                 (Data Lake / Parquet)
          │                                  │
          └─────────────────┬────────────────┘
                            ▼
                   [Layer 4] ClickHouse
                   (Analytics / OLAP Engine)
                            │
                            ▼
                   [Layer 5] Forecast Engine
                   (SMA / WAG / YoY Calculation)
                            │
                            ▼
                   [Layer 6] MCP Server
                   (AI Business Tools / Interfaces)
                            │
                            ▼
                   AI / LLM / Decision Support
```

---

## 📌 [Layer 1] Core ERP Data Extraction & Ingestion Pipeline
**เป้าหมาย:** สร้าง Data Pipeline ดึงข้อมูลที่จำเป็นจาก Oracle Core ERP เข้าสู่ Data Platform

### ข้อมูลที่ต้องดึง (Data Mapping จาก `pkg/planning`):
1. **Master Data Tables**:
   - `shop` (รหัส, ชื่อ, สถานะสาขา)
   - `master_product_brand` (แบรนด์สินค้า)
   - `master_product_category` (หมวดหมู่สินค้า)
   - `product_group` (กลุ่มสินค้า)
   - `master_product_type` (ประเภทสินค้า)
   - `product` (ข้อมูล SKU สินค้า, `pu_product_id`)
2. **Transaction Data Tables**:
   - `sell` (หัวข้อรายการขาย, `sell_date`)
   - `sell_detail` (รายละเอียดการขายสินค้า, `qty`, `shop_id`, `product_id`)
3. **Inventory & Target Stock Tables**:
   - `stock_shop` (สต๊อกคงเหลือปัจจุบันตามสาขา)
   - `bs_balance_stock` (เป้าหมายสต๊อกสูงสุด / `max_qty`)

### ขอบเขตงาน (Tasks):
- [ ] ออกแบบและสร้าง **CDC (Change Data Capture)** หรือ **Sync Job** สำหรับตาราง Transactions (`sell`, `sell_detail`, `stock_shop`)
- [ ] สร้าง **Batch Ingestion Job** รายวันสำหรับตาราง Master Data
- [ ] ตรวจสอบ Data Integrity และสร้างระบบแจ้งเตือนเมื่อการ Sync ข้อมูลล้มเหลว

---

## 📌 [Layer 2] Operational Database Layer (PostgreSQL)
**เป้าหมาย:** ออกแบบ Operational Database ใน PostgreSQL เพื่อเป็นฐานข้อมูลหลักสำหรับระบบใหม่

### ขอบเขตงาน (Tasks):
- [ ] ออกแบบ DDL / Schema ของ PostgreSQL ให้รองรับโครงสร้างข้อมูลทั้ง 10 ตาราง
- [ ] สร้าง B-Tree Index บน Primary Key และ Foreign Keys
- [ ] สร้าง Composite Index สำหรับ Query ยอดนิยม เช่น `(shop_id, product_id, sell_date)`
- [ ] ทำ **Table Partitioning** (ตาม Month/Year) สำหรับตาราง Transaction ขนาดใหญ่ เช่น `sales`, `sales_items`
- [ ] ติดตั้ง `PGVector` สำหรับรองรับการค้นหาข้อมูลบริบทสินค้าเชิงความหมาย (Semantic Search / RAG Document)

---

## 📌 [Layer 3] Data Lake Layer (MinIO)
**เป้าหมาย:** จัดเก็บประวัติข้อมูลยอดขายและสต๊อกย้อนหลัง 3–5 ปีในรูปแบบไฟล์ที่เหมาะกับ Analytics

### ขอบเขตงาน (Tasks):
- [ ] ตั้งค่า MinIO Bucket `siamchai-data-lake`
- [ ] พัฒนา ETL Pipeline แปลงข้อมูลประวัติการขาย (`sell` + `sell_detail`) เป็น **Apache Parquet (Compressed)**
- [ ] จัดโครงสร้าง Directory ใน Data Lake:
  ```text
  siamchai-data-lake/
  ├── sales/YYYY/MM/*.parquet
  ├── inventory/
  ├── products/
  └── shops/
  ```
- [ ] กำหนดนโยบาย Data Retention และ Backup

---

## 📌 [Layer 4] Analytics & OLAP Layer (ClickHouse)
**เป้าหมาย:** สร้าง Database เชิงวิเคราะห์ประมวลผลความเร็วสูง สำหรับรองรับการทำ Dashboard และ Forecast Engine

### ขอบเขตงาน (Tasks):
- [ ] ออกแบบ ClickHouse Tables โดยใช้ Engine ที่เหมาะสม (เช่น `MergeTree`, `SummingMergeTree`)
- [ ] นำข้อมูลจาก MinIO / PostgreSQL เข้าสู่ ClickHouse
- [ ] สร้าง Materialized Views สำหรับคำนวณการสรุปยอดขายรายเดือน (`sales_monthly_agg`) และรายปี (`sales_yearly_agg`)
- [ ] ทำ Benchmark Performance การ Query ยอดขายย้อนหลัง 5 ปีในระดับระดับสาขาและ SKU

---

## 📌 [Layer 5] Forecast & Decision Engine (Planning Logic Migration)
**เป้าหมาย:** ถอดตรรกะการคำนวณและคาดการณ์จาก `pkg/planning` มาสร้างเป็น Analytics & Forecast Microservice บน ClickHouse

### ตรรกะการคำนวณที่ต้องพอร์ตมา (Calculation Logic):
1. **Range A (Month-by-Month Moving Average)**:
   - **SMA (Simple Moving Average)**: คำนวณเปอร์เซ็นต์เปลี่ยนแปลงเคลื่อนที่
   - **WAG (Weighted Average Growth)**: คำนวณเปอร์เซ็นต์เปลี่ยนแปลงถ่วงน้ำหนักตามระยะเวลา
   - **SMA Order & WAG Order**:
     $$\text{Order} = (\text{TargetMax} \times \text{ForecastQty} + \text{ForecastQty}) - \text{CurrentStock}$$
2. **Range B (Year-over-Year Growth Analysis)**:
   - คำนวณ **Growth Avg** ย้อนหลัง 5 ปีสำหรับเดือนเดียวกัน (เช่น คำนวณความต้องการเมษายนของปีถัดไป)
   - คำนวณ **Estimate Sale Order** และ **Avg Sale Order**
3. **Safety Factor Adjustment**:
   - คำนวณบวกเพิ่มเปอร์เซ็นต์ความเสี่ยง (`Factor %`) เข้ากับยอดสั่งซื้อที่แนะนำ

### ขอบเขตงาน (Tasks):
- [ ] พัฒนา API Service สำหรับส่งคืนค่า Forecast & Recommendation ตามตัวกรอง (Shop, Brand, Category, Group, Type, Product)
- [ ] พัฒนาอัลกอริทึมเปรียบเทียบความแม่นยำระหว่าง SMA, WAG และ YoY Growth

---

## 📌 [Layer 6] MCP Server & AI Decision Support Tools
**เป้าหมาย:** สร้าง Business Interface Tools บน MCP Server ให้ AI / LLM สามารถดึงข้อมูลและสร้างข้อแนะนำในการตัดสินใจได้

### รายการ MCP Tools ที่ต้องพัฒนา (เน้นด้าน Planning & Purchasing):
1. **Sales & Demand Tools**:
   - `get_sales_summary`: ดูสรุปยอดขายตามสาขา/กลุ่มสินค้า
   - `get_sales_history`: ดึงสถิติยอดขายย้อนหลัง
   - `detect_sales_trend`: วิเคราะห์แนวโน้มยอดขาย MoM / YoY
2. **Inventory & Stock Tools**:
   - `get_stock_balance`: ตรวจสอบสต๊อกคงเหลือปัจจุบัน
   - `get_low_stock`: ตรวจสอบสินค้าที่สต๊อกต่ำกว่าเป้าหมาย
   - `get_overstock`: ตรวจสอบสินค้าที่สต๊อกเกินเป้าหมาย
3. **Purchase Recommendation Tools**:
   - `calculate_reorder`: คำนวณจำนวนสั่งซื้อที่เหมาะสมตาม SMA/WAG/YoY
   - `recommend_purchase`: แนะนำรายการสั่งซื้อสินค้าพร้อม Safety Factor
   - `create_purchase_proposal`: สร้างข้อเสนอสั่งซื้อ (Proposal) ให้ผู้บริหารอนุมัติ (Action Layer)

### ขอบเขตงาน (Tasks):
- [ ] พัฒนา MCP Server ตามมาตรฐาน Model Context Protocol
- [ ] ลงทะเบียน Tools และกำหนด JSON Schema สำหรับ Parameter Input/Output
- [ ] ทดสอบ AI Prompting และ Tool Calling เพื่อประเมินผลการแนะนำการตัดสินใจ

---

## 📅 สรุปส่วนงานตามความรับผิดชอบ (Role Allocation Summary)

| ทีม / บทบาท | เลเยอร์ที่รับผิดชอบหลัก | หน้าที่สำคัญ |
| :--- | :--- | :--- |
| **Data Engineer** | Layer 1, Layer 2, Layer 3, Layer 4 | ทำ CDC/ETL จาก Oracle, ออกแบบ Postgres, MinIO Parquet และ ClickHouse Aggregations |
| **Backend / Analytics Engineer** | Layer 5 (Forecast Engine) | แปลงสูตร SMA/WAG/YoY จาก `pkg/planning` มาพัฒนา API คำนวณบน ClickHouse |
| **AI / MCP Engineer** | Layer 6 (MCP Server) | สร้าง MCP Tools เชื่อมต่อ Forecast Engine และสร้าง AI Decision Prompts |
