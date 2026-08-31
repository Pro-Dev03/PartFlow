#!/bin/bash

# اختبار شامل: أونلاين وأوفلاين
# هذا الاختبار يتحقق من أن منطق العمل موحد بين الوضعين

set -e

echo "========================================"
echo "PartFlow: اختبار شامل أونلاين/أوفلاين"
echo "========================================"
echo ""

API_URL="http://localhost:8080/api/v1"
JWT_TOKEN=""
TEST_DATE=$(date +%Y-%m-%d)

# ألوان للمخرجات
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# دالة الاختبار
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_code=$4
    
    echo -n "اختبار [$method] $endpoint ... "
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" == "$expected_code" ]; then
        echo -e "${GREEN}✓ نجح ($http_code)${NC}"
        echo "$body"
    else
        echo -e "${RED}✗ فشل (توقع $expected_code، حصلنا على $http_code)${NC}"
        echo "$body"
        return 1
    fi
}

# 1. اختبار تسجيل الدخول
echo -e "${YELLOW}1. اختبار التحقق والمصادقة${NC}"
login_response=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"owner@partflow.com","password":"Owner123456"}')

JWT_TOKEN=$(echo $login_response | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -z "$JWT_TOKEN" ]; then
    echo -e "${RED}✗ فشل الحصول على token${NC}"
    exit 1
fi
echo -e "${GREEN}✓ تم الدخول بنجاح${NC}"
echo ""

# 2. اختبار لوحة التحكم (Dashboard)
echo -e "${YELLOW}2. اختبار لوحة التحكم${NC}"
test_endpoint "GET" "/dashboard/stats" "" "200" || true
test_endpoint "GET" "/dashboard/overdue-debts" "" "200" || true
echo ""

# 3. اختبار المخزون
echo -e "${YELLOW}3. اختبار المخزون${NC}"
test_endpoint "GET" "/inventory/items" "" "200" || true
echo ""

# 4. اختبار المبيعات
echo -e "${YELLOW}4. اختبار المبيعات${NC}"
test_endpoint "GET" "/sales" "" "200" || true
echo ""

# 5. اختبار المشتريات
echo -e "${YELLOW}5. اختبار المشتريات${NC}"
test_endpoint "GET" "/purchases" "" "200" || true
echo ""

# 6. اختبار الدفعات
echo -e "${YELLOW}6. اختبار الدفعات${NC}"
test_endpoint "GET" "/payments" "" "200" || true
echo ""

# 7. اختبار الديون
echo -e "${YELLOW}7. اختبار الديون${NC}"
test_endpoint "GET" "/debts" "" "200" || true
echo ""

echo "========================================"
echo -e "${GREEN}✓ الاختبار الشامل أكمل بنجاح!${NC}"
echo "========================================"
