# تشغيل PartFlow محلياً

## المتطلبات

- Go مثبت ومتاح من الطرفية.
- Node.js و npm مثبتان.
- ملف `backend/.env` موجود.
- اعتماديات الواجهة مثبتة داخل `frontend/node_modules`.

## تشغيل الـ Backend

افتح طرفية PowerShell جديدة من مجلد المشروع ونفّذ:

```powershell
Set-Location .\backend
$env:Path += ";$env:USERPROFILE\go\bin"
air -c .air.toml
```

يستخدم الخادم أداة `Air` لإعادة البناء والتشغيل تلقائيًا عند حفظ أي ملف Go. لا حاجة لإيقاف الخادم وتشغيله يدويًا بعد كل تعديل.

### تثبيت Air على Windows

نفّذ الأمر التالي مرة واحدة:

```powershell
go install github.com/air-verse/air@latest
```

إذا كانت أداة `air` مثبتة لكن غير معروفة في الطرفية، أضف مجلد أدوات Go إلى المسار:

```powershell
$env:Path += ";$env:USERPROFILE\go\bin"
```

يمكن أيضًا تشغيل Air من داخل `backend` عبر هدف Makefile:

```powershell
make dev
```

يتطلب ذلك توفر GNU Make. إعدادات المراقبة محفوظة في `backend/.air.toml`.

يعمل الـ API على:

```text
http://localhost:8080/
```

## تشغيل الـ Frontend

افتح طرفية PowerShell ثانية من مجلد المشروع ونفّذ:

```powershell
Set-Location .\frontend
npm run dev -- --host 0.0.0.0
```

تفتح الواجهة على:

```text
http://localhost:5174/
```

## فحص الحالة

من طرفية ثالثة:

```powershell
Invoke-WebRequest -UseBasicParsing http://localhost:8080/health
```

يجب أن تكون النتيجة `200 OK` وأن تظهر قاعدة البيانات بالحالة `healthy`.

## إيقاف الخدمات

في طرفية كل خدمة اضغط:

```text
Ctrl+C
```
