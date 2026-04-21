# HTTP API (Go)

Small service using the standard library:

- `GET /` → `Hello World!` (plain text)
- `GET /uuid` → UUID v4 ใหม่ทุกครั้ง (ใช้ `github.com/google/uuid`)
- `GET /health` → `{"status":"ok"}` (JSON)
- `GET /users` → ดึง user ทั้งหมดจาก PostgreSQL
- `GET /users/:id` → ดึง user รายคนตาม UUID
- `POST /users` → สร้าง user ใน PostgreSQL local
- `PUT /users/:id` → อัปเดต user แบบเต็ม (ต้องส่งครบทุกฟิลด์)
- `PATCH /users/:id` → อัปเดต user แบบบางส่วน (ส่งเฉพาะฟิลด์ที่อยากแก้)

## Run with nodemon (แนะนำตอนพัฒนา)

รีโหลดเมื่อแก้ไฟล์ `.go` — ต้องมี Node.js แล้วรัน `npm install` ครั้งแรก:

```bash
npm install
npm start
```

(`npm run dev` ทำงานเหมือนกัน)

## Run แบบ Go ล้วน (ไม่มีรีโหลด)

หยุดแอปอื่นที่ใช้พอร์ต 3000 ก่อน (เช่น Next.js) แล้ว:

```bash
npm run go
```

หรือ `go run ./cmd/server` / `make run`

Optional: `PORT=8080 npm start` ถ้าพอร์ต 3000 ถูกใช้อยู่

## PostgreSQL local setup

รองรับ 2 แบบ:

1) ตั้ง `DATABASE_URL` ตรงๆ

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
```

2) ตั้งค่าทีละตัว

```bash
export PGHOST=localhost
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=postgres
export PGDATABASE=postgres
export PGSSLMODE=disable
```

เซิร์ฟเวอร์จะสร้างตาราง `users` ให้อัตโนมัติถ้ายังไม่มี
- `id` จะถูกสร้างอัตโนมัติจาก PostgreSQL (`gen_random_uuid()`) ถ้าไม่ส่งค่าเข้ามา

ยิงสร้าง user ตัวอย่าง:

```bash
curl -X POST "http://localhost:8080/users" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "6d4cbf4e-94f5-4e0f-af4f-b5f96ce267b9",
    "firstname": "Bee",
    "lastname": "Coder"
  }'
```

หรือส่งแค่ชื่อ/นามสกุลให้ DB สร้าง `id` เอง:

```bash
curl -X POST "http://localhost:8080/users" \
  -H "Content-Type: application/json" \
  -d '{
    "firstname": "Bee",
    "lastname": "Coder"
  }'
```

ยิงอ่าน user ทั้งหมด:

```bash
curl -X GET "http://localhost:8080/users"
```

ยิงอ่าน user รายคน:

```bash
curl -X GET "http://localhost:8080/users/6d4cbf4e-94f5-4e0f-af4f-b5f96ce267b9"
```

ยิงอัปเดต user:

```bash
curl -X PUT "http://localhost:8080/users/6d4cbf4e-94f5-4e0f-af4f-b5f96ce267b9" \
  -H "Content-Type: application/json" \
  -d '{
    "firstname": "Bee Updated",
    "lastname": "Coder Updated"
  }'
```

หรือใช้ `PATCH` ได้เหมือนกัน:

```bash
curl -X PATCH "http://localhost:8080/users/6d4cbf4e-94f5-4e0f-af4f-b5f96ce267b9" \
  -H "Content-Type: application/json" \
  -d '{
    "firstname": "Bee Updated"
  }'
```

> ถ้าเรียก `PUT /users` โดยไม่ใส่ `:id` ระบบจะตอบ `400` และให้ใช้ `/users/:id` แทน

## Test

```bash
go test ./...
```

## Docker

```bash
docker build -t api .
docker run --rm -p 3000:3000 api
```
