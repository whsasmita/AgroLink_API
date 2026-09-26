import json
import pandas as pd
from datetime import datetime

with open('seeders/new_transaction.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

df = pd.DataFrame(data)
df['Tanggal_dt'] = pd.to_datetime(df['Tanggal'])

print(f"=== TOTAL TRANSAKSI ===")
print(f"Total Rows: {len(df)}")
print(f"Sukses: {len(df[df['StatusTransaksi'] == 'Sukses'])}")
print(f"Gagal:  {len(df[df['StatusTransaksi'] == 'Gagal'])}")

print("\n=== BREAKDOWN PER LAYANAN & STATUS ===")
print(pd.crosstab(df['Layanan'], df['StatusTransaksi'], margins=True))

print("\n=== BREAKDOWN SEKTOR PEKERJA ===")
# Peternak = Peternakan, Tukang = Tukang Bangunan, Pekerja = Pertanian
pekerja_df = df[df['Layanan'].isin(['Pekerja', 'Peternak', 'Tukang'])]
print(pd.crosstab(pekerja_df['Layanan'], pekerja_df['StatusTransaksi'], margins=True))

print("\n=== VALIDASI FASE 1 (1 Sept 2025 s/d 31 Mei 2026) ===")
fase1_df = df[df['Tanggal_dt'] <= '2026-05-31']
print(f"Total Trx Fase 1: {len(fase1_df)} (Sukses: {len(fase1_df[fase1_df['StatusTransaksi'] == 'Sukses'])}, Gagal: {len(fase1_df[fase1_df['StatusTransaksi'] == 'Gagal'])})")
print("Layanan yang muncul di Fase 1:")
print(fase1_df['Layanan'].value_counts())

fase1_forbidden = fase1_df[fase1_df['Layanan'].isin(['Chatbot Premium', 'Kemitraan', 'Peternak', 'Tukang'])]
print(f"Layanan DILARANG di Fase 1 (harus 0): {len(fase1_forbidden)}")

print("\n=== VALIDASI FASE 2 (1 Juni 2026 s/d 25 Sept 2026) ===")
fase2_df = df[df['Tanggal_dt'] >= '2026-06-01']
print(f"Total Trx Fase 2: {len(fase2_df)} (Sukses: {len(fase2_df[fase2_df['StatusTransaksi'] == 'Sukses'])}, Gagal: {len(fase2_df[fase2_df['StatusTransaksi'] == 'Gagal'])})")
print("Layanan yang muncul di Fase 2:")
print(fase2_df['Layanan'].value_counts())

print("\n=== VALIDASI AI PREMIUM ===")
ai_df = df[df['Layanan'] == 'Chatbot Premium']
print(f"Total AI Trx: {len(ai_df)}")
print(f"Unique Nominal AI: {ai_df['NominalTransaksi'].unique()}")
print(f"Unique KeuntunganKotor AI: {ai_df['KeuntunganKotor'].unique()}")
print(f"Min KeuntunganBersih AI: {ai_df['KeuntunganBersih'].min()}, Max: {ai_df['KeuntunganBersih'].max()}")
print(f"Total Diterima Mitra AI: {ai_df['TotalDiterimaMitra'].unique()}")

print("\n=== FINANSIAL PER BULAN ===")
monthly = df.groupby('Bulan_Tahun').agg(
    Total_Trx=('IDTransaksi', 'count'),
    Sukses_Trx=('StatusTransaksi', lambda s: (s == 'Sukses').sum()),
    Gagal_Trx=('StatusTransaksi', lambda s: (s == 'Gagal').sum()),
    Total_Gross=('NominalTransaksi', lambda x: x[df.loc[x.index, 'StatusTransaksi'] == 'Sukses'].sum()),
    Keuntungan_Kotor=('KeuntunganKotor', 'sum'),
    Biaya_Midtrans=('BiayaMidtrans', 'sum'),
    Keuntungan_Bersih=('KeuntunganBersih', 'sum'),
)
print(monthly.to_string())

print("\n=== CUMULATIVE NET PROFIT ===")
fase1_net = fase1_df['KeuntunganBersih'].sum()
total_net = df['KeuntunganBersih'].sum()
print(f"Cumulative Net Profit Fase 1 (s/d 31 Mei 2026): Rp {fase1_net:,.2f}")
print(f"Total Cumulative Net Profit (s/d 25 Sept 2026): Rp {total_net:,.2f}")
