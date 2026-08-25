# Mitra Link API

Dokumen ini mengikuti kode aktual di repo `AgroLink_API` per 2026-08-12. Semua endpoint di bawah berada di prefix `/api/v1` kecuali webhook Midtrans yang ada di `/api/webhooks/...` dan dipakai sebagai pemicu otomatis status pembayaran.

## Ringkasan Untuk Frontend

Mitra Link adalah alur kemitraan antara Petani dan Mitra/Investor dengan escrow Midtrans.

- Petani dapat mengajukan kerja sama ke Mitra.
- Mitra dapat menawarkan kerja sama ke Petani.
- Admin memverifikasi profil Mitra dan melepaskan dana setelah escrow masuk.

### Alur Status

1. `submitted`
   - Dibuat saat `POST /api/v1/cooperations/offer` atau `POST /api/v1/cooperations/apply`.
   - Pemicu: user action dari pihak pengaju.
2. `reviewed`
   - Dibuat saat pihak penerima mengirim `POST /api/v1/cooperations/:id/review`.
   - Pemicu: user action dari pihak penerima pengajuan.
3. `waiting_payment`
   - Dibuat saat pihak penerima mengirim `POST /api/v1/cooperations/:id/approve`.
   - Saat ini invoice otomatis dibuat pada transaksi yang sama.
4. `escrowed`
   - Dibuat otomatis oleh webhook Midtrans setelah pembayaran sukses terkonfirmasi.
   - Pemicu: sistem otomatis via webhook `POST /api/webhooks/midtrans-notification`.
5. `contract_generated`
   - Dibuat otomatis pada langkah webhook yang sama setelah contract `mitra` digenerate dan di-link ke cooperation.
   - Pemicu: sistem otomatis via webhook.
6. `completed`
   - Dibuat saat admin memanggil `POST /api/v1/admin/cooperations/:id/release`.
   - Pemicu: user action admin.
7. `rejected`
   - Dibuat saat pihak penerima menolak kerja sama lewat `POST /api/v1/cooperations/:id/reject`.
   - Pemicu: user action dari pihak penerima.

Catatan: enum model masih memuat `cancelled`, tetapi tidak ada endpoint di kode saat ini yang mengubah cooperation menjadi `cancelled`.

### Kenapa `agreed_amount` Bisa Berbeda dari `proposed_amount`

`proposed_amount` adalah angka awal yang diajukan oleh pengirim proposal. Saat approval, pihak penerima boleh mengisi `agreed_amount` sebagai nilai final hasil negosiasi. Di service, kalau `agreed_amount` tidak diisi atau nilainya tidak valid, kode akan memakai `proposed_amount` sebagai nilai final.

### Kenapa `Amount` di Invoice Berbeda dari `TotalAmount`

Pada cooperation yang sudah di-approve:

- `invoice.TotalAmount` = nilai bruto yang dibayar lewat Midtrans.
- `invoice.PlatformFee` = 11% dari nilai final.
- `invoice.Amount` = nilai netto setelah fee, yaitu `TotalAmount - PlatformFee`.

Di template kontrak Mitra Link, nilai netto juga ditulis sebagai `agreed_amount * 0.89`.

### Field Yang Perlu Ditampilkan Di UI

- Untuk layar list/detail cooperation, frontend hanya menerima data party brief: `user_id`, `name`, `email`, `phone`. Tidak ada field rekening bank di response cooperation.
- Untuk layar profil Mitra, response `MitraProfileResponse` memang membawa field sensitif seperti `npwp`, `nib`, `nik_ktp`, `dokumen_legalitas`, `nama_bank`, `nomor_rekening`, dan `atas_nama_rekening`. Tampilkan hanya di layar profil pemilik atau layar verifikasi admin jika memang dibutuhkan.
- Untuk layar kontrak, response list kontrak hanya berisi ringkasan: `contract_id`, `contract_type`, `title`, `status`, dan `offered_at`.
- `GET /api/v1/mitra` dan `GET /api/v1/mitra/:id` mengembalikan semua field profil Mitra yang ada di DTO. Jika frontend butuh kartu publik yang lebih aman, masking dilakukan di UI, bukan di backend saat ini.

### Envelope Response

Handler Mitra Link memakai `utils.SuccessResponse` dan `utils.ErrorResponse`:

```json
{
  "status": "success",
  "message": "...",
  "data": {}
}
```

```json
{
  "status": "error",
  "message": "...",
  "error": "..."
}
```

Catatan tambahan: bila request ditolak middleware role, response bisa memakai format `{"success":false,"error":"..."}` karena middleware memanggil helper `utils.Forbidden`.

## Mitra Profile

### GET /api/v1/mitra

List semua Mitra yang sudah diverifikasi.

- Auth: wajib login.
- Role: semua role yang sudah terautentikasi, tidak ada `RoleMiddleware` khusus.
- Query string: `page`, `limit`, `sort`.
- Default pagination: `page=1`, `limit=10`, `sort=created_at desc`.

#### Success 200

```json
{
  "status": "success",
  "message": "Verified mitra list fetched successfully",
  "data": {
    "data": [
      {
        "user_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
        "jenis_mitra": "perusahaan",
        "nama_mitra": "PT Sumber Pangan Nusantara",
        "deskripsi_singkat": "Investasi rantai pasok pertanian organik.",
        "nomor_telepon_bisnis": "081234567890",
        "email_bisnis": "halo@sumberpangan.co.id",
        "website": "https://sumberpangan.co.id",
        "alamat_lengkap": "Jl. Raya Bogor KM 27 No. 18, Depok",
        "provinsi": "Jawa Barat",
        "kota_kabupaten": "Kota Depok",
        "npwp": "12.345.678.9-012.000",
        "nib": "8123456789012",
        "nik_ktp": null,
        "dokumen_legalitas": "https://cdn.goagrolink.com/legal/mitra-spn.pdf",
        "status_verifikasi": "verified",
        "nama_bank": "BCA",
        "nomor_rekening": "1234567890",
        "atas_nama_rekening": "PT Sumber Pangan Nusantara",
        "logo_mitra": "https://cdn.goagrolink.com/logo/spn.png",
        "rating_mitra": 4.75,
        "total_transaksi_berhasil": 18,
        "created_at": "2026-08-01T09:15:30Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10,
    "total_pages": 1
  }
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "Failed to fetch verified mitra list",
  "error": "database timeout"
}
```

### GET /api/v1/mitra/:id

Ambil detail Mitra berdasarkan `user_id` Mitra, bukan UUID profil terpisah.

- Auth: wajib login.
- Role: semua role yang sudah terautentikasi.
- Path param `id`: UUID `users.id` milik Mitra.

#### Success 200

```json
{
  "status": "success",
  "message": "Mitra profile details fetched successfully",
  "data": {
    "user_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
    "jenis_mitra": "perusahaan",
    "nama_mitra": "PT Sumber Pangan Nusantara",
    "deskripsi_singkat": "Investasi rantai pasok pertanian organik.",
    "nomor_telepon_bisnis": "081234567890",
    "email_bisnis": "halo@sumberpangan.co.id",
    "website": "https://sumberpangan.co.id",
    "alamat_lengkap": "Jl. Raya Bogor KM 27 No. 18, Depok",
    "provinsi": "Jawa Barat",
    "kota_kabupaten": "Kota Depok",
    "npwp": "12.345.678.9-012.000",
    "nib": "8123456789012",
    "nik_ktp": null,
    "dokumen_legalitas": "https://cdn.goagrolink.com/legal/mitra-spn.pdf",
    "status_verifikasi": "verified",
    "nama_bank": "BCA",
    "nomor_rekening": "1234567890",
    "atas_nama_rekening": "PT Sumber Pangan Nusantara",
    "logo_mitra": "https://cdn.goagrolink.com/logo/spn.png",
    "rating_mitra": 4.75,
    "total_transaksi_berhasil": 18,
    "created_at": "2026-08-01T09:15:30Z"
  }
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "Invalid mitra ID format",
  "error": "invalid UUID length: 4"
}
```

Kalau `id` valid tapi profil tidak ditemukan, response menjadi 404 dengan pesan dari service.

### POST /api/v1/mitra/profile

Buat profil Mitra milik user yang sedang login.

- Auth: wajib login.
- Role: `mitra`.
- Body JSON: wajib valid.
- Aturan tambahan dari service:
  - `jenis_mitra = perusahaan` atau `organisasi` wajib mengisi `nib`.
  - `jenis_mitra = individu` wajib mengisi `nik_ktp`.
  - Jika profil sudah ada, request ditolak.

#### Request Body Contoh

```json
{
  "jenis_mitra": "perusahaan",
  "nama_mitra": "PT Sumber Pangan Nusantara",
  "deskripsi_singkat": "Investasi rantai pasok pertanian organik dan greenhouse.",
  "nomor_telepon_bisnis": "081234567890",
  "email_bisnis": "halo@sumberpangan.co.id",
  "website": "https://sumberpangan.co.id",
  "alamat_lengkap": "Jl. Raya Bogor KM 27 No. 18, Depok",
  "provinsi": "Jawa Barat",
  "kota_kabupaten": "Kota Depok",
  "npwp": "12.345.678.9-012.000",
  "nib": "8123456789012",
  "nik_ktp": null,
  "dokumen_legalitas": "https://cdn.goagrolink.com/legal/mitra-spn.pdf",
  "nama_bank": "BCA",
  "nomor_rekening": "1234567890",
  "atas_nama_rekening": "PT Sumber Pangan Nusantara",
  "logo_mitra": "https://cdn.goagrolink.com/logo/spn.png"
}
```

#### Success 201

```json
{
  "status": "success",
  "message": "Mitra profile created successfully",
  "data": {
    "user_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
    "jenis_mitra": "perusahaan",
    "nama_mitra": "PT Sumber Pangan Nusantara",
    "deskripsi_singkat": "Investasi rantai pasok pertanian organik dan greenhouse.",
    "nomor_telepon_bisnis": "081234567890",
    "email_bisnis": "halo@sumberpangan.co.id",
    "website": "https://sumberpangan.co.id",
    "alamat_lengkap": "Jl. Raya Bogor KM 27 No. 18, Depok",
    "provinsi": "Jawa Barat",
    "kota_kabupaten": "Kota Depok",
    "npwp": "12.345.678.9-012.000",
    "nib": "8123456789012",
    "nik_ktp": null,
    "dokumen_legalitas": "https://cdn.goagrolink.com/legal/mitra-spn.pdf",
    "status_verifikasi": "pending",
    "nama_bank": "BCA",
    "nomor_rekening": "1234567890",
    "atas_nama_rekening": "PT Sumber Pangan Nusantara",
    "logo_mitra": "https://cdn.goagrolink.com/logo/spn.png",
    "rating_mitra": 0,
    "total_transaksi_berhasil": 0,
    "created_at": "2026-08-12T09:15:30Z"
  }
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "NIB wajib diisi untuk jenis mitra perusahaan/organisasi",
  "error": "NIB wajib diisi untuk jenis mitra perusahaan/organisasi"
}
```

Jika request dipanggil role selain `mitra`, middleware mengembalikan 403.

### GET /api/v1/mitra/profile/my

Ambil profil Mitra milik user yang sedang login.

- Auth: wajib login.
- Role: `mitra`.

#### Success 200

```json
{
  "status": "success",
  "message": "Mitra profile retrieved successfully",
  "data": {
    "user_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
    "jenis_mitra": "perusahaan",
    "nama_mitra": "PT Sumber Pangan Nusantara",
    "deskripsi_singkat": "Investasi rantai pasok pertanian organik dan greenhouse.",
    "nomor_telepon_bisnis": "081234567890",
    "email_bisnis": "halo@sumberpangan.co.id",
    "website": "https://sumberpangan.co.id",
    "alamat_lengkap": "Jl. Raya Bogor KM 27 No. 18, Depok",
    "provinsi": "Jawa Barat",
    "kota_kabupaten": "Kota Depok",
    "npwp": "12.345.678.9-012.000",
    "nib": "8123456789012",
    "nik_ktp": null,
    "dokumen_legalitas": "https://cdn.goagrolink.com/legal/mitra-spn.pdf",
    "status_verifikasi": "verified",
    "nama_bank": "BCA",
    "nomor_rekening": "1234567890",
    "atas_nama_rekening": "PT Sumber Pangan Nusantara",
    "logo_mitra": "https://cdn.goagrolink.com/logo/spn.png",
    "rating_mitra": 4.75,
    "total_transaksi_berhasil": 18,
    "created_at": "2026-08-01T09:15:30Z"
  }
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "profil mitra belum dibuat",
  "error": "profil mitra belum dibuat"
}
```

## Admin — Verifikasi Mitra

### GET /api/v1/admin/mitra/pending-verification

Ambil daftar profil Mitra yang masih `pending`.

- Auth: wajib login.
- Role: `admin`.
- Query string: `page`, `limit`, `sort`.
- Default sorting dari service: `created_at asc`.

#### Success 200

```json
{
  "status": "success",
  "message": "Pending mitra verifications fetched successfully",
  "data": {
    "data": [
      {
        "user_id": "a7f11de2-7d5f-4aa6-8aa9-1b2e6fba9e4f",
        "jenis_mitra": "individu",
        "nama_mitra": "Budi Santoso",
        "deskripsi_singkat": "Investor lokal untuk program hidroponik.",
        "nomor_telepon_bisnis": "081298765432",
        "email_bisnis": "budi.santoso@example.com",
        "website": null,
        "alamat_lengkap": "Jl. Melati No. 21, Sleman",
        "provinsi": "DI Yogyakarta",
        "kota_kabupaten": "Sleman",
        "npwp": null,
        "nib": null,
        "nik_ktp": "3276011701900001",
        "dokumen_legalitas": "https://cdn.goagrolink.com/legal/budi-ktp.pdf",
        "status_verifikasi": "pending",
        "nama_bank": "BRI",
        "nomor_rekening": "123456789012345",
        "atas_nama_rekening": "Budi Santoso",
        "logo_mitra": null,
        "rating_mitra": 0,
        "total_transaksi_berhasil": 0,
        "created_at": "2026-08-10T11:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10,
    "total_pages": 1
  }
}
```

#### Error umum

```json
{
  "success": false,
  "error": "You do not have permission to access this resource"
}
```

### POST /api/v1/admin/mitra/:id/verify

Review verifikasi profil Mitra.

- Auth: wajib login.
- Role: `admin`.
- Path param `id`: UUID `users.id` milik Mitra.
- Body JSON wajib berisi `status` dengan nilai `verified` atau `rejected`.
- `notes` opsional.

#### Request Body Contoh

```json
{
  "status": "verified",
  "notes": "Dokumen legalitas dan rekening bank sudah sesuai."
}
```

#### Success 200

```json
{
  "status": "success",
  "message": "Mitra verification reviewed successfully",
  "data": null
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "Invalid input data",
  "error": "Key: 'ReviewMitraVerificationRequest.Status' Error:Field validation for 'Status' failed on the 'required' tag"
}
```

## Cooperation / Kerjasama

### POST /api/v1/cooperations/offer

Mitra membuat penawaran ke Petani.

- Auth: wajib login.
- Role: `mitra`.
- Aturan service:
  - profil Mitra harus sudah `verified`.
  - target `farmer_id` harus ada dan ber-role `farmer`.
  - status awal `submitted`.
  - `agreed_amount` awal diisi sama dengan `proposed_amount`.

#### Request Body Contoh

```json
{
  "farmer_id": "5c5fbc8f-2f8b-44f0-b3b1-7f4afcd2e0f2",
  "title": "Investasi Greenhouse Cabai Rawit",
  "description": "Pendanaan untuk 3 greenhouse cabai rawit di lahan 2 hektare dengan sistem irigasi tetes.",
  "proposed_amount": 250000000,
  "start_date": "2026-09-01T00:00:00Z",
  "end_date": "2027-02-28T00:00:00Z",
  "notes": "Fokus pada produksi panen tahap pertama."
}
```

#### Success 201

```json
{
  "status": "success",
  "message": "Cooperation offer created successfully",
  "data": {
    "id": "6e3b6a72-9e4a-42c6-bae8-5f9391f7d9e2",
    "mitra_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
    "farmer_id": "5c5fbc8f-2f8b-44f0-b3b1-7f4afcd2e0f2",
    "initiator_type": "mitra",
    "title": "Investasi Greenhouse Cabai Rawit",
    "description": "Pendanaan untuk 3 greenhouse cabai rawit di lahan 2 hektare dengan sistem irigasi tetes.",
    "proposed_amount": 250000000,
    "agreed_amount": 250000000,
    "platform_fee_percentage": 11,
    "start_date": "2026-09-01T00:00:00Z",
    "end_date": "2027-02-28T00:00:00Z",
    "status": "submitted",
    "notes": "Fokus pada produksi panen tahap pertama.",
    "contract_id": null,
    "mitra": {
      "user_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
      "name": "PT Sumber Pangan Nusantara",
      "email": "halo@sumberpangan.co.id",
      "phone": "081234567890"
    },
    "farmer": {
      "user_id": "5c5fbc8f-2f8b-44f0-b3b1-7f4afcd2e0f2",
      "name": "Pak Agus Pratama",
      "email": "agus.pratama@example.com",
      "phone": "081355112233"
    },
    "created_at": "2026-08-12T09:30:00Z",
    "updated_at": "2026-08-12T09:30:00Z"
  }
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "akun Mitra belum terverifikasi oleh Admin",
  "error": "akun Mitra belum terverifikasi oleh Admin"
}
```

### POST /api/v1/cooperations/apply

Petani mengajukan kerja sama ke Mitra.

- Auth: wajib login.
- Role: `farmer`.
- Aturan service:
  - target Mitra harus sudah `verified`.
  - status awal `submitted`.

#### Request Body Contoh

```json
{
  "mitra_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
  "title": "Permintaan Pendanaan Padi Organik",
  "description": "Kami membutuhkan pendanaan untuk pembelian benih, pupuk organik, dan alat tanam untuk 5 hektare sawah.",
  "proposed_amount": 120000000,
  "start_date": "2026-09-10T00:00:00Z",
  "end_date": "2027-01-31T00:00:00Z",
  "notes": "Bisa dinegosiasikan dalam kisaran 10-15%."
}
```

#### Success 201

```json
{
  "status": "success",
  "message": "Cooperation application created successfully",
  "data": {
    "id": "8b8f38d0-2ad2-4a7f-8b7a-4d8c220f7777",
    "mitra_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
    "farmer_id": "5c5fbc8f-2f8b-44f0-b3b1-7f4afcd2e0f2",
    "initiator_type": "farmer",
    "title": "Permintaan Pendanaan Padi Organik",
    "description": "Kami membutuhkan pendanaan untuk pembelian benih, pupuk organik, dan alat tanam untuk 5 hektare sawah.",
    "proposed_amount": 120000000,
    "agreed_amount": 120000000,
    "platform_fee_percentage": 11,
    "start_date": "2026-09-10T00:00:00Z",
    "end_date": "2027-01-31T00:00:00Z",
    "status": "submitted",
    "notes": "Bisa dinegosiasikan dalam kisaran 10-15%.",
    "contract_id": null,
    "mitra": {
      "user_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
      "name": "PT Sumber Pangan Nusantara",
      "email": "halo@sumberpangan.co.id",
      "phone": "081234567890"
    },
    "farmer": {
      "user_id": "5c5fbc8f-2f8b-44f0-b3b1-7f4afcd2e0f2",
      "name": "Pak Agus Pratama",
      "email": "agus.pratama@example.com",
      "phone": "081355112233"
    },
    "created_at": "2026-08-12T09:40:00Z",
    "updated_at": "2026-08-12T09:40:00Z"
  }
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "mitra tujuan belum terverifikasi oleh Admin",
  "error": "mitra tujuan belum terverifikasi oleh Admin"
}
```

### GET /api/v1/cooperations/my

Daftar kerja sama milik user login. Hasilnya berbeda tergantung role.

- Auth: wajib login.
- Role: `farmer` atau `mitra`.
- Query string: `page`, `limit`, `sort`.

#### Success 200

```json
{
  "status": "success",
  "message": "Cooperations fetched successfully",
  "data": {
    "data": [
      {
        "id": "6e3b6a72-9e4a-42c6-bae8-5f9391f7d9e2",
        "title": "Investasi Greenhouse Cabai Rawit",
        "initiator_type": "mitra",
        "status": "waiting_payment",
        "proposed_amount": 250000000,
        "agreed_amount": 230000000,
        "mitra": {
          "user_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
          "name": "PT Sumber Pangan Nusantara",
          "email": "halo@sumberpangan.co.id",
          "phone": "081234567890"
        },
        "farmer": {
          "user_id": "5c5fbc8f-2f8b-44f0-b3b1-7f4afcd2e0f2",
          "name": "Pak Agus Pratama",
          "email": "agus.pratama@example.com",
          "phone": "081355112233"
        },
        "created_at": "2026-08-12T09:30:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10,
    "total_pages": 1
  }
}
```

#### Error umum

```json
{
  "success": false,
  "error": "You do not have permission to access this resource"
}
```

### GET /api/v1/cooperations/:id

Detail kerja sama untuk pihak yang terlibat.

- Auth: wajib login.
- Role: `farmer` atau `mitra`.
- Path param `id`: UUID cooperation.

#### Success 200

```json
{
  "status": "success",
  "message": "Cooperation details fetched successfully",
  "data": {
    "id": "6e3b6a72-9e4a-42c6-bae8-5f9391f7d9e2",
    "mitra_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
    "farmer_id": "5c5fbc8f-2f8b-44f0-b3b1-7f4afcd2e0f2",
    "initiator_type": "mitra",
    "title": "Investasi Greenhouse Cabai Rawit",
    "description": "Pendanaan untuk 3 greenhouse cabai rawit di lahan 2 hektare dengan sistem irigasi tetes.",
    "proposed_amount": 250000000,
    "agreed_amount": 230000000,
    "platform_fee_percentage": 11,
    "start_date": "2026-09-01T00:00:00Z",
    "end_date": "2027-02-28T00:00:00Z",
    "status": "waiting_payment",
    "notes": "Setuju dengan penyesuaian harga peralatan.",
    "contract_id": null,
    "mitra": {
      "user_id": "b2a84dd8-3b3b-4c6d-8f5a-2fcb6d71f7f1",
      "name": "PT Sumber Pangan Nusantara",
      "email": "halo@sumberpangan.co.id",
      "phone": "081234567890"
    },
    "farmer": {
      "user_id": "5c5fbc8f-2f8b-44f0-b3b1-7f4afcd2e0f2",
      "name": "Pak Agus Pratama",
      "email": "agus.pratama@example.com",
      "phone": "081355112233"
    },
    "created_at": "2026-08-12T09:30:00Z",
    "updated_at": "2026-08-12T10:00:00Z"
  }
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "Akses ditolak: Anda bukan bagian dari kerjasama ini",
  "error": "Akses ditolak: Anda bukan bagian dari kerjasama ini"
}
```

### POST /api/v1/cooperations/:id/review

Tandai proposal sebagai `reviewed` dan simpan catatan negosiasi.

- Auth: wajib login.
- Role: `farmer` atau `mitra`.
- Hanya pihak penerima pengajuan yang boleh memanggil endpoint ini.
- Body JSON hanya berisi `notes` dan sifatnya opsional.

#### Request Body Contoh

```json
{
  "notes": "Proposal sudah ditinjau. Kami akan lanjut ke approval setelah revisi nominal."
}
```

#### Success 200

```json
{
  "status": "success",
  "message": "Cooperation status updated to reviewed",
  "data": null
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "hanya pihak penerima pengajuan yang dapat melakukan peninjauan",
  "error": "hanya pihak penerima pengajuan yang dapat melakukan peninjauan"
}
```

### POST /api/v1/cooperations/:id/approve

Setujui kerja sama dan buat invoice secara otomatis.

- Auth: wajib login.
- Role: `farmer` atau `mitra`.
- Hanya pihak penerima pengajuan yang boleh memanggil endpoint ini.
- Status cooperation harus `submitted` atau `reviewed`.
- `agreed_amount` opsional. Kalau tidak diisi atau bernilai tidak valid, sistem memakai `proposed_amount`.
- Invoice dibuat di transaksi yang sama dan cooperation pindah ke `waiting_payment`.

#### Request Body Contoh

```json
{
  "agreed_amount": 230000000,
  "start_date": "2026-09-05T00:00:00Z",
  "end_date": "2027-03-05T00:00:00Z",
  "notes": "Disetujui dengan pengurangan nominal 20 juta dan penyesuaian jadwal tanam."
}
```

#### Success 200

```json
{
  "status": "success",
  "message": "Cooperation approved successfully, invoice generated",
  "data": null
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "kerjasama tidak dapat disetujui pada status saat ini",
  "error": "kerjasama tidak dapat disetujui pada status saat ini"
}
```

### POST /api/v1/cooperations/:id/reject

Tolak kerja sama.

- Auth: wajib login.
- Role: `farmer` atau `mitra`.
- Hanya pihak penerima pengajuan yang boleh memanggil endpoint ini.
- Tidak bisa dipakai jika status sudah `escrowed`, `contract_generated`, atau `completed`.

#### Request Body Contoh

```json
{
  "notes": "Saat ini kapasitas pendanaan belum memungkinkan. Silakan ajukan ulang bulan depan."
}
```

#### Success 200

```json
{
  "status": "success",
  "message": "Cooperation rejected",
  "data": null
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "kerjasama yang sudah dibayar/selesai tidak dapat ditolak",
  "error": "kerjasama yang sudah dibayar/selesai tidak dapat ditolak"
}
```

### POST /api/v1/cooperations/:id/initiate-payment

Mitra memulai pembayaran Midtrans untuk invoice cooperation yang sudah di-approve.

- Auth: wajib login.
- Role: `mitra`.
- Hanya Mitra yang terdaftar sebagai `mitra_id` pada cooperation yang boleh memanggil endpoint ini.
- Cooperation harus berstatus `waiting_payment`.
- Response berisi `snap_token` dan `redirect_url` untuk Snap UI / redirect.

#### Success 200

```json
{
  "status": "success",
  "message": "Payment initiated successfully",
  "data": {
    "snap_token": "AbCdEfGhIjKlMnOpQrStUvWxYz1234567890",
    "order_id": "9a4c6b6e-0d12-4d61-a5bb-4d42b9f2f001",
    "amount": 230000000,
    "redirect_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/AbCdEfGhIjKlMnOpQrStUvWxYz1234567890"
  }
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "cooperation is not awaiting payment (current status: submitted)",
  "error": "cooperation is not awaiting payment (current status: submitted)"
}
```

### POST /api/v1/cooperations/:id/reviews

Buat review setelah cooperation selesai.

- Auth: wajib login.
- Role: `farmer` atau `mitra`.
- Hanya boleh jika status cooperation sudah `completed`.
- Satu reviewer hanya bisa review satu kali per cooperation.

#### Request Body Contoh

```json
{
  "rating": 5,
  "comment": "Proses kerja sama transparan, jadwal pembayaran jelas, dan komunikasi lancar."
}
```

#### Success 201

```json
{
  "status": "success",
  "message": "Review created successfully",
  "data": null
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "ulasan hanya dapat diberikan setelah kerjasama berstatus 'completed'",
  "error": "ulasan hanya dapat diberikan setelah kerjasama berstatus 'completed'"
}
```

## Admin — Disbursement

### POST /api/v1/admin/cooperations/:id/release

Admin melepaskan dana escrow ke Petani.

- Auth: wajib login.
- Role: `admin`.
- Path param `id`: UUID cooperation.
- Cooperation harus berstatus `escrowed` atau `contract_generated`.
- Invoice terkait harus sudah `paid`.
- Setelah sukses, cooperation menjadi `completed` dan `TotalTransaksiBerhasil` Mitra dinaikkan 1.

#### Success 200

```json
{
  "status": "success",
  "message": "Cooperation funds released to farmer successfully",
  "data": null
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "cooperation is not in escrowed or contract_generated state (current: waiting_payment)",
  "error": "cooperation is not in escrowed or contract_generated state (current: waiting_payment)"
}
```

## Kontrak

### GET /api/v1/contracts/my

Daftar kontrak milik user login. Route ini sudah mengizinkan `mitra` di middleware, jadi kontrak Mitra Link ikut muncul di sini.

- Auth: wajib login.
- Role: `worker`, `driver`, `mitra`, atau `farmer`.
- Response hanya summary list.

#### Success 200

```json
{
  "status": "success",
  "message": "Contracts retrieved successfully",
  "data": [
    {
      "contract_id": "3a9f9da6-2a4d-4f61-b1ef-2d3d1c7f8f11",
      "contract_type": "mitra",
      "title": "Kemitraan: Investasi Greenhouse Cabai Rawit",
      "status": "active",
      "offered_at": "2026-08-12T10:05:00Z"
    }
  ]
}
```

#### Error umum

```json
{
  "status": "error",
  "message": "Failed to retrieve contracts",
  "error": "database timeout"
}
```

### GET /api/v1/contracts/:id/download

Download PDF kontrak.

- Auth: wajib login karena route ada di protected group.
- Role middleware: tidak ada tambahan khusus di route ini, jadi aksesnya mengikuti auth login saja.
- Path param `id`: UUID kontrak.
- Untuk `contract_type = mitra`, service memakai template `templates/mitra_contract_template.html` dan PDF dihasilkan otomatis.
- Response bukan JSON, tetapi file PDF dengan header `Content-Disposition: attachment; filename=kontrak_<id>.pdf`.

#### Success 200

- `Content-Type`: `application/pdf`
- `Content-Disposition`: `attachment; filename=kontrak_<contract_id>.pdf`
- Body: stream PDF kontrak

#### Error umum

```json
{
  "status": "error",
  "message": "Failed to generate PDF",
  "error": "contract details not found"
}
```

## Yang Perlu Disiapkan Frontend

- Siapkan integrasi Midtrans Snap di sisi client memakai `snap_token` atau `redirect_url` dari response `POST /api/v1/cooperations/:id/initiate-payment`.
- Setelah user selesai membayar, lakukan polling atau refresh detail cooperation. Status `escrowed` dan `contract_generated` baru muncul setelah webhook Midtrans diproses async.
- Tampilkan tombol/aksi berdasarkan role:
  - `mitra`: buat profile, buat offer, initiate payment, lihat kontrak Mitra Link.
  - `farmer`: buat application, review/approve/reject proposal yang diterima, lihat kontrak Mitra Link, buat review setelah selesai.
  - `admin`: verifikasi Mitra dan release dana escrow.
- Untuk layar verifikasi admin, gunakan data dari `/admin/mitra/pending-verification` dan `POST /admin/mitra/:id/verify`.
- Untuk layar detail cooperation, gunakan field `status` sebagai sumber kebenaran UI, bukan state lokal setelah redirect Midtrans.
- Untuk download kontrak, gunakan `GET /api/v1/contracts/:id/download` dan render sebagai file PDF, bukan JSON.
- Karena `GET /api/v1/mitra` dan `GET /api/v1/mitra/:id` mengembalikan field sensitif juga, pastikan UI hanya menampilkan field yang relevan untuk layar yang sedang dibuka.
