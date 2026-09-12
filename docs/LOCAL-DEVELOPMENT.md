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
go run .\cmd\api
```

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
