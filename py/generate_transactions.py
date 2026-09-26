import os
import sys

# Add parent directory if needed
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from index import generate_transactions
import pandas as pd
import json

if __name__ == '__main__':
    print("Menjalankan generator transaksi dari generate_transactions.py (Total 493)...")
    data_transaksi = generate_transactions()

    if data_transaksi:
        df = pd.DataFrame(data_transaksi)
        
        cols = [
            'IDTransaksi', 'Tanggal', 'Timestamp', 'Bulan_Tahun', 'Layanan', 'StatusTransaksi', 'Keterangan', 
            'KomentarUser', 'MetodePembayaran', 'NominalTransaksi', 'PersentaseKomisi', 'KeuntunganKotor', 
            'BiayaMidtrans', 'KeuntunganBersih', 'TotalDiterimaMitra', 'FarmerEmail', 'FarmerName', 'WorkerEmail', 
            'WorkerName', 'DriverEmail', 'DriverName', 'PenjualEmail', 'PembeliEmail', 'BuyerEmail', 
            'BuyerName', 'MitraEmail', 'MitraName', 'UserEmail', 'PemberiKerjaEmail', 'PekerjaEmail'
        ]
        cols_present = [c for c in cols if c in df.columns]
        df = df[cols_present]
        
        # Export ke JSON
        df.to_json('seeders/new_transaction.json', orient='records', indent=2)
        df.to_json('AgroLink_493_Transaksi.json', orient='records', indent=2)
        df.to_json('AgroLink_484_Transaksi.json', orient='records', indent=2)
        df.to_json('AgroLink_593_Transaksi.json', orient='records', indent=2)
        print("Berhasil menyimpan file JSON (seeders/new_transaction.json, AgroLink_493_Transaksi.json)")

        # Export ke EXCEL
        excel_filename_493 = "AgroLink_493_Transaksi_PerBulan.xlsx"
        with pd.ExcelWriter(excel_filename_493, engine='openpyxl') as writer:
            grouped = df.groupby("Bulan_Tahun")
            for month, group in grouped:
                group_to_save = group.drop(columns=["Bulan_Tahun"])
                group_to_save.to_excel(writer, sheet_name=month, index=False)

        with pd.ExcelWriter('AgroLink_484_Transaksi_PerBulan.xlsx', engine='openpyxl') as writer:
            grouped = df.groupby("Bulan_Tahun")
            for month, group in grouped:
                group_to_save = group.drop(columns=["Bulan_Tahun"])
                group_to_save.to_excel(writer, sheet_name=month, index=False)

        with pd.ExcelWriter('AgroLink_593_Transaksi_PerBulan.xlsx', engine='openpyxl') as writer:
            grouped = df.groupby("Bulan_Tahun")
            for month, group in grouped:
                group_to_save = group.drop(columns=["Bulan_Tahun"])
                group_to_save.to_excel(writer, sheet_name=month, index=False)

        print(f"Berhasil membuat file Excel {excel_filename_493}")
        print(f"Selesai! Total {len(df)} transaksi berhasil digenerate.")
