# PartFlow Worker — إعداد وتشغيل

## ما هو Worker؟

Worker خدمة خلفية تشغّل المهام الدورية التي لا يجب أن تعطل طلبات الـ API:

- إنهاء الحجوزات المنتهية كل 5 دقائق.
- تحويل الديون المستحقة إلى حالة `overdue` كل ساعة.
- فحص المخزون المنخفض كل 30 دقيقة.
- إنشاء ملخص يومي للإشعارات كل 24 ساعة.

التنفيذ الفعلي الوحيد موجود في `backend/cmd/worker`.

## إعداد Render

يُعرَّف Worker في `render.yaml` كخدمة Background Worker، ويُبنى باستخدام:

```text
Dockerfile: ./worker/Dockerfile
Docker context: .
Binary: backend/cmd/worker
```

لا يحتاج Worker إلى منفذ HTTP أو رابط عام أو اتصال قاعدة بيانات مستقل. يجب أن تكون القيمتان التاليتان مطابقتين لخدمة Backend:

```text
DATABASE_URL  # نفس قاعدة البيانات التي يستخدمها Backend
JWT_SECRET    # يُسحب تلقائيًا من متغير Backend نفسه في Render
```

`REDIS_URL` موجود في القالب للتوافق المستقبلي، لكن Worker الحالي لا يستخدم Redis.

إذا كان Backend متصلاً بـ Supabase، فاضبط `DATABASE_URL` في Worker على نفس سلسلة الاتصال المستخدمة فعليًا في Backend، ولا تنشئ قاعدة ثانية.

## التحقق قبل النشر

من مجلد `backend`:

```powershell
go test ./cmd/worker
go test ./...
```

بعد دفع التغييرات، أعد نشر خدمتي `partflow-backend` و`partflow-worker` في Render. راقب سجل Worker للتأكد من ظهور:

```text
Worker service started successfully
```

لا يمكن تشغيل المهام الفعلية إذا كان `DATABASE_URL` ناقصًا أو مختلفًا عن قاعدة Backend.

## التشغيل المحلي

```powershell
cd backend
go run ./cmd/worker
```

يستخدم التشغيل المحلي نفس متغيرات البيئة الموجودة في `backend/.env`، لذلك لا تضع أسرارًا داخل المستودع أو داخل نسخة Electron.
