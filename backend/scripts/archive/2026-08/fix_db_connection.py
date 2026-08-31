#!/usr/bin/env python3
"""
سكربت شامل لحل مشاكل الاتصال بقاعدة بيانات Supabase
"""

import os
import sys
import subprocess
import psycopg2
from urllib.parse import urlparse
import time

def load_env_file():
    """تحميل ملف .env"""
    env_path = os.path.join(os.path.dirname(__file__), '..', '.env')
    env_vars = {}
    
    if os.path.exists(env_path):
        with open(env_path, 'r') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#') and '=' in line:
                    key, value = line.split('=', 1)
                    env_vars[key.strip()] = value.strip()
    
    return env_vars

def test_connection_string(connection_string, connection_name):
    """اختبار سلسلة الاتصال"""
    print(f"🔍 جاري اختبار {connection_name}...")
    
    try:
        conn = psycopg2.connect(connection_string)
        conn.close()
        print(f"✅ نجح الاتصال عبر {connection_name}")
        return True
    except Exception as e:
        print(f"❌ فشل الاتصال عبر {connection_name}: {e}")
        return False

def generate_connection_strings(base_url, password, database="postgres"):
    """توليد سلاسل اتصال مختلفة"""
    parsed = urlparse(base_url)
    
    # استخراج المعلومات الأساسية
    if '@' in base_url:
        # تنسيق postgresql://user:password@host:port/database
        user = parsed.username
        host = parsed.hostname
        port = parsed.port or 5432
    else:
        # تنسيق مختلف
        return []
    
    connections = []
    
    # 1. الاتصال الأصلي مع SSL
    connections.append({
        'name': 'Original with SSL',
        'string': f"postgresql://{user}:{password}@{host}:{port}/{database}?sslmode=require"
    })
    
    # 2. الاتصال الأصلي بدون SSL
    connections.append({
        'name': 'Original without SSL',
        'string': f"postgresql://{user}:{password}@{host}:{port}/{database}?sslmode=disable"
    })
    
    # 3. استبدال pooler بـ db
    if 'pooler' in host:
        db_host = host.replace('pooler', 'db')
        connections.append({
            'name': 'Direct DB connection with SSL',
            'string': f"postgresql://{user}:{password}@{db_host}:{port}/{database}?sslmode=require"
        })
        
        connections.append({
            'name': 'Direct DB connection without SSL',
            'string': f"postgresql://{user}:{password}@{db_host}:{port}/{database}?sslmode=disable"
        })
    
    # 4. استخدام المنفذ 6543 (transaction pooler)
    if 'pooler' in host:
        pooler_6543 = host.replace('5432', '6543') if ':5432' in host else host
        connections.append({
            'name': 'Pooler port 6543 with SSL',
            'string': f"postgresql://{user}:{password}@{pooler_6543}:6543/{database}?sslmode=require"
        })
    
    # 5. إضافة timeout
    connections.append({
        'name': 'With connection timeout',
        'string': f"postgresql://{user}:{password}@{host}:{port}/{database}?sslmode=require&connect_timeout=10"
    })
    
    return connections

def update_env_file(connection_string, env_path):
    """تحديث ملف .env بسلسلة الاتصال الناجحة"""
    print(f"📝 تحديث ملف .env...")
    
    with open(env_path, 'r') as f:
        lines = f.readlines()
    
    new_lines = []
    for line in lines:
        if line.startswith('DATABASE_URL='):
            new_lines.append(f'DATABASE_URL={connection_string}\n')
        else:
            new_lines.append(line)
    
    with open(env_path, 'w') as f:
        f.writelines(new_lines)
    
    print(f"✅ تم تحديث ملف .env")

def check_supabase_settings():
    """فحص إعدادات Supabase المقترحة"""
    print("\n📋 إعدادات Supabase المقترحة:")
    print("1. تأكد من تفعيل SSL في إعدادات قاعدة البيانات")
    print("2. تحقق من قواعد الوصول (Database Access Policies)")
    print("3. تأكد من أن IP الخاص بك غير محظور")
    print("4. جرب إضافة IP الخاص بك إلى القائمة البيضاء")
    print("5. تأكد من أن قاعدة البيانات نشطة")

def create_local_fallback():
    """إنشاء قاعدة بيانات محلية كحل بديل"""
    print("\n🔄 إنشاء قاعدة بيانات محلية كحل بديل...")
    
    docker_compose = """
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
"""
    
    docker_compose_path = os.path.join(os.path.dirname(__file__), '..', 'docker-compose.yml')
    with open(docker_compose_path, 'w') as f:
        f.write(docker_compose)
    
    print(f"✅ تم إنشاء ملف docker-compose.yml")
    print(f"📝 لتشغيل قاعدة البيانات المحلية:")
    print(f"   docker-compose up -d")
    print(f"📝 ثم قم بتحديث DATABASE_URL في .env إلى:")
    print(f"   DATABASE_URL=postgresql://partflow:partflow123@localhost:5432/partflow")

def main():
    print("🚀 بدء تشخيص وحل مشاكل الاتصال بقاعدة بيانات Supabase")
    print("=" * 60)
    
    # تحميل المتغيرات البيئية
    env_vars = load_env_file()
    
    if 'DATABASE_URL' not in env_vars:
        print("❌ لم يتم العثور على DATABASE_URL في ملف .env")
        sys.exit(1)
    
    original_url = env_vars['DATABASE_URL']
    print(f"📌 سلسلة الاتصال الأصلية: {original_url}")
    
    # استخراج كلمة المرور
    parsed = urlparse(original_url)
    password = parsed.password
    
    # توليد سلاسل اتصال مختلفة
    connections = generate_connection_strings(original_url, password)
    
    print(f"\n🔍 جاري اختبار {len(connections)} سلاسل اتصال مختلفة...")
    print("=" * 60)
    
    successful_connection = None
    
    for conn in connections:
        if test_connection_string(conn['string'], conn['name']):
            successful_connection = conn
            break
        time.sleep(1)  # تأخير قصير بين المحاولات
    
    if successful_connection:
        print("\n🎉 تم العثور على اتصال ناجح!")
        print(f"✅ السلسلة الناجحة: {successful_connection['name']}")
        print(f"📌 سلسلة الاتصال: {successful_connection['string']}")
        
        # تحديث ملف .env
        env_path = os.path.join(os.path.dirname(__file__), '..', '.env')
        update_env_file(successful_connection['string'], env_path)
        
        print("\n🚀 يمكنك الآن تشغيل التطبيق:")
        print("   cd backend && ./api")
        
    else:
        print("\n❌ لم يتمكن من الاتصال بأي من السلاسل")
        print("\n🔧 الحلول المقترحة:")
        
        check_supabase_settings()
        
        print("\n❓ هل تريد إنشاء قاعدة بيانات محلية كحل بديل؟ (y/n)")
        # في الاستخدام التفاعلي، هنا نطلب من المستخدم
        # للسكربت التلقائي، سنقوم بإنشاء الملف
        
        response = input().lower() if len(sys.argv) > 1 and sys.argv[1] == '--interactive' else 'y'
        
        if response == 'y':
            create_local_fallback()
        else:
            print("\n💡 توصيات إضافية:")
            print("1. تحقق من إعدادات الشبكة لديك")
            print("2. جرب استخدام VPN")
            print("3. تواصل مع دعم Supabase")
            print("4. استخدم قاعدة بيانات محلية للتطوير")

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n⚠️ تم إيقاف السكربت")
        sys.exit(0)
    except Exception as e:
        print(f"\n❌ حدث خطأ: {e}")
        sys.exit(1)
