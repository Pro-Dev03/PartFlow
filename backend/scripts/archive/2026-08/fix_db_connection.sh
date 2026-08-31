#!/bin/bash

echo "🚀 سكربت حل مشاكل الاتصال بقاعدة بيانات Supabase"
echo "=================================================="

# التحقق من وجود ملف .env
ENV_FILE="/home/dev-bit/project/PartFlow/backend/.env"

if [ ! -f "$ENV_FILE" ]; then
    echo "❌ ملف .env غير موجود"
    exit 1
fi

# استخراج DATABASE_URL الحالي
CURRENT_DB_URL=$(grep "^DATABASE_URL=" "$ENV_FILE" | cut -d'=' -f2)

if [ -z "$CURRENT_DB_URL" ]; then
    echo "❌ لم يتم العثور على DATABASE_URL في ملف .env"
    exit 1
fi

echo "📌 سلسلة الاتصال الحالية: $CURRENT_DB_URL"

# تثبيت postgresql-client إذا لم يكن موجوداً
if ! command -v psql &> /dev/null; then
    echo "📦 تثبيت postgresql-client..."
    sudo apt-get update
    sudo apt-get install -y postgresql-client
fi

# اختبار اتصالات مختلفة
test_connection() {
    local conn_string="$1"
    local conn_name="$2"
    
    echo "🔍 جاري اختبار: $conn_name"
    
    if PGPASSWORD=$(echo "$conn_string" | sed 's/.*://\([^@]*\)@.*/\1/') psql "$conn_string" -c "SELECT 1" > /dev/null 2>&1; then
        echo "✅ نجح الاتصال: $conn_name"
        return 0
    else
        echo "❌ فشل الاتصال: $conn_name"
        return 1
    fi
}

# استخراج معلومات الاتصال
USER=$(echo "$CURRENT_DB_URL" | sed 's/.*:\/\/\([^:]*\):.*/\1/')
PASSWORD=$(echo "$CURRENT_DB_URL" | sed 's/.*:\/\/[^:]*:\([^@]*\)@.*/\1/')
HOST=$(echo "$CURRENT_DB_URL" | sed 's/.*@\([^:]*\):.*/\1/')
PORT=$(echo "$CURRENT_DB_URL" | sed 's/.*:\([0-9]*\)\/.*/\1/')
DB=$(echo "$CURRENT_DB_URL" | sed 's/.*\/\([^?]*\).*/\1/')

echo "📊 معلومات الاتصال:"
echo "   المستخدم: $USER"
echo "   المضيف: $HOST"
echo "   المنفذ: $PORT"
echo "   قاعدة البيانات: $DB"

# قائمة الاتصالات المختلفة
CONNECTIONS=(
    "$CURRENT_DB_URL:الأصلي"
    "postgresql://$USER:$PASSWORD@$HOST:$PORT/$DB?sslmode=require:مع SSL"
    "postgresql://$USER:$PASSWORD@$HOST:$PORT/$DB?sslmode=disable:بدون SSL"
)

# إذا كان المضيف يحتوي على pooler، جرب db مباشرة
if [[ "$HOST" == *"pooler"* ]]; then
    DB_HOST="${HOST/pooler/db}"
    CONNECTIONS+=("postgresql://$USER:$PASSWORD@$DB_HOST:$PORT/$DB?sslmode=require:Direct DB")
    CONNECTIONS+=("postgresql://$USER:$PASSWORD@$DB_HOST:$PORT/$DB?sslmode=disable:Direct DB بدون SSL")
fi

# جرب المنفذ 6543 إذا كان pooler
if [[ "$HOST" == *"pooler"* ]]; then
    CONNECTIONS+=("postgresql://$USER:$PASSWORD@$HOST:6543/$DB?sslmode=require:Pooler 6543")
fi

echo ""
echo "🔍 جاري اختبار الاتصالات المختلفة..."
echo "=================================================="

SUCCESSFUL_CONN=""
SUCCESSFUL_NAME=""

for conn in "${CONNECTIONS[@]}"; do
    IFS=':' read -r conn_string conn_name <<< "$conn"
    
    if test_connection "$conn_string" "$conn_name"; then
        SUCCESSFUL_CONN="$conn_string"
        SUCCESSFUL_NAME="$conn_name"
        break
    fi
done

if [ -n "$SUCCESSFUL_CONN" ]; then
    echo ""
    echo "🎉 تم العثور على اتصال ناجح!"
    echo "✅ الاتصال الناجح: $SUCCESSFUL_NAME"
    echo "📌 سلسلة الاتصال: $SUCCESSFUL_CONN"
    
    # تحديث ملف .env
    echo "📝 تحديث ملف .env..."
    sed -i "s|^DATABASE_URL=.*|DATABASE_URL=$SUCCESSFUL_CONN|" "$ENV_FILE"
    echo "✅ تم تحديث ملف .env"
    
    echo ""
    echo "🚀 يمكنك الآن تشغيل التطبيق:"
    echo "   cd /home/dev-bit/project/PartFlow/backend && ./api"
else
    echo ""
    echo "❌ لم يتمكن من الاتصال بأي من السلاسل"
    echo ""
    echo "🔧 الحلول المقترحة:"
    echo "1. تحقق من إعدادات Supabase:"
    echo "   - تأكد من تفعيل SSL"
    echo "   - تحقق من قواعد الوصول"
    echo "   - تأكد من أن IP الخاص بك غير محظور"
    echo ""
    echo "2. جرب قاعدة بيانات محلية:"
    echo "   - استخدم Docker لإنشاء PostgreSQL محلي"
    echo "   - قم بتحديث DATABASE_URL للاتصال بالقاعدة المحلية"
    echo ""
    echo "3. تواصل مع دعم Supabase"
    echo ""
    echo "💡 هل تريد إنشاء قاعدة بيانات محلية؟ (y/n)"
    read -r response
    
    if [[ "$response" == "y" || "$response" == "Y" ]]; then
        echo "🔄 إنشاء قاعدة بيانات محلية..."
        
        # إنشاء docker-compose.yml
        DOCKER_COMPOSE="/home/dev-bit/project/PartFlow/backend/docker-compose.yml"
        cat > "$DOCKER_COMPOSE" <<EOF
version: '3.8'
services:
  postgres:
    image: postgres:15
    container_name: partflow-db
    environment:
      POSTGRES_DB: partflow
      POSTGRES_USER: partflow
      POSTGRES_PASSWORD: partflow123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U partflow"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
EOF
        
        echo "✅ تم إنشاء ملف docker-compose.yml"
        echo "📝 لتشغيل قاعدة البيانات المحلية:"
        echo "   cd /home/dev-bit/project/PartFlow/backend"
        echo "   docker-compose up -d"
        echo ""
        echo "📝 ثم قم بتحديث DATABASE_URL في .env إلى:"
        echo "   DATABASE_URL=postgresql://partflow:partflow123@localhost:5432/partflow"
    fi
fi

echo ""
echo "🏁 انتهى السكربت"
