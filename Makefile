.PHONY: run dev test

# รันแค่เซิร์ฟเวอร์ Go (พอร์ต 3000 หรือตั้ง PORT=...)
run:
	go run ./cmd/server

# รัน Go + รีโหลดเมื่อแก้ .go (nodemon)
dev:
	npm start

test:
	go test ./...
