# تقرير بناء PartFlow وإصلاح الأيقونة

**التاريخ:** 2026-09-13  
**الفرع:** `main`  
**آخر commit متعلق بالإصلاح:** `ba665d9 fix: pin shortcuts to branded ico`

## 1. الهدف

تجهيز نسخة Windows موقعة من PartFlow، وإظهار شعار PartFlow الصحيح في:

- صفحة تسجيل الدخول.
- شاشة التحقق من الاشتراك.
- الشريط الجانبي.
- نقطة البيع POS.
- نافذة البرنامج وTray.
- اختصار سطح المكتب.
- قائمة ابدأ.
- Installer والنسخة المحمولة.

## 2. المشكلة الأصلية

كانت الواجهة تستخدم شعاراً نصياً قديماً مثل `PF` في بعض أجزاء نقطة البيع، كما أن شعار PNG كان يُحمّل بمسار مطلق يبدأ بـ `/`. هذا المسار يعمل في Vite داخل المتصفح، لكنه لا يعمل بشكل موثوق داخل Electron عندما تُحمّل الواجهة من `file://`، فظهرت صورة مفقودة في صفحة تسجيل الدخول وبعض الواجهات.

كذلك كانت Windows تستخدم اختصاراً قديماً يشير إلى:

```text
PartFlow.exe,0
```

فاستمر ظهور الأيقونة القديمة في سطح المكتب وقائمة ابدأ بسبب الاختصار وIcon Cache، حتى بعد أن أصبحت الأيقونة المضمنة داخل البرنامج صحيحة.

## 3. الحل المنفذ

### 3.1 شعار الواجهة

تم حفظ الشعار المقدم في:

```text
frontend/public/partflow-logo.png
```

ثم أضيفت نسخة يديرها Vite ضمن:

```text
frontend/src/assets/partflow-logo.png
```

واستُخدم مكوّن مشترك:

```text
frontend/src/components/branding/PartFlowLogo.tsx
```

المكوّن يستورد الصورة من `src/assets`، لذلك يحصل Vite على رابط مجمّع يعمل في التطوير وداخل Electron الإنتاجي.

تم استخدام المكوّن في:

- `frontend/src/components/navigation/sidebar.tsx`
- `frontend/src/features/auth/components/BrandPanel.tsx`
- `frontend/src/features/auth/components/SubscriptionVerificationScreen.tsx`
- `frontend/src/features/sales/components/modern/ModernPOSLayout.tsx`
- `frontend/src/features/sales/pages/POSPage.tsx`

وبذلك استُبدل شعار `PF` القديم في نقطة البيع بالشعار الحقيقي.

### 3.2 أيقونة Windows متعددة الأحجام

تم إنشاء ملف ICO من الشعار باستخدام:

```text
scripts/generate-partflow-icon.ps1
```

الملف الناتج:

```text
frontend/public/partflow-logo.ico
```

ويحتوي على الأحجام:

```text
16x16, 24x24, 32x32, 48x48, 64x64, 128x128, 256x256
```

### 3.3 إعداد Electron Builder

تم ضبط `frontend/electron-builder.json` لاستخدام:

```json
"icon": "public/partflow-logo.ico"
```

كما تُنسخ ملفات PNG وICO إلى موارد التطبيق:

```text
resources/partflow-logo.png
resources/partflow-logo.ico
```

ويُستخدم ICO متعدد الأحجام لأيقونة Windows بدلاً من الأيقونة الافتراضية لـ Electron.

### 3.4 اختصارات سطح المكتب وقائمة ابدأ

تم تحديث:

```text
frontend/build/installer.nsh
```

ليقوم Installer بـ:

1. حذف الاختصارات القديمة.
2. إنشاء اختصار سطح المكتب من جديد.
3. إنشاء اختصار داخل قائمة ابدأ.
4. جعل كلا الاختصارين يستخدمان مباشرة:

```text
$INSTDIR\resources\partflow-logo.ico
```

هذا يمنع Windows من الاعتماد على Icon Cache قديم أو على الأيقونة الافتراضية داخل executable.

## 4. أمر توليد ICO

من جذر المشروع:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\generate-partflow-icon.ps1
```

## 5. أمر بناء Installer

من مجلد `frontend`:

```powershell
$output = 'C:\Users\Administrator\AppData\Local\Temp\PartFlow-one-click'
Remove-Item -Recurse -Force $output -ErrorAction SilentlyContinue
npm run build
npx electron-builder --win --publish never --config.directories.output=$output
```

ينتج البناء:

```text
C:\Users\Administrator\AppData\Local\Temp\PartFlow-one-click\PartFlow-0.0.0-setup.exe
C:\Users\Administrator\AppData\Local\Temp\PartFlow-one-click\PartFlow-0.0.0-portable.exe
```

## 6. التحقق النهائي

تم التحقق من الآتي:

- `npm run build:check` نجح.
- Installer بُني بنجاح.
- النسخة المحمولة بُنيت بنجاح.
- حالة توقيع Installer كانت `Valid`.
- الموقّع:

```text
CN=PartFlow Internal Code Signing
```

- ملف ICO يحتوي سبعة أحجام.
- ملف ICO موجود داخل موارد التطبيق.
- الشعار المجمّع موجود داخل `app.asar` كملف Vite مُجزّأ.
- `PartFlow.exe` المثبت يحمل الأيقونة الصحيحة.
- اختصار سطح المكتب وقائمة ابدأ يستخدمان مسار ICO الخارجي الصحيح.
- تم تحديث Windows Icon Cache وإعادة تشغيل Explorer أثناء إصلاح الجهاز الحالي.

## 7. commits ذات الصلة

```text
10edb65 fix: use tracked logo asset for electron builds
567e2c8 feat: add multi-resolution Windows app icon
7f7b8d6 fix: load logo in packaged electron renderer
e6c4a5b fix: bundle logo asset for electron renderer
6822341 fix: use branded logo in point of sale
ba665d9 fix: pin shortcuts to branded ico
```

## 8. ملاحظات التوزيع

- استخدم Installer الموجود في مسار `PartFlow-one-click` بعد آخر بناء.
- لا تستخدم Installer أقدم، لأنه لا يحتوي إصلاح مسار الشعار أو اختصارات ICO.
- لا تحذف `%APPDATA%\PartFlow\data\partflow.db` عند تحديث البرنامج؛ فهي تحتوي قاعدة SQLite المحلية.
- مجلدات البناء مثل `frontend/dist-electron-internal` كبيرة ومقصود أن تبقى خارج Git.
- لا تُرفع المفاتيح الخاصة أو ملفات الأسرار إلى GitHub.
