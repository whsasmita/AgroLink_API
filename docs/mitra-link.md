# TASK: Implementasi Fitur "Mitra Link" — Backend AgroLink_API (Go/Gin/GORM)

Kamu bekerja di repo `AgroLink_API` (Go, Gin, GORM, MySQL, Midtrans Snap, wkhtmltopdf untuk generate kontrak PDF).
Sebelum menulis kode apa pun, **baca dulu file-file berikut secara utuh** untuk memahami pola/konvensi yang sudah dipakai di project ini — semua kode baru WAJIB konsisten dengan pola ini:

- `models/user.go`, `models/project.go`, `models/contract.go`, `models/invoice.go`
- `middleware/auth.go`, `middleware/role.go`
- `utils/response.go`
- `repositories/project_repositories.go` (atau file repo lain yang mirip), `repositories/invoice_repositories.go`, `repositories/contract_repositories.go`
- `services/project_service.go`, `services/contract_service.go`, `services/payment_service.go`
- `handlers/project_handlers.go`
- `dto/project_response.go`, `dto/pagination_dto.go`
- `routes/protected_routes.go`
- `config/midtrans.go`, `config/migration.go`
- `templates/contract_template.html`

Pola yang harus diikuti:

- Layered architecture: `models -> repositories -> services -> handlers -> routes`, dengan dependency injection manual di `routes/protected_routes.go` (constructor `NewXxxRepository(db)`, `NewXxxService(...)`, `NewXxxHandler(...)`).
- Semua model punya `BeforeCreate` hook untuk generate UUID.
- Repository selalu punya parameter `tx *gorm.DB` untuk operasi tulis (`Create`, `Update`) agar bisa ikut dalam transaksi service; kalau `tx == nil`, fallback ke `r.db`.
- Response API pakai helper di `utils/response.go`: `utils.SuccessResponse(c, code, message, data)` dan `utils.ErrorResponse(c, code, message, err)`.
- Role-based access pakai `middleware.RoleMiddleware("role1", "role2", ...)`.
- Ambil user dari context lewat `c.Get("user")` lalu type assert ke `*models.User`.
- DTO request pakai `binding` tag Gin untuk validasi.

---

## 1. Konteks Bisnis

Agrolink sedang mengembangkan fitur **"Mitra Link"**: menghubungkan Petani (Farmer) dengan Mitra (Investor/Perusahaan/Organisasi/Individu) untuk investasi/kerjasama pertanian.

### Aturan Bisnis

- **Dua arah**: Mitra bisa langsung menawarkan kerjasama ke Petani (offer), dan sebaliknya Petani bisa mengajukan form kerjasama ke Mitra (application).
- **Escrow via Midtrans**: dana dari Mitra ditahan sistem. Platform memotong **11%** sebagai biaya layanan sebelum dana diteruskan ke Petani.
- **Lifecycle status**:
  1. `submitted` (diajukan)
  2. `reviewed` (ditinjau/dinegosiasikan oleh pihak penerima — opsional, bisa langsung approve dari `submitted`)
  3. `waiting_payment` (disetujui pihak penerima, menunggu Mitra membayar ke Midtrans)
  4. `escrowed` (dana sudah masuk & di-hold sistem, dikonfirmasi lewat webhook Midtrans)
  5. `contract_generated` (kontrak PDF otomatis dibuat dengan variabel dinamis kedua pihak)
  6. `completed` (dana sudah dicairkan ke Petani / disbursed)
  - Bisa juga berhenti di `rejected` atau `cancelled` sebelum `escrowed`.
- **Validasi & keamanan**: akun Mitra harus diverifikasi manual oleh Admin (verifikasi dokumen legalitas: NIB untuk perusahaan/organisasi, NIK KTP untuk individu) sebelum bisa membuat/menerima kerjasama.
- **Review dua arah**: setelah kerjasama `completed`, Petani dan Mitra bisa saling memberi rating & ulasan.

### Skema Database

**`mitra_profiles`** (relasi 1-1 ke `users`, PK `user_id` — ikuti pola `Farmer`/`Worker`/`Driver` yang sudah ada di `models/user.go`, BUKAN PK UUID terpisah):

```
user_id (PK, FK -> users.id)
jenis_mitra          enum('perusahaan','organisasi','individu')
nama_mitra           varchar(150)
deskripsi_singkat    text, nullable
nomor_telepon_bisnis varchar(20)
email_bisnis         varchar(100)
website               varchar(255), nullable
alamat_lengkap       text
provinsi             varchar(100)
kota_kabupaten       varchar(100)
npwp                 varchar(30), nullable
nib                  varchar(30), nullable  -- wajib jika jenis_mitra = perusahaan/organisasi
nik_ktp              varchar(20), nullable  -- wajib jika jenis_mitra = individu
dokumen_legalitas    text, nullable          -- path/URL file
status_verifikasi    enum('pending','verified','rejected') default 'pending'
nama_bank            varchar(50)
nomor_rekening       varchar(50)
atas_nama_rekening   varchar(100)
logo_mitra           text, nullable
rating_mitra         decimal(3,2) default 0
total_transaksi_berhasil int default 0
created_at, updated_at, deleted_at (soft delete)
```

**`mitra_cooperations`** (entitas transaksi/kerjasama, menampung dua arah lewat kolom `initiator_type`):

```
id                     UUID PK
mitra_id               FK -> users.id (lewat mitra_profiles.user_id)
farmer_id              FK -> users.id (lewat farmers.user_id)
initiator_type         enum('mitra','farmer')
title                  varchar(150)
description            text
proposed_amount        decimal(15,2)
agreed_amount          decimal(15,2), nullable
platform_fee_percentage decimal(5,2) default 11.00
start_date, end_date   date, nullable
status                 enum('submitted','reviewed','waiting_payment','escrowed','contract_generated','completed','rejected','cancelled') default 'submitted'
notes                  text, nullable
contract_id            UUID, nullable (FK -> contracts.id, diisi setelah kontrak digenerate)
created_at, updated_at
```

**`mitra_reviews`** (ulasan dua arah setelah `completed`):

```
id             UUID PK
cooperation_id FK -> mitra_cooperations.id
reviewer_type  enum('farmer','mitra')
reviewer_id    FK -> users.id
rating         int (1-5)
comment        text, nullable
created_at
```

---

## 2. Perubahan pada Model yang Sudah Ada

- **`models/user.go`**: tambahkan `'mitra'` ke enum `Role`, tambahkan relasi `Mitra *MitraProfile` (pola sama seperti `Farmer`, `Worker`, `Driver`).
- **`models/contract.go`**: tambahkan `'mitra'` ke enum `ContractType`. Tambahkan field `MitraCooperationID *uuid.UUID` dan `MitraID *uuid.UUID`, plus relasi `MitraCooperation *MitraCooperation` dan `Mitra *MitraProfile`.
- **`models/invoice.go`**: tambahkan field `MitraCooperationID *uuid.UUID` (nullable, seperti pola `ProjectID`/`DeliveryID` yang sudah ada) beserta relasinya.
- **`repositories/invoice_repositories.go`**: tambahkan method `FindByMitraCooperationID(cooperationID string) (*models.Invoice, error)`. **Sekalian perbaiki bug**: method `Create(tx *gorm.DB, invoice *models.Invoice)` yang ada sekarang mengabaikan parameter `tx` dan selalu pakai `r.db` — perbaiki agar memakai `tx` jika tidak nil (penting karena invoice Mitra Link dibuat di dalam transaksi service).
- **`repositories/contract_repositories.go`**: tambahkan `Preload("MitraCooperation")`/`Preload("Mitra.User")` di `FindByID`/`FindByIDWithDetails`, dan tambahkan `mitra_id` ke kondisi `WHERE` di `FindByUserID` (saat ini hanya `worker_id OR driver_id`).

---

## 3. File Baru yang Harus Dibuat

1. **`models/mitra_profile.go`** — struct `MitraProfile` sesuai skema di atas + `BeforeCreate` jika perlu (PK bukan UUID auto-generate karena = `user_id`, jadi tidak perlu `BeforeCreate` UUID).
2. **`models/mitra_cooperation.go`** — struct `MitraCooperation` dan `MitraReview` sesuai skema di atas, masing-masing dengan `BeforeCreate` untuk generate UUID.
3. **`dto/mitra_dto.go`** — request DTO: `CreateMitraProfileRequest`, `ReviewMitraVerificationRequest`, `CreateOfferRequest` (Mitra→Farmer), `CreateApplicationRequest` (Farmer→Mitra), `ReviewCooperationRequest`, `CreateMitraReviewRequest`.
4. **`dto/mitra_response.go`** — response DTO: `MitraProfileResponse`, `CooperationPartyResponse`, `CooperationBriefResponse`, `CooperationDetailResponse`.
5. **`repositories/mitra_repositories.go`** — `MitraProfileRepository`, `MitraCooperationRepository`, `MitraReviewRepository` (interface + implementasi gorm), method-method CRUD + query yang relevan (`FindByUserID`, `FindAllVerified` dengan pagination, `FindPendingVerification`, `FindAllByFarmerID`, `FindAllByMitraID`, `UpdateStatus`, dst).
6. **`services/mitra_service.go`** — dua service:
   - `MitraProfileService`: `CreateProfile` (validasi NIB/NIK sesuai jenis_mitra, cegah profil duplikat), `GetMyProfile`, `FindByUserID`, `FindAllVerified`, `GetPendingVerifications`, `ReviewVerification` (admin).
   - `MitraCooperationService`: `CreateOffer`, `CreateApplication` (keduanya wajib cek mitra sudah `verified`), `FindMyCooperations` (beda query tergantung role), `FindByID`, `ReviewCooperation` (hanya pihak PENERIMA pengajuan yang boleh, cek lawan dari `initiator_type`), `ApproveCooperation` (hanya pihak penerima; generate `Invoice` dengan `MitraCooperationID`, `Amount` = agreed_amount dikurangi fee 11%, `TotalAmount` = agreed_amount penuh, status cooperation → `waiting_payment`, semua dalam 1 transaksi DB), `RejectCooperation`, `CreateReview` (hanya jika status `completed`).
7. **Modifikasi `services/payment_service.go`**:
   - Tambah dependency `coopRepo repositories.MitraCooperationRepository` dan `contractRepo repositories.ContractRepository` ke struct & constructor.
   - Tambah method `InitiateCooperationPayment(cooperationID string, mitraID uuid.UUID) (*dto.PaymentInitiationResponse, error)` — mirip `InitiateInvoicePayment` tapi ambil invoice lewat `FindByMitraCooperationID`, validasi `coop.MitraID == mitraID` dan status `waiting_payment`.
   - Di `HandleWebhookNotification`, tambahkan branch: kalau `invoice.MitraCooperationID != nil` setelah pembayaran sukses (`finalizeSuccess`), panggil method baru `finalizeCooperationEscrow(cooperationID)` yang: update status cooperation → `escrowed`, generate `models.Contract` baru (`ContractType: "mitra"`, isi `FarmerID`, `MitraID`, `MitraCooperationID`, `SignedByFarmer: true`, `SignedBySecondParty: true`, `Status: "active"`, `SignedAt: now` — auto-signed karena sudah lewat alur approve+escrow), lalu update `cooperation.ContractID` dan status → `contract_generated`. Semua dalam satu transaksi DB.
   - Tambah method `ReleaseCooperationFunds(cooperationID string, adminID uuid.UUID) error` — dipanggil ADMIN, validasi status `escrowed`/`contract_generated` dan invoice `paid`, buat `models.Payout{PayeeID: coop.FarmerID, PayeeType: "farmer", Amount: invoice.Amount}` (amount sudah dipotong fee 11%), lalu update status cooperation → `completed`.
8. **Modifikasi `services/contract_service.go`**: di `GenerateContractPDF`, tambahkan branch untuk `ContractType == "mitra"` yang pakai template `templates/mitra_contract_template.html` (buat helper `generateMitraContractPDF`). Di `SignContract`, untuk `ContractType == "mitra"` kembalikan error informatif bahwa kontrak Mitra Link auto-signed saat escrow (tidak perlu sign manual).
9. **`handlers/mitra_handlers.go`** — `MitraProfileHandler` dan `MitraCooperationHandler` yang membungkus service-service di atas, konsisten dengan pola `project_handlers.go` (ambil user dari context, validasi role kalau perlu, bind JSON, panggil service, balikan lewat `utils.SuccessResponse`/`utils.ErrorResponse`).
10. **`templates/mitra_contract_template.html`** — template HTML kontrak kerjasama (gaya sama seperti `contract_template.html`: A4, Times New Roman, struktur PIHAK PERTAMA (Petani) & PIHAK KEDUA (Mitra), pasal-pasal: objek kerjasama, jangka waktu, nilai kerjasama & skema pembayaran (sebutkan potongan biaya layanan 11%), hak & kewajiban, pengakhiran, penyelesaian perselisihan, tanda tangan elektronik dua pihak.
11. **Wiring di `routes/protected_routes.go`**: inisialisasi repo/service/handler baru, lalu daftarkan route group berikut:
    - `GET /mitra` (list mitra verified, publik untuk user login), `GET /mitra/:id` (detail)
    - `POST /mitra/profile` (role: `mitra`), `GET /mitra/profile/my` (role: `mitra`)
    - `POST /cooperations/offer` (role: `mitra`), `POST /cooperations/apply` (role: `farmer`)
    - `GET /cooperations/my`, `GET /cooperations/:id` (role: `farmer`,`mitra`)
    - `POST /cooperations/:id/review`, `POST /cooperations/:id/approve`, `POST /cooperations/:id/reject` (role: `farmer`,`mitra`)
    - `POST /cooperations/:id/initiate-payment` (role: `mitra`)
    - `POST /cooperations/:id/reviews` (role: `farmer`,`mitra`)
    - Di dalam group `admin` yang sudah ada: `GET /admin/mitra/pending-verification`, `POST /admin/mitra/:id/verify`, `POST /admin/cooperations/:id/release`
    - Tambahkan `"mitra"` ke `middleware.RoleMiddleware(...)` pada route `GET /contracts/my` yang sudah ada supaya Mitra juga bisa melihat kontraknya.

---

## 4. Tugas Tambahan (Jangan Dilewatkan)

1. **`config/migration.go`**: tambahkan `&models.MitraProfile{}`, `&models.MitraCooperation{}`, `&models.MitraReview{}` ke daftar `AutoMigrate(...)`.
2. **`services/auth_service.go`** (baca dulu isinya): pastikan proses registrasi/`Register` mengizinkan `role = "mitra"` sebagai pilihan valid saat sign up, sejajar dengan `farmer`/`worker`/`driver`.
3. Cek apakah ada tempat lain yang melakukan validasi/whitelist terhadap `models.User.Role` (misal di `dto` request registrasi dengan tag `oneof=...`) dan tambahkan `mitra` di sana juga.

---

## 5. Acceptance Criteria

- `go build ./...` berhasil tanpa error.
- Alur end-to-end bisa diikuti secara logis: Mitra daftar profil → Admin verifikasi → Mitra buat offer ATAU Farmer buat application → pihak penerima review/approve → invoice otomatis terbuat (fee 11% sudah terhitung benar: `Amount = agreed_amount - fee`, `TotalAmount = agreed_amount`) → Mitra initiate payment → webhook Midtrans sukses → status jadi `escrowed` lalu otomatis `contract_generated` dengan kontrak PDF bisa di-generate lewat endpoint `GET /contracts/:id/download` yang sudah ada → Admin release funds → status `completed` → kedua pihak bisa saling review.
- Semua endpoint baru mengikuti format response yang sama dengan endpoint lain di project ini (`utils.SuccessResponse`/`utils.ErrorResponse`).
- Tidak ada breaking change pada fitur Project/Contract/Delivery/Payment yang sudah ada — jalankan sanity check pada alur project (`CreateProject` → `Apply` → `Accept` → `SignContract` → `InitiateInvoicePayment` → webhook → `ReleaseProjectPayment`) untuk memastikan tidak rusak akibat perubahan di `payment_service.go`, `contract_service.go`, `invoice_repositories.go`, `contract_repositories.go`.

Setelah selesai, tampilkan ringkasan file yang dibuat/diubah dan jalankan `go vet ./...` untuk memastikan tidak ada masalah.
