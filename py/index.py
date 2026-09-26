import pandas as pd
import random
import json
from datetime import datetime, date, timedelta

def generate_transactions():
    random.seed(4242)

    # 1. Load users & products
    try:
        with open('seeders/users_seed.json', 'r', encoding='utf-8') as f:
            users = json.load(f)
    except FileNotFoundError:
        with open('./seeders/users_seed.json', 'r', encoding='utf-8') as f:
            users = json.load(f)

    for u in users:
        u['created_at_dt'] = datetime.strptime(u['CreatedAt'], "%Y-%m-%d %H:%M:%S")

    try:
        with open('seeders/products_seed.json', 'r', encoding='utf-8') as f:
            products = json.load(f)
    except FileNotFoundError:
        with open('./seeders/products_seed.json', 'r', encoding='utf-8') as f:
            products = json.load(f)

    for p in products:
        p['created_at_dt'] = datetime.strptime(p['CreatedAt'], "%Y-%m-%d %H:%M:%S")

    # Categorize users
    farmers_agri = [u for u in users if u['Role'] == 'farmer' and u.get('Type') == 'agriculture']
    farmers_live = [u for u in users if u['Role'] == 'farmer' and u.get('Type') == 'livestock']
    farmers_const = [u for u in users if u['Role'] == 'farmer' and u.get('Type') == 'construction']
    all_farmers = [u for u in users if u['Role'] == 'farmer']

    workers_agri = [u for u in users if u['Role'] == 'worker' and u.get('Skills') == ['Pertanian']]
    workers_live = [u for u in users if u['Role'] == 'worker' and u.get('Skills') == ['Peternakan']]
    workers_const = [u for u in users if u['Role'] == 'worker' and u.get('Skills') == ['Tukang Bangunan']]
    all_workers = [u for u in users if u['Role'] == 'worker']

    drivers = [u for u in users if u['Role'] == 'driver']
    mitras = [u for u in users if u['Role'] == 'mitra']
    generals = [u for u in users if u['Role'] == 'general']
    all_non_admin = [u for u in users if u['Role'] != 'admin']

    # Kriteria Layanan dan Komisi
    komisi_map = {
        "Pekerja": 0.08,
        "Ekspedisi": 0.11,
        "E-Commerce": 0.10,
        "Tukang": 0.08,
        "Peternak": 0.08,
        "Chatbot Premium": 1.0,
        "Kemitraan": 0.15
    }
    
    metode_bayar = ["QRIS", "DANA", "GoPay", "ShopeePay", "BCA", "BNI", "MANDIRI"]
    metode_weights = [25, 20, 20, 15, 8, 6, 6]

    komentar_pekerja_gagal = [
        "si ini males-malesan terus dan membangkang saat diarahkan di lahan",
        "gak dateng-dateng padahal udah janji dari pagi buta",
        "pekerja membatalkan sepihak karena ada upacara adat mendadak tanpa konfirmasi awal",
        "pekerja minta uang muka lebih dulu lalu tidak hadir di lokasi kebun",
        "hasil pengerjaan lahan tidak rapi dan tidak sesuai standar kesepakatan",
        "pekerja hanya datang setengah hari lalu pamit dan mangkir di hari berikutnya",
        "kondisi fisik pekerja kurang fit sehingga tidak sanggup kerja seharian di lapangan",
        "komunikasi sulit sekali, pekerja tidak bisa dihubungi saat hari pelaksanaan",
        "pekerja menolak mengerjakan tugas sesuai instruksi dan memilih pulang lebih awal",
        "pekerja terlambat lebih dari 3 jam dan berselisih paham saat di kebun",
        "pekerja membatalkan sepihak karena menerima tawaran kerja di tempat lain",
        "pengerjaan terhenti di tengah jalan karena pekerja tidak membawa peralatan sendiri"
    ]

    komentar_ekspedisi_gagal = [
        "sopir terlambat lebih dari 3 jam dari jadwal penjemputan hasil panen",
        "armada truk ekspedisi mogok di jalan saat menuju lokasi perkebunan",
        "kapasitas muatan kendaraan tidak mencukupi untuk volume hasil panen yang dijadwalkan",
        "driver tidak bisa dihubungi saat muatan sudah siap diangkut di gudang",
        "sopir membatalkan rute karena akses jalan ke kebun dinilai terlalu curam dan licin",
        "terjadi kerusakan armada pickup di tengah perjalanan sehingga pengiriman batal",
        "driver menolak membawa muatan karena ada selisih kesepakatan titik bongkar muat"
    ]

    komentar_ecommerce_gagal = [
        "stok panen komoditas tiba-tiba habis rusak terkena hama sebelum sempat dikirim",
        "pembeli mengajukan pembatalan karena salah memasukkan alamat titik pengiriman",
        "kualitas buah saat disortir tidak fresh sehingga penjual menolak meneruskan pesanan",
        "pembeli membatalkan pesanan karena estimasi waktu kirim dianggap terlalu lama",
        "petani belum sempat memanen komoditas tepat waktu sesuai jadwal pesanan",
        "pembayaran kadaluarsa karena pembeli tidak menyelesaikan transfer dalam batas waktu"
    ]

    komentar_kemitraan_gagal = [
        "pihak mitra belum menyepakati klausul skema bagi hasil pada draf final MoU kerjasama",
        "mitra membatalkan rencana kerjasama karena adanya penyesuaian alokasi anggaran tahunan"
    ]

    # Distribusi Task Sukses (404) & Gagal (89) = 493
    # Fase 1: Sept 2025 - Mei 2026 (218 trans: 175 Sukses, 43 Gagal)
    # HANYA ada 3 layanan: Pekerja (Pertanian), Ekspedisi, E-Commerce.
    # TIDAK BOLEH ADA: AI Premium, Kemitraan, Peternakan, Tukang Bangunan.
    p1_tasks = (
        [("Pekerja", "Pertanian", "Sukses")] * 112 +
        [("Ekspedisi", "-", "Sukses")] * 32 +
        [("E-Commerce", "-", "Sukses")] * 31 +
        [("Pekerja", "Pertanian", "Gagal")] * 20 +
        [("Ekspedisi", "-", "Gagal")] * 13 +
        [("E-Commerce", "-", "Gagal")] * 10
    )

    # Fase 2: Juni 2026 - 25 Sept 2026 (275 trans: 229 Sukses, 46 Gagal)
    # Muncul tambahan: Peternakan, Tukang Bangunan, Chatbot Premium, Kemitraan + lanjutan Pertanian, Ekspedisi, E-Commerce
    p2_tasks = (
        [("Pekerja", "Pertanian", "Sukses")] * 62 +
        [("Pekerja", "Peternakan", "Sukses")] * 55 +
        [("Pekerja", "Tukang Bangunan", "Sukses")] * 54 +
        [("Ekspedisi", "-", "Sukses")] * 17 +
        [("E-Commerce", "-", "Sukses")] * 17 +
        [("Chatbot Premium", "-", "Sukses")] * 21 +
        [("Kemitraan", "-", "Sukses")] * 3 +
        [("Pekerja", "Pertanian", "Gagal")] * 12 +
        [("Pekerja", "Peternakan", "Gagal")] * 10 +
        [("Pekerja", "Tukang Bangunan", "Gagal")] * 10 +
        [("Ekspedisi", "-", "Gagal")] * 7 +
        [("E-Commerce", "-", "Gagal")] * 6 +
        [("Kemitraan", "-", "Gagal")] * 1
    )

    p1_start = date(2025, 9, 1)
    p1_end = date(2026, 5, 31)
    p1_num_days = (p1_end - p1_start).days + 1

    p1_weights = []
    for d in range(p1_num_days):
        dt = p1_start + timedelta(days=d)
        m_base = {
            (2025, 9): 0.50,
            (2025, 10): 0.70,
            (2025, 11): 0.80,
            (2025, 12): 0.65,
            (2026, 1): 1.05,
            (2026, 2): 1.15,
            (2026, 3): 1.45,
            (2026, 4): 1.65,
            (2026, 5): 1.80,
        }[(dt.year, dt.month)]
        noise = random.uniform(0.6, 1.4)
        if random.random() < 0.20:
            w = 0.0
        else:
            w = m_base * noise
        p1_weights.append(w)

    p1_date_choices = [p1_start + timedelta(days=i) for i in range(p1_num_days)]
    selected_p1_dates = random.choices(p1_date_choices, weights=p1_weights, k=len(p1_tasks))
    selected_p1_dates.sort()

    p2_start = date(2026, 6, 1)
    p2_end = date(2026, 9, 25)
    p2_num_days = (p2_end - p2_start).days + 1

    p2_weights = []
    for d in range(p2_num_days):
        dt = p2_start + timedelta(days=d)
        m_base = {
            6: 2.3,
            7: 2.8,
            8: 2.1,
            9: 2.4,
        }[dt.month]
        noise = random.uniform(0.7, 1.3)
        if random.random() < 0.10:
            w = 0.0
        else:
            w = m_base * noise
        p2_weights.append(w)

    p2_date_choices = [p2_start + timedelta(days=i) for i in range(p2_num_days)]
    selected_p2_dates = random.choices(p2_date_choices, weights=p2_weights, k=len(p2_tasks))
    selected_p2_dates.sort()

    random.shuffle(p1_tasks)
    random.shuffle(p2_tasks)

    hours = list(range(7, 22))
    h_weights = [3, 6, 9, 10, 10, 8, 7, 8, 9, 8, 7, 5, 5, 3, 2]

    def make_timestamp(d):
        h = random.choices(hours, weights=h_weights)[0]
        m = random.randint(0, 59)
        s = random.randint(0, 59)
        return datetime(d.year, d.month, d.day, h, m, s)

    all_raw_items = []
    for task, d in zip(p1_tasks, selected_p1_dates):
        all_raw_items.append((task, make_timestamp(d), 1))

    for task, d in zip(p2_tasks, selected_p2_dates):
        if task[0] == "Kemitraan":
            valid_p2_dates = [x for x in p2_date_choices if x >= date(2026, 7, 1)]
            kem_date = random.choice(valid_p2_dates)
            all_raw_items.append((task, make_timestamp(kem_date), 2))
        else:
            all_raw_items.append((task, make_timestamp(d), 2))

    all_raw_items.sort(key=lambda x: x[1])

    keterangan_pertanian = [
        "Jasa Olah Lahan & Penyiapan Bedengan",
        "Jasa Tanam Padi & Penyiangan Gulma",
        "Jasa Pemupukan & Perawatan Padi",
        "Jasa Panen Padi & Perontokan Gabah",
        "Jasa Pemeliharaan Kebun Hortikultura",
        "Jasa Perawatan Kebun Kopi & Kakao",
        "Jasa Pemangkasan & Pemupukan Jeruk",
        "Jasa Panen & Pemilahan Hasil Sayur"
    ]
    keterangan_peternakan = [
        "Jasa Pemeliharaan & Pemberian Pakan Ternak",
        "Jasa Pembersihan Kandang & Sanitasi Lingkungan",
        "Jasa Pemeriksaan Kesehatan & Perawatan Ternak",
        "Jasa Penggemukan Sapi & Pengolahan Pakan Silase",
        "Jasa Perawatan Kandang Kambing & Unggas"
    ]
    keterangan_tukang = [
        "Jasa Pembuatan Saluran Irigasi Tersier Kebun",
        "Jasa Renovasi Gudang Penyimpanan Hasil Panen",
        "Jasa Pembuatan Kandang Ternak Permanen",
        "Jasa Pembangunan Rangka Green House Hortikultura",
        "Jasa Pemasangan Pagar Pengaman Lahan Pertanian"
    ]
    keterangan_ekspedisi = [
        "Pengiriman Komoditas Tani ke Denpasar",
        "Pengangkutan Pupuk Organik & Bibit Unggul",
        "Distribusi Sayur & Buah Segar Antar Kabupaten",
        "Pengiriman Kopi & Kakao Menuju Gudang Sentral",
        "Ekspedisi Panen Hortikultura Bedugul - Badung",
        "Pengangkutan Hasil Panen Jeruk Kintamani"
    ]
    keterangan_kemitraan = [
        "Kerjasama Kemitraan Pengadaan Sarana Produksi Pertanian",
        "Kontrak Kerjasama Serap Gabah & Hasil Panen B2B",
        "Kerjasama Investasi Pengembangan Fasilitas Pasca Panen Modern",
        "MoU Kemitraan Distribusi & Pemasaran Komoditas AgroLink"
    ]

    transactions = []

    for (layanan_type, sektor, status), ts, phase in all_raw_items:
        tgl_str = ts.strftime("%Y-%m-%d")
        tgl_dt = ts
        metode = random.choices(metode_bayar, weights=metode_weights)[0]
        
        rec = {}
        komentar = "-"
        
        if layanan_type == "Pekerja":
            layanan_name = "Pekerja"
            if sektor == "Peternakan":
                layanan_name = "Peternak"
                valid_farmers = [u for u in farmers_live if u['created_at_dt'] <= tgl_dt] or farmers_live
                valid_workers = [u for u in workers_live if u['created_at_dt'] <= tgl_dt] or workers_live
                keterangan = random.choice(keterangan_peternakan)
            elif sektor == "Tukang Bangunan":
                layanan_name = "Tukang"
                valid_farmers = [u for u in farmers_const if u['created_at_dt'] <= tgl_dt] or farmers_const
                valid_workers = [u for u in workers_const if u['created_at_dt'] <= tgl_dt] or workers_const
                keterangan = random.choice(keterangan_tukang)
            else: # Pertanian
                valid_farmers = [u for u in farmers_agri if u['created_at_dt'] <= tgl_dt] or farmers_agri
                valid_workers = [u for u in workers_agri if u['created_at_dt'] <= tgl_dt] or workers_agri
                keterangan = random.choice(keterangan_pertanian)
                
            farmer = random.choice(valid_farmers)
            worker = random.choice(valid_workers)
            
            if phase == 1:
                nominal = random.randint(140, 180) * 1000
            else:
                nominal = random.randint(95, 130) * 1000
            komisi = komisi_map[layanan_name]
            
            if status == "Gagal":
                komentar = random.choice(komentar_pekerja_gagal)
                
            rec.update({
                "Layanan": layanan_name,
                "Keterangan": keterangan,
                "FarmerEmail": farmer["Email"],
                "FarmerName": farmer["Nama"],
                "FarmerType": farmer.get("Type", ""),
                "FarmerCreatedAt": farmer["CreatedAt"],
                "WorkerEmail": worker["Email"],
                "WorkerName": worker["Nama"],
                "WorkerSkills": worker.get("Skills", []),
                "WorkerCreatedAt": worker["CreatedAt"],
                "PemberiKerjaEmail": farmer["Email"],
                "PekerjaEmail": worker["Email"],
            })
            
        elif layanan_type == "Ekspedisi":
            valid_farmers = [u for u in all_farmers if u['created_at_dt'] <= tgl_dt] or all_farmers
            valid_drivers = [u for u in drivers if u['created_at_dt'] <= tgl_dt] or drivers
            farmer = random.choice(valid_farmers)
            driver = random.choice(valid_drivers)
            
            if phase == 1:
                nominal = random.randint(135, 180) * 1000
            else:
                nominal = random.randint(80, 115) * 1000
            komisi = komisi_map["Ekspedisi"]
            keterangan = random.choice(keterangan_ekspedisi)
            
            if status == "Gagal":
                komentar = random.choice(komentar_ekspedisi_gagal)
                
            rec.update({
                "Layanan": "Ekspedisi",
                "Keterangan": keterangan,
                "FarmerEmail": farmer["Email"],
                "FarmerName": farmer["Nama"],
                "FarmerType": farmer.get("Type", ""),
                "FarmerCreatedAt": farmer["CreatedAt"],
                "DriverEmail": driver["Email"],
                "DriverName": driver["Nama"],
                "DriverCreatedAt": driver["CreatedAt"],
                "PemberiKerjaEmail": farmer["Email"],
            })
            
        elif layanan_type == "E-Commerce":
            valid_products = [p for p in products if p['created_at_dt'] <= tgl_dt] or products
            prod = random.choice(valid_products)
            
            valid_buyers = [u for u in (generals + all_farmers + all_workers) if u['created_at_dt'] <= tgl_dt] or all_non_admin
            buyer = random.choice(valid_buyers)
            
            if phase == 1:
                qty = random.randint(3, 6)
                nominal = int(prod['Price']) * qty
                while nominal < 85000:
                    qty += 1
                    nominal = int(prod['Price']) * qty
                if nominal > 140000:
                    qty = max(1, 140000 // int(prod['Price']))
                    nominal = int(prod['Price']) * qty
            else:
                qty = random.randint(2, 4)
                nominal = int(prod['Price']) * qty
                while nominal < 60000:
                    qty += 1
                    nominal = int(prod['Price']) * qty
                if nominal > 95000:
                    qty = max(1, 95000 // int(prod['Price']))
                    nominal = int(prod['Price']) * qty
                
            komisi = komisi_map["E-Commerce"]
            keterangan = f"{qty}x {prod['Title']}"
            
            if status == "Gagal":
                komentar = random.choice(komentar_ecommerce_gagal)
                
            rec.update({
                "Layanan": "E-Commerce",
                "Keterangan": keterangan,
                "FarmerEmail": prod["FarmerEmail"],
                "FarmerName": prod["FarmerName"],
                "PenjualEmail": prod["FarmerEmail"],
                "PembeliEmail": buyer["Email"],
                "BuyerEmail": buyer["Email"],
                "BuyerName": buyer["Nama"],
                "BuyerCreatedAt": buyer["CreatedAt"],
            })
            
        elif layanan_type == "Chatbot Premium":
            valid_users = [u for u in all_non_admin if u['created_at_dt'] <= tgl_dt] or all_non_admin
            user_buyer = random.choice(valid_users)
            
            nominal = 30000
            komisi = 1.0
            keterangan = "Langganan AgroLink AI Premium"
            
            rec.update({
                "Layanan": "Chatbot Premium",
                "Keterangan": keterangan,
                "BuyerEmail": user_buyer["Email"],
                "BuyerName": user_buyer["Nama"],
                "BuyerCreatedAt": user_buyer["CreatedAt"],
                "UserEmail": user_buyer["Email"],
            })
            
        elif layanan_type == "Kemitraan":
            valid_farmers = [u for u in all_farmers if u['created_at_dt'] <= tgl_dt] or all_farmers
            valid_mitras = [u for u in mitras if u['created_at_dt'] <= tgl_dt] or mitras
            farmer = random.choice(valid_farmers)
            mitra = random.choice(valid_mitras)
            
            nominal = random.randint(350, 480) * 1000
            komisi = komisi_map["Kemitraan"]
            keterangan = random.choice(keterangan_kemitraan)
            
            if status == "Gagal":
                komentar = random.choice(komentar_kemitraan_gagal)
                
            rec.update({
                "Layanan": "Kemitraan",
                "Keterangan": keterangan,
                "FarmerEmail": farmer["Email"],
                "FarmerName": farmer["Nama"],
                "FarmerType": farmer.get("Type", ""),
                "FarmerCreatedAt": farmer["CreatedAt"],
                "MitraEmail": mitra["Email"],
                "MitraName": mitra["Nama"],
                "MitraCreatedAt": mitra["CreatedAt"],
                "PemberiKerjaEmail": farmer["Email"],
            })

        # Calculate financials
        if status == "Gagal":
            keuntungan_kotor = 0
            biaya_midtrans = 0
            keuntungan_bersih = 0
            total_mitra = 0
        else:
            keuntungan_kotor = nominal * komisi
            if metode in ["BCA", "BNI", "MANDIRI"]:
                biaya_midtrans = 4000 if nominal > 0 else 0
            else:
                biaya_midtrans = nominal * 0.007
            keuntungan_bersih = keuntungan_kotor - biaya_midtrans
            total_mitra = nominal - keuntungan_kotor if layanan_type != "Chatbot Premium" else 0

        rec.update({
            "IDTransaksi": "",
            "Tanggal": tgl_str,
            "Timestamp": ts.strftime("%Y-%m-%d %H:%M:%S"),
            "Bulan_Tahun": ts.strftime("%Y-%m"),
            "MetodePembayaran": metode,
            "NominalTransaksi": round(nominal),
            "PersentaseKomisi": komisi,
            "KeuntunganKotor": round(keuntungan_kotor),
            "BiayaMidtrans": round(biaya_midtrans),
            "KeuntunganBersih": round(keuntungan_bersih),
            "TotalDiterimaMitra": round(total_mitra),
            "StatusTransaksi": status,
            "KomentarUser": komentar
        })
        
        transactions.append(rec)

    # Urutkan berdasarkan timestamp
    transactions.sort(key=lambda x: x["Timestamp"])

    # Beri IDTransaksi berurutan
    for i, trx in enumerate(transactions, start=1):
        trx["IDTransaksi"] = f"INV/{trx['Tanggal'].replace('-', '')}/{i:03d}"

    return transactions

if __name__ == '__main__':
    print("Memulai proses pembuatan data transaksi AgroLink (Total 493)...")
    data_transaksi = generate_transactions()

    if data_transaksi:
        df = pd.DataFrame(data_transaksi)
        
        # Primary columns for Excel & JSON
        cols = [
            'IDTransaksi', 'Tanggal', 'Timestamp', 'Bulan_Tahun', 'Layanan', 'StatusTransaksi', 'Keterangan', 
            'KomentarUser', 'MetodePembayaran', 'NominalTransaksi', 'PersentaseKomisi', 'KeuntunganKotor', 
            'BiayaMidtrans', 'KeuntunganBersih', 'TotalDiterimaMitra', 'FarmerEmail', 'FarmerName', 'WorkerEmail', 
            'WorkerName', 'DriverEmail', 'DriverName', 'PenjualEmail', 'PembeliEmail', 'BuyerEmail', 
            'BuyerName', 'MitraEmail', 'MitraName', 'UserEmail', 'PemberiKerjaEmail', 'PekerjaEmail'
        ]
        # Keep only existing columns from the dict
        cols_present = [c for c in cols if c in df.columns]
        df = df[cols_present]
        
        # Export ke JSON
        df.to_json('seeders/new_transaction.json', orient='records', indent=2)
        df.to_json('AgroLink_493_Transaksi.json', orient='records', indent=2)
        df.to_json('AgroLink_484_Transaksi.json', orient='records', indent=2)
        df.to_json('AgroLink_593_Transaksi.json', orient='records', indent=2)
        print("Berhasil membuat file JSON: seeders/new_transaction.json dan AgroLink_493_Transaksi.json")

        # Export ke EXCEL (Terpisah per sheet bulan)
        excel_filename_493 = "AgroLink_493_Transaksi_PerBulan.xlsx"
        with pd.ExcelWriter(excel_filename_493, engine='openpyxl') as writer:
            grouped = df.groupby("Bulan_Tahun")
            for month, group in grouped:
                group_to_save = group.drop(columns=["Bulan_Tahun"])
                group_to_save.to_excel(writer, sheet_name=month, index=False)

        with pd.ExcelWriter("AgroLink_484_Transaksi_PerBulan.xlsx", engine='openpyxl') as writer:
            grouped = df.groupby("Bulan_Tahun")
            for month, group in grouped:
                group_to_save = group.drop(columns=["Bulan_Tahun"])
                group_to_save.to_excel(writer, sheet_name=month, index=False)

        with pd.ExcelWriter("AgroLink_593_Transaksi_PerBulan.xlsx", engine='openpyxl') as writer:
            grouped = df.groupby("Bulan_Tahun")
            for month, group in grouped:
                group_to_save = group.drop(columns=["Bulan_Tahun"])
                group_to_save.to_excel(writer, sheet_name=month, index=False)

        print(f"Berhasil membuat file Excel {excel_filename_493}")
        print(f"Selesai! Total {len(df)} transaksi berhasil digenerate dan diexport.")