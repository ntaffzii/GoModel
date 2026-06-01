# System Test AI-LLM API

โฟลเดอร์นี้มีเครื่องมือสำหรับทดสอบการทำงานของ GoModel Gateway:

- **Open WebUI** — Web UI สำหรับสนทนากับโมเดล AI ทุกตัวที่ผ่าน GoModel
- **api_tester.py** — CLI script สำหรับทดสอบ API แบบ quick-check

## Prerequisites

- [Docker](https://www.docker.com/) ติดตั้งและรันอยู่
- GoModel Gateway รันอยู่ที่ port `8080` (ไม่ว่าจะรันแบบ `go run` หรือ Docker)

## Open WebUI (แนะนำ)

### เริ่มต้นใช้งาน

```bash
cd system-test-ai-llm-api
docker compose up -d
```

เปิดเบราว์เซอร์ไปที่ **http://localhost:3000**

> **หมายเหตุ:** ครั้งแรกที่เข้าใช้งาน ระบบจะให้สร้างบัญชี Admin อัตโนมัติ

### การตั้งค่า

แก้ไขไฟล์ `.env` ตามต้องการ:

| ตัวแปร | ค่าเริ่มต้น | คำอธิบาย |
|---|---|---|
| `GOMODEL_BASE_URL` | `http://host.docker.internal:8080/v1` | URL ของ GoModel Gateway |
| `GOMODEL_API_KEY` | `not-needed` | API Key ของ GoModel (ถ้าตั้ง `GOMODEL_MASTER_KEY` ไว้) |
| `OPEN_WEBUI_PORT` | `3000` | Port ที่ Open WebUI จะเปิดให้เข้าใช้ |

### หยุดการทำงาน

```bash
docker compose down
```

ข้อมูลแชทและการตั้งค่าจะถูกเก็บไว้ใน Docker volume `open-webui_data`

### ลบข้อมูลทั้งหมด (reset)

```bash
docker compose down -v
```

## CLI Tester (`api_tester.py`)

ใช้ทดสอบ API แบบ quick-check ผ่าน command line (ต้องติดตั้ง `openai` Python package)

```bash
pip install openai
```

**ทดสอบ GoModel (ค่าเริ่มต้น):**
```bash
python api_tester.py
```

**ระบุโมเดลเฉพาะ:**
```bash
python api_tester.py --provider gomodel --model "ollama/scb10x/typhoon2.5-qwen3-4b:latest"
```

**ระบุข้อความ Prompt:**
```bash
python api_tester.py --provider ollama --prompt "ช่วยเขียนฟังก์ชันบวกเลขใน Python ให้ดูหน่อย"
```

## สถาปัตยกรรม

```
┌──────────────┐        ┌──────────────┐        ┌──────────────────┐
│  Open WebUI  │───────▶│   GoModel    │───────▶│  AI Providers    │
│  :3000       │  HTTP  │   :8080      │  HTTP  │  (OpenAI, Ollama,│
│              │        │              │        │   UBUAI, etc.)   │
└──────────────┘        └──────────────┘        └──────────────────┘
```

Open WebUI ส่งทุก request ผ่าน GoModel Gateway ซึ่งจัดการ:
- **Multi-provider routing** — สลับค่ายได้อัตโนมัติ
- **Caching** — ลดค่าใช้จ่าย
- **Token usage tracking** — ติดตามการใช้งาน
- **Fallback & resilience** — retry/circuit breaker อัตโนมัติ
