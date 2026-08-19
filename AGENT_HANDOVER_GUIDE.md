# 🤖 Agent Handover & Context Guide: Siamchai AI Decision Platform

> **คำชี้แจงสำหรับ AI Agent และผู้พัฒนา:**
> เอกสารฉบับนี้สรุปบริบททั้งหมด ข้อตกลง สถาปัตยกรรม โครงสร้างฐานข้อมูล และกฎความปลอดภัยของโปรเจกต์ **Siamchai AI Decision Platform** ที่ได้จากการวิเคราะห์โมดูล `pkg/planning` และการจัดเตรียมโครงสร้างโปรเจกต์ใหม่ เพื่อให้ AI Agent ตัวอื่นหรือผู้พัฒนาสามารถเข้ามาอ่านทำความเข้าใจและปฏิบัติงานต่อได้อย่างถูกต้องปลอดภัย 100%

---

## 🎯 1. วัตถุประสงค์และภาพรวมระบบ (Project Context & Purpose)

เป้าหมายคือการสร้าง **Siamchai AI Decision Platform** เพื่อเชื่อมต่อกับ Core ERP เดิมของ Siamchai Service ช่วยในการวิเคราะห์ และสนับสนุนการตัดสินใจของผู้บริหาร เช่น:
- วิเคราะห์และ คาดการณ์ (Forecast) ยอดขายล่วงหน้า
- คำนวณปริมาณการสั่งซื้อ (Reorder Recommendation)
- แนะนำระดับ Stock ที่เหมาะสมแต่ละสาขา และการโยกย้ายสินค้าระหว่างคลัง
- ให้ผู้บริหารถามข้อมูล ERP ด้วยภาษาธรรมชาติ (Natural Language) ผ่าน MCP Server

จุดเริ่มต้นถูกถอดแบบมาจากตรรกะและโครงสร้างข้อมูลของโมดูล **Planning** (`d:\SIAMCHAI\CODE\SECURE_ON_CLOUD\api-purchase\pkg\planning`)

---

## 📊 2. สรุปตารางข้อมูล 10 ตารางหลักจาก Oracle Core ERP

จากการวิเคราะห์ `pkg/planning/repositories/planning_repo.go` ระบบ Planning ใช้ข้อมูล 10 ตารางหลักดังนี้:

| ลำดับ | ตารางเดิมใน Oracle | หน้าที่ / ประเภทข้อมูล | ตารางใหม่ใน PostgreSQL (`siamchai_decision_db`) |
| :---: | :--- | :--- | :--- |
| 1 | `shop` | Master ข้อมูลสาขา | `branches` |
| 2 | `master_product_brand` | Master แบรนด์สินค้า | `product_brands` |
| 3 | `master_product_category` | Master หมวดหมู่สินค้า | `product_categories` |
| 4 | `product_group` | Master กลุ่มสินค้า | `product_groups` |
| 5 | `master_product_type` | Master ประเภทสินค้า | `product_types` |
| 6 | `product` | Master ข้อมูล SKU สินค้า | `products` |
| 7 | `sell` | Transaction หัวข้อบิลขาย | `sales` *(Partitioned)* |
| 8 | `sell_detail` | Transaction รายการขายสินค้า | `sales_items` *(Partitioned)* |
| 9 | `stock_shop` | ยอดสต๊อกคงเหลือปัจจุบันตามสาขา | `stock_balance` |
| 10 | `bs_balance_stock` | เป้าหมายสต๊อกสูงสุด (Target Max) | `stock_targets` |

---

## 🏗️ 3. สถาปัตยกรรม 6 เลเยอร์ และขอบเขตงานปัจจุบัน

```text
Oracle Core ERP (api-purchase)
          │
          ▼
   [Layer 1] ETL / CDC / Data Sync  <-- (งานที่รับผิดชอบปัจจุบัน)
          │
          ├──────────────────────────────────┐
          ▼                                  ▼
   [Layer 2] PostgreSQL             [Layer 3] MinIO
   (Operational DB)                 (Data Lake / Parquet)
   <-- (งานที่รับผิดชอบปัจจุบัน)
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
```

---

## 💻 4. ข้อมูลโปรเจกต์ใหม่ (`siamchai-decision-platform`)

- **Location Path**: `d:\SIAMCHAI\CODE\SECURE_ON_CLOUD\siamchai-decision-platform\`
- **PostgreSQL Server**: `122.155.164.15:5434`
  - **Username**: `postgres`
  - **Password**: `dev!4555@2026`
  - **Database Name**: `siamchai_decision_db` *(มีโค้ดตรวจสอบอัตโนมัติ สั่งสร้างก้อน DB ใหม่ให้อัตโนมัติหากยังไม่มี)*
- **Oracle ERP DB (เดิม)**: `10.0.1.32:1521` (Service: `ORCL`, User: `siamchai_stock`) — **Read-Only Mode**

### โครงสร้างไฟล์ในโปรเจกต์:
- `.env`: ไฟล์ตั้งค่า DB Connection และ Security Flags
- `go.mod`: โมดูล Go Dependencies
- `migrations/0001_init_schema.sql`: DDL สำหรับสร้าง 10 ตารางใน PostgreSQL, Monthly Partitioning บน `sales` & `sales_items`, Composite Indexes และ PGVector Extension
- `database/database.go`: แพกเกจเชื่อมต่อ DB ทั้งสองระบบ (มีระบบ Auto-create Postgres DB)
- `pkg/models/models.go`: GORM Structs ของ 10 ตาราง
- `pkg/ingestion/extractor.go`: Extractor พร้อม Guardrails จำกัด Query ไม่เกิน 7 วัน
- `cmd/main.go`: Entry point รัน Migration และสอบทานระบบ

---

## ⛔ 5. กฎเหล็กและความปลอดภัย (CRITICAL SAFETY RULES FOR AGENTS)

> [!CAUTION]
> **AI Agent ตัวใดก็ตามที่เข้ามาทำงานในโปรเจกต์นี้ ต้องปฏิบัติตามกฎ 3 ข้อนี้อย่างเคร่งครัด:**

1. **ห้ามรันการดึงข้อมูลทั้งหมดไปเก็บจริง (No Full Ingestion Yet)**:
   - ห้ามรัน Bulk Sync หรือ ETL ข้อมูลย้อนหลังทั้งหมดจาก Oracle ลง PostgreSQL หรือ Disk ในช่วงนี้ เพราะจะสร้างภาระหนักกระทบกับการทำงานของระบบ Core ERP Production หลัก
2. **จำกัดช่วงเวลาค้นหาไม่เกิน 7 วัน (Strict <= 7 Days Filter)**:
   - การทดสอบ Query ข้อมูลจาก Oracle ERP **ต้องกำหนดเงื่อนไขย้อนหลังไม่เกิน 7 วันเท่านั้น** (`WHERE sell_date >= SYSDATE - 7` หรือ `TRUNC(SYSDATE - 7)`)
3. **เปิดใช้ Safety Dry-Run Mode เสมอ**:
   - ต้องคงค่า `SAFETY_DRY_RUN=true` ใน `.env` และใน Extractor ไว้ระหว่างการทดสอบระบบ

---

## 🚀 6. คำแนะนำสำหรับ AI Agent ในการสานต่องาน (Instruction Prompt for Next Agent)

หาก AI Agent หรือผู้พัฒนาต้องการดำเนินงานต่อในเฟสถัดไป ให้ปฏิบัติตามขั้นตอนดังนี้:

```text
1. อ่านไฟล์ AGENT_HANDOVER_GUIDE.md และ README.md ในโฟลเดอร์ siamchai-decision-platform ให้เข้าใจ
2. ตรวจสอบการเชื่อมต่อ PostgreSQL 122.155.164.15:5434/siamchai_decision_db
3. พัฒนา Unit Test / Integration Test สำหรับ Data Mapping ระหว่าง Oracle DTO กับ PostgreSQL Structs
4. ยึดถือข้อจำกัด Query ไม่เกิน 7 วันเมื่อทดสอบดึงข้อมูลจาก Oracle ERP เสมอ
```
