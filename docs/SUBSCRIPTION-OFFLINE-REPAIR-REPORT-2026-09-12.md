# تقرير إصلاح سياسة الاشتراك والمزامنة

**التاريخ:** 2026-09-12  
**المرجع:** [SUBSCRIPTION-OFFLINE-STRATEGY.md](../SUBSCRIPTION-OFFLINE-STRATEGY.md)

## الحكم بعد الإصلاح

أصبح مسار التطبيق متوافقاً مع جوهر السياسة في الكود الحالي:

- الحساب والاشتراك يقررهما الخادم السحابي.
- عمليات المتجر تحفظ في SQLite المحلية.
- لا تفتح الجلسة دون اتصال وتحقق سحابي ناجح.
- لا تنفذ عمليات التعديل دون تحقق سحابي جديد.
- السحب يتم من API السحابي إلى SQLite.
- الدفع يرسل عناصر `sync_queue` فقط إلى API سحابي محمي.
- التعارضات تبقى محفوظة محلياً ولا تُحذف العملية عند رفضها.
- لم يعد هناك endpoint لتبديل Online/Offline.

## التغييرات المنفذة

### 1. إصلاح دفع المزامنة

**الملف:** `backend/internal/settings/local_database_handler.go`

كان backend المحلي يقرأ `sync_queue` ثم يكتب مباشرة إلى اتصال PostgreSQL العام. تم استبدال ذلك بـ:

1. قراءة العناصر المعلقة من SQLite المحلية.
2. إرسال دفعة لا تتجاوز 50 عنصراً إلى `/api/v1/sync/push` في الخادم السحابي.
3. تمرير `Authorization` و`X-PartFlow-Cloud-Token`.
4. تعليم العناصر المقبولة فقط بأنها تمت مزامنتها.
5. إبقاء العناصر المرفوضة قابلة لإعادة المحاولة.
6. تسجيل التعارضات في SQLite المحلية.

أضيفت تعليقات قبل كل كتلة مهمة تشرح سبب إبقاء حدود المزامنة عند API السحابي وعدم استخدام اتصال PostgreSQL المحلي كقناة ضمنية.

### 2. إضافة endpoint سحابي للـ queue

**الملفات:**

- `backend/internal/sync/handler.go`
- `backend/internal/sync/service.go`
- `backend/internal/api/router.go`

تمت إضافة `POST /api/v1/sync/push` ويقوم بـ:

- قبول عناصر queue فقط، وليس نسخة SQLite كاملة.
- رفض دفعة فارغة أو أكبر من 50 عنصراً.
- تطبيق العمليات بعد مرورها عبر المصادقة وCloudGuard وAdmin.
- إرجاع `accepted_ids` و`rejected` ونتيجة التعارضات.
- استخدام نفس فحص تعارضات البيانات الموجود في خدمة المزامنة.

أضيفت دالة `ApplyCloudOperation` كي يطبق الخادم العملية دون فتح قاعدة SQLite الخاصة بالجهاز.

### 3. تمرير توكن السحابة أثناء السحب

**الملف:** `backend/internal/settings/local_database_handler.go`

تم إصلاح طلب `sync/initial-data` ليحمل `X-PartFlow-Cloud-Token` بالإضافة إلى JWT المحلي. هذا يمنع رفض السحب الصحيح من CloudGuard ويضمن أن قرار الاشتراك لا يعتمد على JWT المحلي وحده.

### 4. إزالة مسارات Online/Offline القديمة

**الملفات:**

- `backend/internal/api/router.go`
- `backend/internal/settings/local_database_handler.go`

تم حذف:

- `GET /api/v1/settings/operating-mode`
- `PUT /api/v1/settings/operating-mode`
- handlers القديمة التي كانت تحفظ وضع التشغيل وتحتوي مسار seed مباشر من PostgreSQL.

أضيف تعليق في الراوتر يوضح أن SQLite مكان تخزين محلي وليست وضع تشغيل يمكن تبديله لتجاوز التحقق.

### 5. اختبارات جديدة

**الملف:** `backend/internal/sync/handler_test.go`

تمت إضافة اختبارات تضمن أن endpoint الدفع:

- يرفض الدفعة الفارغة.
- يرفض دفعة تتجاوز 50 عملية.

أضيفت تعليقات توضح سبب كل اختبار وعلاقته بحدود الأمان والمزامنة.

## التحقق المنفذ

| الفحص | النتيجة |
|---|---:|
| `go test ./...` من مجلد `backend` | نجح بالكامل |
| `npm run test:run -- --passWithNoTests` | 43 اختباراً ناجحاً |
| `npm run build:check` | نجح typecheck وVite production build |
| اختبارات `internal/sync` بعد الإصلاح | نجحت |
| اختبارات `internal/api` و`internal/settings` بعد الإصلاح | نجحت |
| فحص عدم وجود direct PostgreSQL push في local settings | مؤكد |
| فحص إزالة routes الخاصة بـ operating-mode | مؤكد |

## التحقق من تطبيق السياسة

### فتح التطبيق

`authStore.checkAuth` يرفض الجلسة عند غياب الإنترنت، ثم يطلب `auth/validate` من الخادم السحابي قبل فتح المسارات المحمية.

### العمليات الحساسة

`ApiClient` يمنع `POST` و`PUT` و`PATCH` و`DELETE` عند غياب الإنترنت أو فشل التحقق السحابي قبل إرسال الطلب إلى SQLite API.

### حماية backend

المسارات المحمية تمر عبر JWT middleware ثم CloudGuard. غياب أو رفض `X-PartFlow-Cloud-Token` يمنع الطلب حتى لو كان هناك JWT محلي صالح.

### البيانات المحلية

لا يتم حذف بيانات SQLite عند انتهاء الاشتراك أو تسجيل الخروج. يتم حذف الجلسة، بينما تبقى بيانات المتجر المحلية.

### المزامنة

السحب يمر من cloud API إلى SQLite، والدفع يرسل queue entries فقط. لا يتم رفع ملف SQLite كاملاً.

## إجراء أمني مطلوب خارج الكود

يوجد ملف `backend/.env` محلياً ويحتوي بيانات اتصال بقاعدة Supabase. الملف غير متتبع في Git وموجود ضمن `.gitignore`، لكن إذا كانت كلمة المرور فعالة فيجب تدويرها من Supabase/Render فوراً، ثم تحديث متغيرات البيئة في بيئة التشغيل.

لا يمكن تنفيذ تدوير سر خارجي بأمان من داخل تعديل الكود، لذلك لم يتم تغيير الملف المحلي أو طباعة قيمة السر.

## ملاحظة التشغيل

يجب إعادة تشغيل backend بعد بناء النسخة الجديدة حتى تعمل routes الجديدة وإزالة routes القديمة في العملية العاملة حالياً. لم يتم تنفيذ نشر إلى Render أو بناء Installer موقع في هذه المهمة.

## الخلاصة

تم إصلاح الفجوات البرمجية الأساسية التي ظهرت في الفحص السابق، خصوصاً الكتابة المباشرة إلى PostgreSQL، نقص توكن السحابة في السحب، وبقاء endpoint تغيير وضع التشغيل. نتائج الاختبارات والتجميع كلها ناجحة. الإجراء المتبقي الوحيد خارج المستودع هو تدوير credential قاعدة البيانات والتحقق من إعدادات الأسرار في بيئة الإنتاج.

## ملحق: الشهادة الداخلية

تم إنشاء شهادة Windows داخلية باسم `PartFlow Internal Code Signing` على حساب Windows الحالي.

- thumbprint: `7E53B9CA43FB60FCBA0564D62BA0581394E46818`
- الصلاحية: حتى 2029-09-12.
- المفتاح الخاص بقي داخل `Cert:\CurrentUser\My` وغير قابل للتصدير.
- الشهادة العامة محفوظة خارج المستودع في `C:\Users\Administrator\Desktop\PartFlow-signing\PartFlow-Internal-Code-Signing.cer`.
- تم تفعيل `forceCodeSigning` في `frontend/electron-builder.json`.
- تم ضبط `signtoolOptions.certificateSubjectName` لاستخدام الشهادة من مخزن Windows الحالي.

## نتيجة بناء Windows

تم بناء artifacts داخل:

`C:\Users\Administrator\AppData\Local\Temp\PartFlow-electron-signed`

الملفات الموقعة:

- `PartFlow-0.0.0-setup.exe`
- `PartFlow-0.0.0-portable.exe`
- `win-unpacked\PartFlow.exe`

تحقق Authenticode وجد التوقيع والموقّع والـthumbprint الصحيحين. يعرض Windows حالة `UnknownError` لأن الشهادة Self-Signed وتنتهي سلسلة الثقة في جذر داخلي، وليست شهادة عامة من جهة إصدار معتمدة. هذا مقبول للاستخدام الداخلي بعد تثبيت ملف `.cer` على أجهزة المتاجر، لكنه لا يزيل SmartScreen على أجهزة عامة.

## تثبيت الثقة على جهاز متجر

انسخ ملف الشهادة العامة إلى الجهاز ثم نفّذ PowerShell بحساب المستخدم الذي سيشغّل التطبيق:

```powershell
Import-Certificate -FilePath .\PartFlow-Internal-Code-Signing.cer -CertStoreLocation Cert:\CurrentUser\TrustedPublisher
```

لإدارة عدة أجهزة، وزّع الشهادة عبر Group Policy أو إدارة الأجهزة المؤسسية بدلاً من إرسال المفتاح الخاص. لا ترفع ملف `.pfx` أو المفتاح الخاص إلى GitHub أو Render.

## حدود الحل

هذه ليست شهادة عامة موثوقة من Windows. هي حل مجاني داخلي لمتاجر وأجهزة تحت الإدارة. التوزيع العام ما زال يحتاج شهادة Authenticode مدفوعة أو سيعرض تحذير SmartScreen.
