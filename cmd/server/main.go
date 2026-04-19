package main

import (
	"fmt"
	"net/http"
	"os"
)

// หน้าแรก — ส่งข้อความ Hello World! กลับไปที่เบราว์เซอร์
func hello(w http.ResponseWriter, r *http.Request) {
	// ถ้าไม่ใช่ path / จริงๆ (เช่น /foo) ให้ตอบ 404
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Print(w, "Hello World!11")
}

// ตรวจว่าเซิร์ฟเวอร์ยังรันอยู่ — ใช้กับ health check ของ Docker / Render
func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok"}`)
}

func main() {
	// ลงทะเบียน URL กับฟังก์ชันที่จะรัน
	// ลงทะเบียน /health ก่อน / เพื่อไม่ให้ / ไปดักทุก path
	http.HandleFunc("/health", health)
	http.HandleFunc("/", hello)

	// พอร์ต: อ่านจากตัวแปรสภาพแวดล้อม PORT ถ้าไม่มีใช้ 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Println("เปิดเซิร์ฟเวอร์ที่พอร์ต", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("error:", err)
	}
}
