import pandas as pd
import openpyxl
import json
import random
from datetime import datetime

# 1. Load exact Excel transactions
excel_path = "AgroLink_593_Transaksi_PerBulan.xlsx"
wb = openpyxl.load_workbook(excel_path)
all_dfs = []
for sheet in wb.sheetnames:
    df = pd.read_excel(excel_path, sheet_name=sheet)
    all_dfs.append(df)
df_excel = pd.concat(all_dfs, ignore_index=True)

print(f"Loaded {len(df_excel)} transactions from Excel.")
print(f"Total NominalTransaksi: Rp {df_excel['NominalTransaksi'].sum():,.2f}")
print(f"Total KeuntunganBersih: Rp {df_excel['KeuntunganBersih'].sum():,.2f}")

# 2. Load users
with open("seeders/users_seed.json", "r", encoding="utf-8") as f:
    users = json.load(f)

for u in users:
    u["created_at_dt"] = datetime.strptime(u["CreatedAt"], "%Y-%m-%d %H:%M:%S")

# Categorize users
all_farmers_agri = [u for u in users if u["Role"] == "farmer" and u.get("Type") == "agriculture"]
all_farmers_live = [u for u in users if u["Role"] == "farmer" and u.get("Type") == "livestock"]
all_farmers_const = [u for u in users if u["Role"] == "farmer" and u.get("Type") == "construction"]
all_farmers = [u for u in users if u["Role"] == "farmer"]

all_workers_agri = [u for u in users if u["Role"] == "worker" and "Pertanian" in u.get("Skills", [])]
all_workers_live = [u for u in users if u["Role"] == "worker" and "Peternakan" in u.get("Skills", [])]
all_workers_const = [u for u in users if u["Role"] == "worker" and "Tukang Bangunan" in u.get("Skills", [])]
all_workers = [u for u in users if u["Role"] == "worker"]

all_drivers = [u for u in users if u["Role"] == "driver"]
all_mitras = [u for u in users if u["Role"] == "mitra"]
all_general = [u for u in users if u["Role"] == "general"]

random.seed(42)

processed_transactions = []
time_travel_violations = 0
role_mismatches = 0

for idx, row in df_excel.iterrows():
    tgl_str = str(row["Tanggal"]).split()[0]
    # create timestamp (e.g. adding time of day)
    hour = random.randint(7, 21)
    minute = random.randint(0, 59)
    second = random.randint(0, 59)
    tgl_dt = datetime.strptime(f"{tgl_str} {hour:02d}:{minute:02d}:{second:02d}", "%Y-%m-%d %H:%M:%S")
    
    layanan = str(row["Layanan"]).strip()
    keterangan = str(row.get("Keterangan", "-")).strip()
    
    rec = {
        "IDTransaksi": str(row["IDTransaksi"]).strip(),
        "Tanggal": tgl_str,
        "Timestamp": tgl_dt.strftime("%Y-%m-%d %H:%M:%S"),
        "Layanan": layanan,
        "Keterangan": keterangan,
        "MetodePembayaran": str(row["MetodePembayaran"]).strip(),
        "NominalTransaksi": int(row["NominalTransaksi"]),
        "PersentaseKomisi": float(row["PersentaseKomisi"]),
        "KeuntunganKotor": int(row["KeuntunganKotor"]),
        "BiayaMidtrans": int(row["BiayaMidtrans"]),
        "KeuntunganBersih": int(row["KeuntunganBersih"]),
        "TotalDiterimaMitra": int(row["TotalDiterimaMitra"]),
    }
    
    # User linking
    if layanan in ["Pekerja", "Tukang", "Peternak"]:
        if layanan == "Tukang" or "Tukang" in keterangan:
            valid_farmers = [u for u in all_farmers_const if u["created_at_dt"] <= tgl_dt]
            valid_workers = [u for u in all_workers_const if u["created_at_dt"] <= tgl_dt]
            if not valid_farmers:
                valid_farmers = all_farmers_const
            if not valid_workers:
                valid_workers = all_workers_const
        elif layanan == "Peternak" or "Ternak" in keterangan:
            valid_farmers = [u for u in all_farmers_live if u["created_at_dt"] <= tgl_dt]
            valid_workers = [u for u in all_workers_live if u["created_at_dt"] <= tgl_dt]
            if not valid_farmers:
                valid_farmers = all_farmers_live
            if not valid_workers:
                valid_workers = all_workers_live
        else: # Tani Link / Pekerja Pertanian
            valid_farmers = [u for u in all_farmers_agri if u["created_at_dt"] <= tgl_dt]
            valid_workers = [u for u in all_workers_agri if u["created_at_dt"] <= tgl_dt]
            if not valid_farmers:
                valid_farmers = all_farmers_agri
            if not valid_workers:
                valid_workers = all_workers_agri
                
        farmer = random.choice(valid_farmers)
        worker = random.choice(valid_workers)
        
        rec["FarmerEmail"] = farmer["Email"]
        rec["FarmerName"] = farmer["Nama"]
        rec["FarmerType"] = farmer.get("Type", "")
        rec["FarmerCreatedAt"] = farmer["CreatedAt"]
        
        rec["WorkerEmail"] = worker["Email"]
        rec["WorkerName"] = worker["Nama"]
        rec["WorkerSkills"] = worker.get("Skills", [])
        rec["WorkerCreatedAt"] = worker["CreatedAt"]
        
    elif layanan == "Ekspedisi":
        valid_farmers = [u for u in all_farmers if u["created_at_dt"] <= tgl_dt]
        valid_drivers = [u for u in all_drivers if u["created_at_dt"] <= tgl_dt]
        if not valid_farmers:
            valid_farmers = all_farmers
        if not valid_drivers:
            valid_drivers = all_drivers
            
        farmer = random.choice(valid_farmers)
        driver = random.choice(valid_drivers)
        
        rec["FarmerEmail"] = farmer["Email"]
        rec["FarmerName"] = farmer["Nama"]
        rec["FarmerType"] = farmer.get("Type", "")
        rec["FarmerCreatedAt"] = farmer["CreatedAt"]
        
        rec["DriverEmail"] = driver["Email"]
        rec["DriverName"] = driver["Nama"]
        rec["DriverCreatedAt"] = driver["CreatedAt"]
        
    elif layanan == "E-Commerce":
        valid_sellers = [u for u in (all_farmers_agri + all_farmers_live) if u["created_at_dt"] <= tgl_dt]
        valid_buyers = [u for u in (all_general + all_farmers + all_workers) if u["created_at_dt"] <= tgl_dt]
        if not valid_sellers:
            valid_sellers = all_farmers_agri + all_farmers_live
        if not valid_buyers:
            valid_buyers = all_general + all_farmers + all_workers
            
        seller = random.choice(valid_sellers)
        buyer = random.choice(valid_buyers)
        
        rec["FarmerEmail"] = seller["Email"]
        rec["FarmerName"] = seller["Nama"]
        rec["FarmerType"] = seller.get("Type", "")
        rec["FarmerCreatedAt"] = seller["CreatedAt"]
        
        rec["BuyerEmail"] = buyer["Email"]
        rec["BuyerName"] = buyer["Nama"]
        rec["BuyerCreatedAt"] = buyer["CreatedAt"]
        
    elif layanan == "Chatbot Premium":
        valid_users = [u for u in users if u["Role"] != "admin" and u["created_at_dt"] <= tgl_dt]
        if not valid_users:
            valid_users = [u for u in users if u["Role"] != "admin"]
        user = random.choice(valid_users)
        
        rec["BuyerEmail"] = user["Email"]
        rec["BuyerName"] = user["Nama"]
        rec["BuyerCreatedAt"] = user["CreatedAt"]
        
    elif layanan == "Kemitraan":
        valid_farmers = [u for u in all_farmers if u["created_at_dt"] <= tgl_dt]
        valid_mitras = [u for u in all_mitras if u["created_at_dt"] <= tgl_dt]
        if not valid_farmers:
            valid_farmers = all_farmers
        if not valid_mitras:
            valid_mitras = all_mitras
            
        farmer = random.choice(valid_farmers)
        mitra = random.choice(valid_mitras)
        
        rec["FarmerEmail"] = farmer["Email"]
        rec["FarmerName"] = farmer["Nama"]
        rec["FarmerType"] = farmer.get("Type", "")
        rec["FarmerCreatedAt"] = farmer["CreatedAt"]
        
        rec["MitraEmail"] = mitra["Email"]
        rec["MitraName"] = mitra["Nama"]
        rec["MitraCreatedAt"] = mitra["CreatedAt"]
        
    processed_transactions.append(rec)

# Validation of results
print(f"\nProcessed {len(processed_transactions)} transactions.")

# Save to agrolink_593_transaksi.json
with open("agrolink_593_transaksi.json", "w", encoding="utf-8") as f:
    json.dump(processed_transactions, f, indent=2, ensure_ascii=False)

print("Saved to agrolink_593_transaksi.json successfully.")
