#!/bin/bash

# PartFlow Performance Test Script
# اختبار الأداء الشامل مع البيانات الحقيقية

BASE_URL="http://localhost:8080/api/v1"
OUTPUT_FILE="performance-results.txt"

echo "PartFlow Performance Test Results" > $OUTPUT_FILE
echo "====================================" >> $OUTPUT_FILE
echo "Date: $(date)" >> $OUTPUT_FILE
echo "" >> $OUTPUT_FILE

# Test 1: Dashboard Stats
echo "Test 1: Dashboard Stats" >> $OUTPUT_FILE
curl -s -w "Time: %{time_total}s\n" "$BASE_URL/dashboard/stats" >> $OUTPUT_FILE
echo "" >> $OUTPUT_FILE

# Test 2: Products List
echo "Test 2: Products List" >> $OUTPUT_FILE
curl -s -w "Time: %{time_total}s\n" "$BASE_URL/products?page=1&per_page=20" >> $OUTPUT_FILE
echo "" >> $OUTPUT_FILE

# Test 3: Products Search
echo "Test 3: Products Search (SSD)" >> $OUTPUT_FILE
curl -s -w "Time: %{time_total}s\n" "$BASE_URL/products?search=SSD" >> $OUTPUT_FILE
echo "" >> $OUTPUT_FILE

# Test 4: Customers List
echo "Test 4: Customers List" >> $OUTPUT_FILE
curl -s -w "Time: %{time_total}s\n" "$BASE_URL/customers?page=1&per_page=20" >> $OUTPUT_FILE
echo "" >> $OUTPUT_FILE

# Test 5: Debts List
echo "Test 5: Debts List" >> $OUTPUT_FILE
curl -s -w "Time: %{time_total}s\n" "$BASE_URL/debts?page=1&per_page=20" >> $OUTPUT_FILE
echo "" >> $OUTPUT_FILE

# Test 6: Sales List
echo "Test 6: Sales List" >> $OUTPUT_FILE
curl -s -w "Time: %{time_total}s\n" "$BASE_URL/sales?page=1&per_page=20" >> $OUTPUT_FILE
echo "" >> $OUTPUT_FILE

# Test 7: Barcode Lookup (existing)
echo "Test 7: Barcode Lookup (Product not found)" >> $OUTPUT_FILE
curl -s -w "Time: %{time_total}s\n" "$BASE_URL/barcodes/product/1234567890129" >> $OUTPUT_FILE
echo "" >> $OUTPUT_FILE

echo "Performance test completed. Results saved to $OUTPUT_FILE"
cat $OUTPUT_FILE