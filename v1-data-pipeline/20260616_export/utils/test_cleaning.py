import pytest
import pandas as pd
import numpy as np
from cleaning import (
    clean_page_column,
    clean_depth_column,
    clean_wind_dirs,
    normalize_coords,
    clean_remaining_strings,
    PAGE_MAP,
    PAGE_NAN,
    DEPTH_MAP,
    DEPTH_NAN
)

def test_clean_page_column():
    test_values = list(PAGE_MAP.keys()) + list(PAGE_NAN) + ['42']
    df_test = pd.DataFrame({'Page': test_values})
    
    expected_page = [float(v) for v in PAGE_MAP.values()] + [np.nan] * len(PAGE_NAN) + [42.0]
    expected_df = pd.DataFrame({
        'Page': expected_page,
        'Page_og': test_values
    })
    
    result_df = clean_page_column(df_test.copy())
    pd.testing.assert_frame_equal(result_df, expected_df)

def test_clean_page_column_already_clean():
    test_values = [1.0, 42.0, 100.0, np.nan]
    df_test = pd.DataFrame({'Page': test_values})
    
    expected_df = pd.DataFrame({
        'Page': test_values,
        'Page_og': test_values
    })
    
    result_df = clean_page_column(df_test.copy())
    pd.testing.assert_frame_equal(result_df, expected_df)

def test_clean_page_column_string_numbers():
    test_values = ['1', '42', '100', ' 50 ', '"10"']
    df_test = pd.DataFrame({'Page': test_values})
    
    expected_page = [1.0, 42.0, 100.0, 50.0, 10.0]
    expected_df = pd.DataFrame({
        'Page': expected_page,
        'Page_og': test_values
    })
    
    result_df = clean_page_column(df_test.copy())
    pd.testing.assert_frame_equal(result_df, expected_df)

def test_clean_depth_column():
    test_values = list(DEPTH_MAP.keys()) + list(DEPTH_NAN) + ['42']
    df_test = pd.DataFrame({'Depth': test_values})
    
    expected_depth = [float(v) for v in DEPTH_MAP.values()] + [np.nan] * len(DEPTH_NAN) + [42.0]
    expected_df = pd.DataFrame({
        'Depth': expected_depth,
        'Depth_og': test_values
    })
    
    result_df = clean_depth_column(df_test.copy())
    pd.testing.assert_frame_equal(result_df, expected_df)

def test_clean_depth_column_already_clean():
    test_values = [1.0, 42.0, 100.5, np.nan]
    df_test = pd.DataFrame({'Depth': test_values})
    
    expected_df = pd.DataFrame({
        'Depth': test_values,
        'Depth_og': test_values
    })
    
    result_df = clean_depth_column(df_test.copy())
    pd.testing.assert_frame_equal(result_df, expected_df)

def test_clean_depth_column_string_numbers():
    test_values = ['1', '42', '100', ' 50 ', '10.5']
    df_test = pd.DataFrame({'Depth': test_values})
    
    expected_depth = [1.0, 42.0, 100.0, 50.0, 10.5]
    expected_df = pd.DataFrame({
        'Depth': expected_depth,
        'Depth_og': test_values
    })
    
    result_df = clean_depth_column(df_test.copy())
    pd.testing.assert_frame_equal(result_df, expected_df)

def test_clean_wind_dirs():
    # TODO: Add test implementation
    pass

def test_normalize_coords():
    # TODO: Add test implementation
    pass

def test_clean_remaining_strings():
    # TODO: Add test implementation
    pass
