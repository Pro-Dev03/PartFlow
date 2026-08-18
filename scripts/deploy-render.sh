#!/bin/bash

# PartFlow - Render Deployment Script
# هذا السكريبت يساعد في نشر المشروع على Render

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

# التحقق من المتطلبات
check_requirements() {
    print_info "التحقق من المتطلبات..."

    if ! command -v git &> /dev/null; then
        print_error "Git غير مثبت"
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        print_error "Docker غير مثبت"
        exit 1
    fi

    print_success "جميع المتطلبات متوفرة"
}

# إعداد المشروع للنشر
setup_project() {
    print_info "إعداد المشروع للنشر..."

    # التحقق من وجود render.yaml
    if [ ! -f "render.yaml" ]; then
        print_error "ملف render.yaml غير موجود"
        exit 1
    fi

    # التحقق من Dockerfiles
    if [ ! -f "backend/Dockerfile" ]; then
        print_error "backend/Dockerfile غير موجود"
        exit 1
    fi

    if [ ! -f "frontend/Dockerfile" ]; then
        print_error "frontend/Dockerfile غير موجود"
        exit 1
    fi

    if [ ! -f "worker/Dockerfile" ]; then
        print_error "worker/Dockerfile غير موجود"
        exit 1
    fi

    print_success "المشروع جاهز للنشر"
}

# اختبار البناء محلياً
test_build() {
    print_info "اختبار البناء محلياً..."

    # اختبار Backend
    print_info "بناء Backend..."
    docker build -t partflow-backend-test ./backend
    print_success "Backend تم بناؤه بنجاح"

    # اختبار Frontend
    print_info "بناء Frontend..."
    docker build -t partflow-frontend-test ./frontend
    print_success "Frontend تم بناؤه بنجاح"

    # اختبار Worker
    print_info "بناء Worker..."
    docker build -t partflow-worker-test ./worker
    print_success "Worker تم بناؤه بنجاح"

    # تنظيف الصور التجريبية
    docker rmi partflow-backend-test partflow-frontend-test partflow-worker-test

    print_success "جميع البناءات نجحت"
}

# إعداد Environment Variables
setup_env_vars() {
    print_info "إعداد Environment Variables..."

    if [ ! -f ".env" ]; then
        print_warning "ملف .env غير موجود، سيتم إنشاؤه من .env.example"
        cp .env.example .env
        print_warning "يرجى تحديث .env بالقيم الصحيحة"
    fi

    print_success "Environment Variables جاهزة"
}

# رفع الكود إلى GitHub
push_to_github() {
    print_info "رفع الكود إلى GitHub..."

    # التحقق من وجود تغييرات
    if git diff --quiet && git diff --cached --quiet; then
        print_info "لا توجد تغييرات للرفع"
    else
        git add .
        git commit -m "Prepare for Render deployment"
        git push origin main
        print_success "تم رفع الكود بنجاح"
    fi
}

# تعليمات النشر
deployment_instructions() {
    print_info "تعليمات النشر على Render:"
    echo ""
    echo "1. اذهب إلى https://render.com"
    echo "2. سجل الدخول أو أنشئ حساب جديد"
    echo "3. اربط حساب GitHub الخاص بك"
    echo "4. اختر 'New + > Blueprint'"
    echo "5. اختر مستودع PartFlow"
    echo "6. سيكتشف Render render.yaml تلقائياً"
    echo "7. راجع الإعدادات واضغط 'Apply'"
    echo ""
    print_info "أو يمكنك النشر يدوياً:"
    echo "1. أنشئ PostgreSQL Database"
    echo "2. أنشئ Redis"
    echo "3. أنشئ Backend Web Service"
    echo "4. أنشئ Frontend Web Service"
    echo "5. أنشئ Worker"
    echo ""
    print_info "راجع docs/DEPLOYMENT_RENDER.md للتفاصيل الكاملة"
}

# القائمة الرئيسية
main_menu() {
    echo ""
    echo "======================================"
    echo "  PartFlow - Render Deployment"
    echo "======================================"
    echo ""
    echo "1. التحقق من المتطلبات"
    echo "2. إعداد المشروع"
    echo "3. اختبار البناء محلياً"
    echo "4. إعداد Environment Variables"
    echo "5. رفع الكود إلى GitHub"
    echo "6. عرض تعليمات النشر"
    echo "7. تشغيل كل شيء"
    echo "8. خروج"
    echo ""
    read -p "اختر رقم (1-8): " choice

    case $choice in
        1)
            check_requirements
            ;;
        2)
            setup_project
            ;;
        3)
            test_build
            ;;
        4)
            setup_env_vars
            ;;
        5)
            push_to_github
            ;;
        6)
            deployment_instructions
            ;;
        7)
            check_requirements
            setup_project
            test_build
            setup_env_vars
            push_to_github
            deployment_instructions
            print_success "تم إكمال جميع الخطوات!"
            ;;
        8)
            print_info "وداعاً!"
            exit 0
            ;;
        *)
            print_error "اختيار غير صحيح"
            ;;
    esac
}

# التشغيل
if [ "$1" == "--all" ]; then
    check_requirements
    setup_project
    test_build
    setup_env_vars
    push_to_github
    deployment_instructions
    print_success "تم إكمال جميع الخطوات!"
else
    while true; do
        main_menu
        echo ""
        read -p "اضغط Enter للمتابعة أو q للخروج: " continue
        if [ "$continue" == "q" ]; then
            print_info "وداعاً!"
            exit 0
        fi
    done
fi
