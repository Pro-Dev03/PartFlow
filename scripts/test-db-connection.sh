#!/bin/bash

# PartFlow Database Connection Test
# هذا السكريبت يختبر الاتصال بقاعدة البيانات

set -e

# الألوان للإخراج
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# الدوال المساعدة
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# التحقق من وجود .env
if [ ! -f ".env" ]; then
    print_error "ملف .env غير موجود"
    print_info "قم بإنشاء ملف .env من .env.example"
    exit 1
fi

# تحميل متغيرات البيئة
export $(grep -v '^#' .env | xargs)

# التحقق من DATABASE_URL
if [ -z "$DATABASE_URL" ]; then
    print_error "DATABASE_URL غير موجود في .env"
    exit 1
fi

print_info "DATABASE_URL: ${DATABASE_URL:0:20}..."

# التحقق من وجود psql
if ! command -v psql &> /dev/null; then
    print_error "psql غير مثبت"
    print_info "قم بتثبيت PostgreSQL client"
    exit 1
fi

print_info "اختبار الاتصال بقاعدة البيانات..."

if psql "$DATABASE_URL" -c "SELECT version();" > /dev/null 2>&1; then
    print_success "الاتصال بقاعدة البيانات ناجح!"

    # عرض معلومات قاعدة البيانات
    print_info "معلومات قاعدة البيانات:"
    psql "$DATABASE_URL" -c "
        SELECT 
            current_database() as database,
            version() as version,
            current_user as user,
            inet_server_addr() as server_address;
    "

    # عرض الجداول
    print_info "الجداول الموجودة:"
    psql "$DATABASE_URL" -c "\dt"

    print_success "قاعدة البيانات جاهزة للاستخدام!"
else
    print_error "فشل الاتصال بقاعدة البيانات"
    print_info "تأكد من صحة DATABASE_URL في ملف .env"
    exit 1
fi
