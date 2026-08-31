#!/bin/bash

echo "🔧 تثبيت المتطلبات لسكربت إصلاح قاعدة البيانات"

# تثبيت Python و psycopg2
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 غير مثبت"
    echo "📝 تثبيت Python3..."
    sudo apt-get update
    sudo apt-get install -y python3 python3-pip
fi

# تثبيت المكتبات المطلوبة
echo "📦 تثبيت المكتبات المطلوبة..."
pip3 install psycopg2-binary

# جعل السكربت قابلاً للتنفيذ
chmod +x backend/scripts/fix_db_connection.py

echo "✅ تم تثبيت جميع المتطلبات"
echo "🚀 يمكنك الآن تشغيل السكربت:"
echo "   python3 backend/scripts/fix_db_connection.py"
