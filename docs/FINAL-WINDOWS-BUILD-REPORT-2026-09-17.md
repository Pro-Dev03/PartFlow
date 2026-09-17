# تقرير بناء وإصدار PartFlow لنظام Windows

**التاريخ:** 2026-09-17  
**الإصدار:** 1.0.0  
**الفرع:** `main`  
**Commit المصدر:** `22a0283c80757ffb2106b252fe56db6b4bdd412b`

## الملخص

تم بناء PartFlow كبرنامج Windows تنفيذي باستخدام Electron وElectron Builder. نجح البناء في إنشاء Installer ونسخة Portable ونسخة `win-unpacked`، كما تم توقيع الملفات التنفيذية بشهادة Code Signing داخلية.

## ما تم تنفيذه

### سكربت البناء

تم تحسين `scripts/build-win-final-installer.ps1` ليقوم بالآتي:

- بناء Backend بصيغة Windows executable.
- استخدام مجلد Frontend مؤقت ونظيف لتجنب أقفال `node_modules`.
- إيقاف العملية عند فشل أي أمر.
- استخدام رقم الإصدار فعليًا.
- إنشاء إعداد Electron Builder مؤقت يربط Backend المبني الحالي.
- إنشاء Installer وPortable و`win-unpacked`.
- نسخ Backend داخل الحزمة.
- توقيع Backend المضمّن يدويًا.
- إنشاء `release-manifest.json` يتضمن SHA-256 للملفات.

### شهادة التوقيع

تم استخدام الشهادة الداخلية:

```text
CN=PartFlow Internal Code Signing
```

وتم التحقق من توقيع الملفات التالية:

- `PartFlow-1.0.0-setup.exe`
- `PartFlow-1.0.0-portable.exe`
- `partflow-api.exe`
- `win-unpacked\PartFlow.exe`
- `win-unpacked\resources\backend\partflow-api.exe`

حالة التوقيع: `Valid`.

### تثبيت الشهادة تلقائيًا

ملف `frontend/build/installer.nsh` مربوط بإعداد NSIS في `frontend/electron-builder.json`، ويقوم أثناء التثبيت بإضافة الشهادة العامة للمستخدم الحالي إلى:

```text
Root
TrustedPublisher
```

المفتاح الخاص لا يدخل الحزمة ولا يتم تثبيته على جهاز المتجر.

### البيانات والصور

تبقى بيانات المستخدم خارج مجلد تثبيت البرنامج:

```text
%APPDATA%\PartFlow\data\partflow.db
%APPDATA%\PartFlow\data\product-images\
%APPDATA%\PartFlow\data\part-type-images\
%APPDATA%\PartFlow\data\category-images\
```

إعداد NSIS يحتوي على:

```json
"deleteAppDataOnUninstall": false
```

لذلك لا يُفترض أن يحذف التحديث قاعدة البيانات أو الصور.

## مخرجات البناء

المجلد النهائي:

```text
dist\windows-release\partflow-bundle\
```

ويحتوي على:

```text
PartFlow-1.0.0-setup.exe
PartFlow-1.0.0-portable.exe
partflow-api.exe
launch-partflow.bat
README.txt
release-manifest.json
win-unpacked\
```

## SHA-256

```text
PartFlow-1.0.0-setup.exe
B6209DE69D266421300F8807203D850010D82B59C55E8741BA98FFA622F4E821

PartFlow-1.0.0-portable.exe
FAE4A7C327B878E1B372F8DDC5A34FACDE2F7A418358D5F3EE103652B792375E
```

## التحقق المنفذ

نجح ما يلي:

- بناء Vite وPWA.
- تثبيت اعتماديات Frontend في staging نظيف.
- تغليف Electron Builder.
- إنشاء Installer وPortable.
- وجود Backend داخل `resources\backend`.
- توقيع الملفات التنفيذية المطلوبة.
- إنشاء manifest وSHA-256.
- فحص صياغة PowerShell و`git diff --check` في مراحل الإصلاح.

## ملاحظة اختبار التشغيل

تم تشغيل `win-unpacked\PartFlow.exe` تجريبيًا، لكن endpoint `/health` لم يعطِ استجابة خلال نافذة الاختبار البالغة خمس ثوانٍ. لذلك لم يتم اعتبار اختبار التشغيل الكامل ناجحًا، ويجب تنفيذ اختبار إضافي على جهاز Windows مستقل يشمل:

1. تشغيل Installer.
2. التأكد من تثبيت الشهادة في مخزني المستخدم.
3. تسجيل الدخول.
4. إنشاء منتج وحفظ صورة.
5. إعادة تشغيل البرنامج والتأكد من بقاء البيانات والصور.
6. تثبيت تحديث فوق النسخة السابقة والتأكد من عدم حذف البيانات.

## حالة الإصدار

**Build status: SUCCESS**  
**Artifact validation: SUCCESS**  
**Code signing validation: SUCCESS**  
**Clean-machine installation test: PENDING**  
**Runtime smoke test: PENDING / لم يكتمل محليًا**

الحزمة جاهزة للاختبار على جهاز Windows مستقل، لكنها لا تُعلن Release عامة نهائية قبل إكمال اختبار التثبيت والتشغيل وحفظ البيانات على جهاز نظيف.
