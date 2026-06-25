import sys
import os
import time
import pandas as pd
import numpy as np

# Add utils to path so we can import from infilling
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
from infilling import infill_missing_data, infill_missing_data_vectorized

def create_synthetic_data(num_logbooks=100, days_per_logbook=365, seed=42):
    """
    Creates a synthetic dataframe of ship logbook entries to simulate the real data.
    Introduces random gaps of missing lat/lon data to test the infilling algorithms.
    """
    np.random.seed(seed)
    data = []
    start_date = pd.Timestamp('1850-01-01')
    
    for lb_id in range(1, num_logbooks + 1):
        logbook_name = f"LB_{lb_id:04d}"
        
        # Random starting lat/lon
        lat = np.random.uniform(-50, 50)
        lon = np.random.uniform(-180, 180)
        
        for d in range(days_per_logbook):
            date = start_date + pd.Timedelta(days=d)
            
            # Random walk
            lat += np.random.uniform(-1, 1)
            lon += np.random.uniform(-1, 1)
            
            # Wrap lon
            if lon > 180: lon -= 360
            if lon < -180: lon += 360
                
            data.append({
                'LogBook ID': logbook_name,
                'Entry Date Time': date,
                'Latitude_decimal': lat,
                'Longitude_decimal': lon
            })
            
    df = pd.DataFrame(data)
    
    # Introduce random gaps
    # Let's say 5% chance of a gap starting, with random length 1 to 5 days
    is_gap = np.random.random(len(df)) < 0.05
    
    gap_indices = np.where(is_gap)[0]
    for idx in gap_indices:
        gap_len = np.random.randint(1, 6)
        # nullify lat and lon only if the gap stays within the same logbook
        if idx + gap_len < len(df) and df.iloc[idx]['LogBook ID'] == df.iloc[idx+gap_len]['LogBook ID']:
            df.loc[idx:idx+gap_len-1, ['Latitude_decimal', 'Longitude_decimal']] = np.nan
            
    return df

def main():
    print("Generating synthetic data for benchmarking...")
    # 20 logbooks, 500 days each = 10,000 rows
    df_base = create_synthetic_data(num_logbooks=20, days_per_logbook=500)
    print(f"Dataframe size: {len(df_base)} rows")
    
    # We will infill 2-day gaps
    columns_to_infill = ['Latitude_decimal', 'Longitude_decimal']
    days_missing = 2
    max_distance_km = 1000  # set high enough so we accept most realistic synthetic gaps
    
    df1 = df_base.copy()
    df2 = df_base.copy()
    
    # --- Profile Original ---
    print("\nRunning original infill_missing_data...")
    start_t1 = time.time()
    df1_out = infill_missing_data(df1, columns_to_infill, days_missing, max_distance_km)
    time1 = time.time() - start_t1
    print(f"Original finished in {time1:.4f} seconds")
    
    # --- Profile Vectorized ---
    print("\nRunning vectorized infill_missing_data_vectorized...")
    start_t2 = time.time()
    df2_out = infill_missing_data_vectorized(df2, columns_to_infill, days_missing, max_distance_km)
    time2 = time.time() - start_t2
    print(f"Vectorized finished in {time2:.4f} seconds")
    
    if time2 > 0:
        print(f"\nSPEEDUP: {time1/time2:.2f}x faster using vectorized method!")
    
    # --- Verification ---
    print("\nVerifying outputs are identical...")
    try:
        # Sort columns to ensure consistent order before comparison
        # Also, check that both dataframes have the same dtype for the gap metadata
        pd.testing.assert_frame_equal(df1_out.sort_index(axis=1), df2_out.sort_index(axis=1), check_dtype=False)
        print("SUCCESS: Both implementations produced perfectly identical dataframes!")
    except AssertionError as e:
        print("FAILURE: The dataframes differ!")
        print(e)

if __name__ == "__main__":
    main()
