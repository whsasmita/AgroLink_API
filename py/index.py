import pandas as pd
import random
import json
from datetime import datetime, timedelta

def generate_transactions():
    transactions = []

    # 1. Mengambil data json produk yang baru saja dibuat
    try:
        with open('./data_produk.json', 'r', encoding='utf-8') as f:
            produk_data = json.load(f)
    except FileNotFoundError:
        print("Error: File './data_produk.json' tidak ditemukan. Pastikan file ada di direktori yang sama.")
        return []

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
    
    metode_bayar = ["BCA", "BNI", "MANDIRI", "QRIS", "DANA", "GoPay", "ShopeePay"]

    # Fungsi helper untuk membuat baris record
    def create_record(tgl, layanan, komisi, nominal, metode, keterangan="-"):
        keuntungan_kotor = nominal * komisi

        # Biaya potong pihak ketiga
        if metode in ["BCA", "BNI", "MANDIRI"]:
            biaya_midtrans = 4000 if nominal > 0 else 0
        else:
            biaya_midtrans = nominal * 0.007

        keuntungan_bersih = keuntungan_kotor - biaya_midtrans
        # Kemitraan dan layanan biasa dibagi hasil ke Mitra, Chatbot Premium full ke platform
        total_mitra = nominal - keuntungan_kotor if layanan != "Chatbot Premium" else 0

        return {
            "IDTransaksi": "", 
            "Tanggal": tgl.strftime("%Y-%m-%d"),
            "Bulan_Tahun": tgl.strftime("%Y-%m"), # Kolom bantuan pemisah sheet excel
            "Layanan": layanan,
            "Keterangan": keterangan, # Tambahan kolom untuk menunjukkan produk JSON
            "MetodePembayaran": metode,
            "NominalTransaksi": round(nominal),
            "PersentaseKomisi": komisi,
            "KeuntunganKotor": round(keuntungan_kotor),
            "BiayaMidtrans": round(biaya_midtrans),
            "KeuntunganBersih": round(keuntungan_bersih),
            "TotalDiterimaMitra": round(total_mitra)
        }

    # --- FASE 1: Sept 2025 - Mei 2026 (236 Transaksi) ---
    start_date_1 = datetime(2025, 9, 1)
    end_date_1 = datetime(2026, 5, 31)
    days_1 = (end_date_1 - start_date_1).days

    layanan_fase_1 = ["Pekerja", "Ekspedisi", "E-Commerce"]
    # Bobot probabilitas: Pekerja (50%), Ekspedisi (35%), E-Commerce (15%)
    bobot_fase_1 = [50, 35, 15] 

    for _ in range(236):
        tgl = start_date_1 + timedelta(days=random.randint(0, days_1))
        # Pilih layanan berdasarkan bobot dominasi
        layanan = random.choices(layanan_fase_1, weights=bobot_fase_1, k=1)[0]
        komisi = komisi_map[layanan]
        metode = random.choice(metode_bayar)
        keterangan = "-"

        if layanan == "E-Commerce":
            # Ambil produk acak dari data JSON
            produk = random.choice(produk_data)
            qty = random.randint(1, 15) # Beli 1 sampai 15 quantity
            nominal = produk['harga'] * qty
            
            # Agar transaksi masuk akal secara nominal, kita pastikan minimal belanja 75k
            while nominal < 75000:
                qty += 1
                nominal = produk['harga'] * qty
                
            keterangan = f"{qty}x {produk['item']} ({produk['satuan']})"
        else:
            # Nominal acak minimal 75.000 (kelipatan ribuan)
            nominal = random.randint(75, 500) * 1000 

        transactions.append(create_record(tgl, layanan, komisi, nominal, metode, keterangan))


    # --- FASE 2: Juni 2026 - 20 Ags 2026 (354 Transaksi Biasa + 3 Kemitraan = 357) ---
    start_date_2 = datetime(2026, 6, 1)
    end_date_2 = datetime(2026, 8, 20)
    days_2 = (end_date_2 - start_date_2).days

    layanan_fase_2 = ["Pekerja", "Ekspedisi", "Chatbot Premium", "E-Commerce", "Tukang", "Peternak"]
    # Dominasi berurutan: Pekerja (35) > Ekspedisi (25) > Premium (20) > lainnya
    bobot_fase_2 = [35, 25, 20, 10, 5, 5]

    for _ in range(354):
        tgl = start_date_2 + timedelta(days=random.randint(0, days_2))
        layanan = random.choices(layanan_fase_2, weights=bobot_fase_2, k=1)[0]
        komisi = komisi_map[layanan]
        metode = random.choice(metode_bayar)
        keterangan = "-"

        if layanan == "Chatbot Premium":
            # Harga Rp 150.000 + Pajak 11% (Rp 16.500)
            nominal = 166500 
            keterangan = "Langganan AgroLink AI Premium"
        elif layanan == "E-Commerce":
            produk = random.choice(produk_data)
            qty = random.randint(1, 15)
            nominal = produk['harga'] * qty
            while nominal < 75000:
                qty += 1
                nominal = produk['harga'] * qty
            keterangan = f"{qty}x {produk['item']} ({produk['satuan']})"
        else:
            # Nominal acak minimal 75.000 (kelipatan ribuan)
            nominal = random.randint(75, 500) * 1000

        transactions.append(create_record(tgl, layanan, komisi, nominal, metode, keterangan))


    # --- SISIPAN KHUSUS: 3 Transaksi Kemitraan (2 Juta s/d 5 Juta) khusus Juli - Ags 2026 ---
    start_kemitraan = datetime(2026, 7, 1)
    end_kemitraan = datetime(2026, 8, 20)
    days_kemitraan = (end_kemitraan - start_kemitraan).days

    for _ in range(3):
        tgl = start_kemitraan + timedelta(days=random.randint(0, days_kemitraan))
        metode = random.choice(["BCA", "BNI", "MANDIRI"])
        # Nominal acak 2 jt - 5 jt kelipatan ribuan
        nominal = random.randint(2000, 5000) * 1000
        keterangan = "Pembayaran Kemitraan B2B"
        transactions.append(create_record(tgl, "Kemitraan", komisi_map["Kemitraan"], nominal, metode, keterangan))


    # Urutkan berdasarkan tanggal
    transactions.sort(key=lambda x: datetime.strptime(x["Tanggal"], "%Y-%m-%d"))

    # Tambahkan ID Transaksi yang berurutan
    for i, trx in enumerate(transactions, start=1):
        trx["IDTransaksi"] = f"INV/{trx['Tanggal'].replace('-', '')}/{i:03d}"

    return transactions

print("Memulai proses pembuatan data...")
data_transaksi = generate_transactions()

if data_transaksi:
    # 1. Bikin format DataFrame dan atur urutan kolom
    df = pd.DataFrame(data_transaksi)
    cols = ['IDTransaksi', 'Tanggal', 'Bulan_Tahun', 'Layanan', 'Keterangan', 'MetodePembayaran', 
            'NominalTransaksi', 'PersentaseKomisi', 'KeuntunganKotor', 'BiayaMidtrans', 
            'KeuntunganBersih', 'TotalDiterimaMitra']
    df = df[cols]
    
    # Export ke JSON (Hapus kolom Bulan_Tahun karena itu hanya bantuan untuk Excel)
    df_export_json = df.drop(columns=['Bulan_Tahun'])
    json_filename = 'AgroLink_593_Transaksi.json'
    df_export_json.to_json(json_filename, orient='records', indent=2)
    print(f"Berhasil membuat file {json_filename}")

    # 3. Export ke EXCEL (Terpisah per sheet bulan)
    excel_filename = "AgroLink_593_Transaksi_PerBulan.xlsx"
    with pd.ExcelWriter(excel_filename, engine='openpyxl') as writer:
        grouped = df.groupby("Bulan_Tahun")
        for month, group in grouped:
            group_to_save = group.drop(columns=["Bulan_Tahun"])
            group_to_save.to_excel(writer, sheet_name=month, index=False)

    print(f"Berhasil membuat file {excel_filename}")
    print(f"Selesai! Dua format file (JSON dan Excel) dengan total {len(df)} transaksi sudah berhasil dicetak.")