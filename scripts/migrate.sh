#!/bin/bash

# PartFlow Database Migration Script
# هذا السكريبت يشغل migrations على قاعدة بيانات Supabase

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

# تشغيل migrations
print_info "بدء تشغيل migrations..."

MIGRATIONS_DIR="backend/migrations"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    print_error "مجلد migrations غير موجود"
    exit 1
fi

# تشغيل ملفات migrations بالترتيب
for migration_file in "$MIGRATIONS_DIR"/*.sql; do
    if [ -f "$migration_file" ]; then
        filename=$(basename "$migration_file")
        print_info "تشغيل: $filename"

        if psql "$DATABASE_URL" -f "$migration_file"; then
            print_success "تم تشغيل $filename بنجاح"
        else
            print_error "فشل تشغيل $filename"
            exit 1
        fi
    fi
done

print_success "تم تشغيل جميع migrations بنجاح!"
print_info "قاعدة البيانات جاهزة للاستخدام"
