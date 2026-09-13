# تقرير تحقق سياسة الاشتراك وتغليف Windows

**التاريخ:** 2026-09-13  
**الفرع:** `main`  
**آخر commit مصدر تم اختباره:** `327eb47`

## النتيجة المختصرة

نجحت اختبارات سياسة الاشتراك المحلية، واختبارات المشروع، وبناء واجهة Electron، وتغليف Windows. لا توجد تغييرات مصدرية جديدة ناتجة عن الاختبار.

## اختبارات سياسة الاشتراك

### Backend

تم تشغيل:

```powershell
go test ./...
```

النتيجة: نجحت جميع حزم Backend، بما فيها:

- `internal/auth`
- `pkg/middleware`
- `internal/sync`
- `internal/settings`
- `internal/localdb`
- `internal/products`
- `internal/barcodes`

وتغطي الاختبارات المتاحة التحقق السحابي، رفض الجلسة المحلية دون cloud token، إعدادات release، دمج snapshot السحابي في SQLite، وقائمة `sync_queue` والتعارضات.

### Frontend وElectron

تم تشغيل:

```powershell
npm run test:run
npm run lint:types
node --check electron/main.js
node --check electron/preload.js
npm run build:check
```

النتيجة:

- 13 ملف اختبار ناجح.
- 43 اختبارًا ناجحًا.
- فحص TypeScript ناجح.
- ملفات Electron صالحة نحويًا.
- بناء Vite وPWA ناجح.

السلوك المتحقق في الكود:

- لا تُستعاد جلسة محفوظة دون cloud token واتصال إنترنت.
- يتم التحقق من الاشتراك عند بدء التطبيق.
- انقطاع الاتصال يغلق الجلسة ويعيدها إلى تسجيل الدخول.
- عمليات `POST` و`PUT` و`PATCH` و`DELETE` تمر عبر تحقق سحابي قبل التنفيذ.
- بيانات المتجر التشغيلية تبقى في SQLite المحلية.
- المزامنة السحابية يدوية وليست رفعًا تلقائيًا لقاعدة SQLite كاملة.

## تغليف Windows

تم تنفيذ بناء نظيف في:

```text
C:\Users\Administrator\AppData\Local\Temp\PartFlow-policy-release
```

الأمر المستخدم:

```powershell
npm run build
npx electron-builder --win --publish never --config.directories.output=C:\Users\Administrator\AppData\Local\Temp\PartFlow-policy-release
```

المخرجات:

```text
PartFlow-0.0.0-setup.exe
PartFlow-0.0.0-portable.exe
win-unpacked\PartFlow.exe
```

التحقق:

- توقيع Installer: `Valid`.
- توقيع Portable: `Valid`.
- توقيع `PartFlow.exe`: `Valid`.
- الشهادة: `CN=PartFlow Internal Code Signing`.
- `resources\partflow-logo.ico` موجود.
- `resources\backend\partflow-api.exe` موجود.
- إعداد `deleteAppDataOnUninstall` ما زال `false`.
- SQLite والصور تستخدم مسارات `%APPDATA%\PartFlow\data` خارج مجلد التثبيت.

## ما لا يمكن إثباته محليًا

- لا يمكن اختبار قرار حساب Render الحقيقي دون تنفيذ تحقق على حساب سحابي فعلي.
- لا يمكن اختبار ماسح USB أو Bluetooth فعلي دون توصيل الجهاز.
- يجب بعد نشر Render التأكد من نجاح Deploy ثم تجربة تسجيل الدخول والتحقق من الاشتراك.
- يجب على صاحب المتجر تجربة Installer على جهاز Windows مستقل، ثم تحديثه، والتأكد من بقاء:
  - `%APPDATA%\PartFlow\data\partflow.db`
  - `%APPDATA%\PartFlow\data\product-images`
  - `%APPDATA%\PartFlow\data\category-images`
  - `%APPDATA%\PartFlow\data\part-type-images`

## المسار المقترح للتسليم

1. انتظار اكتمال Deploy في Render من آخر commit على `main`.
2. تجربة تسجيل الدخول من نسخة Windows.
3. تجربة إنشاء منتج وباركود من مسار التصنيف.
4. تجربة البيع بالماسح إذا كان الجهاز متاحًا.
5. تثبيت النسخة على جهاز اختبار ثم تنفيذ تحديث فوقها دون حذف بيانات المستخدم.
