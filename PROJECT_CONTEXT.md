
# 📄 Dokumen Konteks Proyek: AgroLink  

## 1. 🎯 Tujuan Proyek
AgroLink adalah **platform digital untuk sektor pertanian** yang menghubungkan berbagai peran seperti **petani, pekerja, ekspedisi, dan konsumen**. Tujuan utamanya:  
- Menyediakan **marketplace pertanian** yang aman dengan sistem autentikasi berbasis **JWT**.  
- Memfasilitasi **transaksi dan kolaborasi** antara aktor pertanian dengan transparansi dan akuntabilitas.  
- Menyediakan **layanan pendukung** seperti manajemen pekerja, logistik/ekspedisi, monitoring, serta sistem notifikasi.  
- Mendukung **penyimpanan file** (misalnya foto produk, dokumen transaksi) baik melalui **local storage** maupun **cloud storage** di masa depan.  

---

## 2. 🏗️ Arsitektur & Teknologi

### Arsitektur  
- **Backend**: Monolithic API berbasis **Go + Gin** dengan pola **layered architecture** (handlers → services → repositories → models).  
- **Autentikasi & Autorisasi**: JWT (JSON Web Token) dengan role-based access control (RBAC).  
- **Database**: MySQL/MariaDB menggunakan **GORM** sebagai ORM.  
- **Frontend (rencana)**: React.js.  
- **Deployment**: Bisa dijalankan secara lokal atau di cloud. File static disajikan melalui folder `public/`.  
- **Konfigurasi**: Environment variables via `.env` dengan fallback ke system environment.  

### Teknologi Utama  
- **Backend**: Go (Gin, GORM, godotenv).  
- **Auth**: JWT (configurable secret + expiry).  
- **Middleware**: CORS, JWT Auth Middleware.  
- **Database**: MySQL/MariaDB.  
- **File Upload**: Local storage (upload path: `./uploads`), dengan opsi cloud storage di masa depan.  
- **Notifikasi**: SMTP (Gmail) untuk email alert/notification.  

---

## 3. ⚙️ Detail Implementasi Fitur

### Infrastruktur & Konfigurasi
- `.env` untuk konfigurasi (APP_ENV, PORT, DB_HOST, DB_USER, JWT_SECRET, SMTP, UPLOAD_PATH, dll).  
- `config/` berisi **LoadConfig, ConnectDatabase, CloseDatabase, AutoMigrate**.  
- Graceful shutdown menggunakan `os.Signal` & `syscall.SIGTERM`.  

### Routing
- **Public routes**: register, login, akses umum.  
- **Protected routes**: hanya bisa diakses dengan JWT, middleware memastikan user valid.  
- Versi API: `/api/v1`.  
- Endpoint kesehatan: `/health`.  

### Middleware
- **CORS**: mengizinkan `http://localhost:3000` dan `http://localhost:8080`.  
- **AuthMiddleware**: validasi JWT & inject user ke context.  

### User Management
- **Repository** untuk CRUD user.  
- **Role-based model** (petani, pekerja, ekspedisi, dll).  
- Sub-model relational sesuai peran (misalnya farmer detail, worker profile, dsb).  

### File & Storage
- Static files disajikan melalui `/static`.  
- Upload path default: `./uploads`.  
- Ada catatan *TODO*: pertimbangkan migrasi ke **cloud storage** untuk skalabilitas.  

---

## 4. 🚧 Tantangan & Solusi  

### 🔑 Isu JWT_SECRET tidak terbaca
- **Masalah**: Muncul peringatan `JWT_SECRET tidak diatur. Menggunakan secret key default yang tidak aman.` meski sudah ada di `.env`.  
- **Solusi**:  
  - Pastikan `godotenv.Load()` membaca file `.env` dari direktori root.  
  - Tambahkan debug `log.Println(os.Getenv("JWT_SECRET"))`.  
  - Validasi penggunaan `getAndValidateEnv("JWT_SECRET")` di `config.LoadConfig()`.  

### 📂 Local Storage vs Cloud Storage
- **Diskusi**:  
  - Local storage → sederhana, cepat, tapi terbatas (scalability & backup).  
  - Cloud storage → lebih **reliable**, scalable, ada redundancy, cocok untuk production.  
- **Kesepakatan**: Saat ini pakai **local storage**, namun sistem disiapkan agar mudah migrasi ke **cloud storage** (misalnya AWS S3, GCP Storage, Supabase).  

### 🔒 UX dalam Auth & Route Separation
- **Masalah**: bagaimana memisahkan **public routes** & **protected routes** dengan user experience yang baik?  
- **Solusi**: dibuat grouping di Gin:  
  - `/api/v1/public` → tanpa JWT.  
  - `/api/v1/protected` → dengan JWT middleware.  
  - Hal ini memudahkan developer & frontend untuk mengkonsumsi API dengan jelas.  

### ⚙️ Konfigurasi & Deployment
- **Masalah**: `.env` tidak selalu terbaca jika file tidak ada di path yang sesuai.  
- **Solusi**: gunakan fallback ke **system environment variables** dan log peringatan jika `.env` tidak ditemukan.  

---

# ✅ Ringkasan
AgroLink sudah memiliki fondasi kuat dengan **Go (Gin, GORM, JWT)**, struktur modular, serta konfigurasi environment yang fleksibel. Fokus pengembangan berikutnya:  
1. Menyempurnakan **user role & profile management**.  
2. Menambahkan **cloud storage** untuk file upload.  
3. Mengintegrasikan notifikasi (email & mungkin WhatsApp/SMS).  
4. Menyiapkan pipeline deployment ke server/cloud.  
