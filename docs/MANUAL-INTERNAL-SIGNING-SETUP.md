# دليل إنشاء توقيع داخلي لـ PartFlow يدوياً

**الغرض:** إنشاء شهادة Windows داخلية مجانية، توقيع Installer الخاص بـElectron، وتجهيز Installer واحد يثبت التطبيق والشهادة على أجهزة متاجر محددة.

**النطاق:** هذا الحل للاستخدام الداخلي فقط. لا يمنح ثقة عامة على أجهزة Windows ولا يلغي SmartScreen عند التوزيع العام.

## 1. المتطلبات

على جهاز البناء:

- Windows PowerShell.
- Node.js وnpm.
- اعتماديات frontend مثبتة (`npm install` إذا لزم).
- ملف backend التنفيذي في `dist/windows-release/partflow-api.exe`.
- قاعدة البيانات المحلية في `backend/partflow-local.db`.
- صلاحية تشغيل `New-SelfSignedCertificate`.

تحقق من الأدوات:

```powershell
Get-Command New-SelfSignedCertificate
node --version
npm --version
```

## 2. إنشاء مجلد خاص بالتوقيع

لا تحفظ المفتاح الخاص داخل المشروع أو GitHub. استخدم مجلداً خارج المستودع:

```powershell
$signingDir = 'C:\Users\Administrator\Desktop\PartFlow-signing'
New-Item -ItemType Directory -Force -Path $signingDir | Out-Null
```

غيّر المسار بما يناسب جهازك.

## 3. إنشاء شهادة Code Signing داخلية

الأمر التالي ينشئ شهادة مدتها ثلاث سنوات، ويضع المفتاح الخاص في مخزن المستخدم الحالي مع منع تصديره:

```powershell
$cert = New-SelfSignedCertificate `
  -Type CodeSigningCert `
  -Subject 'CN=PartFlow Internal Code Signing' `
  -FriendlyName 'PartFlow Internal Code Signing' `
  -CertStoreLocation 'Cert:\CurrentUser\My' `
  -NotAfter (Get-Date).AddYears(3) `
  -KeyExportPolicy NonExportable
```

تحقق من الشهادة:

```powershell
$cert | Select-Object Subject, Thumbprint, HasPrivateKey, NotAfter
```

احتفظ بقيمة `Thumbprint`. في هذه النسخة كان الاسم:

```text
PartFlow Internal Code Signing
```

## 4. تصدير الشهادة العامة فقط

صدّر ملف `.cer`، وهو يحتوي الشهادة العامة فقط ولا يحتوي المفتاح الخاص:

```powershell
$publicCert = Join-Path $signingDir 'PartFlow-Internal-Code-Signing.cer'
Export-Certificate -Cert $cert -FilePath $publicCert -Force | Out-Null
```

لا تصدّر ملف `.pfx` ولا تنسخ المفتاح الخاص إلى جهاز المتجر.

## 5. تثبيت الثقة على جهاز البناء

لتتمكن أدوات Windows وElectron Builder من التحقق من السلسلة محلياً:

```powershell
Import-Certificate `
  -FilePath $publicCert `
  -CertStoreLocation 'Cert:\CurrentUser\Root' | Out-Null

Import-Certificate `
  -FilePath $publicCert `
  -CertStoreLocation 'Cert:\CurrentUser\TrustedPublisher' | Out-Null
```

يجب تنفيذ ذلك للمستخدم الذي سيبني أو يختبر التطبيق. لا تستخدم مخزن `LocalMachine` إلا إذا كنت تدير أجهزة المؤسسة بصلاحيات إدارية.

## 6. وضع الشهادة العامة في موارد المشروع

انسخ الشهادة العامة إلى موارد Electron:

```powershell
New-Item -ItemType Directory -Force -Path .\frontend\build | Out-Null
Copy-Item $publicCert .\frontend\build\PartFlow-Internal-Code-Signing.cer -Force
```

هذا الملف العام يمكن رفعه إلى GitHub. المفتاح الخاص لا يرفع أبداً.

## 7. إعداد Electron Builder

في `frontend/electron-builder.json` يجب أن توجد الإعدادات التالية:

```json
{
  "directories": {
    "output": "dist-electron-internal",
    "buildResources": "build"
  },
  "extraResources": [
    {
      "from": "build/PartFlow-Internal-Code-Signing.cer",
      "to": "PartFlow-Internal-Code-Signing.cer"
    }
  ],
  "win": {
    "target": [
      { "target": "nsis", "arch": ["x64"] },
      { "target": "portable", "arch": ["x64"] }
    ],
    "signAndEditExecutable": true,
    "signtoolOptions": {
      "certificateSubjectName": "PartFlow Internal Code Signing"
    },
    "forceCodeSigning": true
  },
  "nsis": {
    "oneClick": false,
    "allowToChangeInstallationDirectory": true,
    "deleteAppDataOnUninstall": false,
    "include": "build/installer.nsh",
    "artifactName": "PartFlow-${version}-setup.exe"
  }
}
```

ملاحظات:

- `certificateSubjectName` يجب أن يكون داخل `signtoolOptions`.
- `forceCodeSigning: true` يمنع إنشاء نسخة صامتة غير موقعة.
- لا تستخدم `certificateStore` مع Electron Builder 26.15.3؛ هذا الخيار غير مقبول في schema المستخدم هنا.
- لا تضع مسار المفتاح الخاص أو كلمة مروره في JSON.

## 8. إعداد تثبيت الشهادة تلقائياً

أنشئ `frontend/build/installer.nsh`:

```nsh
; Trust the public internal certificate for this Windows user before the app is launched.
!macro customInstall
  ; The private key never ships with the installer; only the public certificate is imported.
  ExecWait '"$SYSDIR\certutil.exe" -user -addstore "Root" "$INSTDIR\resources\PartFlow-Internal-Code-Signing.cer"'
  ExecWait '"$SYSDIR\certutil.exe" -user -addstore "TrustedPublisher" "$INSTDIR\resources\PartFlow-Internal-Code-Signing.cer"'
!macroend
```

هذا hook يعمل أثناء تثبيت NSIS ويضيف الشهادة العامة إلى مخزني المستخدم الحالي. لا يثبت المفتاح الخاص.

## 9. بناء Installer

من جذر المشروع:

```powershell
Set-Location .\frontend
npm install
npm run build
npx electron-builder --win --publish never
```

أو استخدم script المشروع:

```powershell
npm run build:windows:installer
```

إذا ظهر خطأ `EPERM` أثناء استخراج Electron في مجلد `dist-electron-internal`، استخدم مجلد إخراج مؤقتاً جديداً:

```powershell
$output = Join-Path $env:TEMP 'PartFlow-one-click'
Remove-Item -Recurse -Force $output -ErrorAction SilentlyContinue
npm run build
npx electron-builder --win --publish never --config.directories.output=$output
```

في بيئة البناء الحالية نجح البناء بهذا المسار:

```text
C:\Users\Administrator\AppData\Local\Temp\PartFlow-one-click
```

## 10. التحقق من التوقيع

تحقق من Installer والنسخة المحمولة:

```powershell
$output = 'C:\Users\Administrator\AppData\Local\Temp\PartFlow-one-click'

Get-ChildItem $output -Filter '*.exe' | ForEach-Object {
  $signature = Get-AuthenticodeSignature $_.FullName
  [PSCustomObject]@{
    File = $_.Name
    Status = $signature.Status
    Signer = if ($signature.SignerCertificate) { $signature.SignerCertificate.Subject } else { 'NONE' }
    Thumbprint = if ($signature.SignerCertificate) { $signature.SignerCertificate.Thumbprint } else { '' }
  }
}
```

تحقق من وجود الشهادة داخل الحزمة المفكوكة:

```powershell
Test-Path "$output\win-unpacked\resources\PartFlow-Internal-Code-Signing.cer"
```

النتيجة يجب أن تكون `True`، ويجب أن يظهر الموقّع:

```text
CN=PartFlow Internal Code Signing
```

لأن الشهادة Self-Signed، قد يعرض `Get-AuthenticodeSignature` الحالة `UnknownError` بسبب سلسلة الثقة الداخلية. هذا لا يعني أن التوقيع غير موجود؛ افحص `Signer` و`Thumbprint` وتأكد من تثبيت الشهادة على جهاز الاختبار.

## 11. تثبيت التطبيق على جهاز صاحب المتجر

أرسل فقط:

```text
PartFlow-0.0.0-setup.exe
```

لا ترسل المفتاح الخاص أو مجلد الشهادات الخاص. شغّل Installer بحساب المستخدم الذي سيستخدم التطبيق. سيقوم NSIS تلقائياً بـ:

1. تثبيت التطبيق.
2. تثبيت backend وقاعدة SQLite المحلية.
3. إضافة الشهادة العامة إلى `CurrentUser\Root`.
4. إضافة الشهادة العامة إلى `CurrentUser\TrustedPublisher`.

إذا أردت تثبيت الشهادة يدوياً بدلاً من Installer:

```powershell
Import-Certificate `
  -FilePath .\PartFlow-Internal-Code-Signing.cer `
  -CertStoreLocation Cert:\CurrentUser\Root

Import-Certificate `
  -FilePath .\PartFlow-Internal-Code-Signing.cer `
  -CertStoreLocation Cert:\CurrentUser\TrustedPublisher
```

## 12. اختبار جهاز المتجر

بعد التثبيت:

1. افتح PartFlow.
2. تأكد من وجود اتصال بالإنترنت.
3. سجّل الدخول بحساب المتجر.
4. تحقق من ظهور لوحة التحكم.
5. نفّذ عملية بيع تجريبية صغيرة.
6. تحقق من المزامنة إذا كان الحساب مالكاً أو إدارياً.
7. أعد تشغيل التطبيق وتأكد من بقاء قاعدة SQLite المحلية.
8. اختبر رفض العملية عند فصل الإنترنت.

## 13. ما يجب عدم رفعه إلى GitHub

ممنوع رفع:

- ملف `.pfx` أو `.p12`.
- المفتاح الخاص.
- كلمة مرور الشهادة.
- مجلد `Cert:\CurrentUser\My`.
- مجلدات build الكبيرة وملفات Installer، إلا إذا كان ذلك مقصوداً عبر Releases.
- أي ملف `.env` يحتوي كلمات مرور أو مفاتيح.

المسموح رفعه:

- ملف `.cer` العام.
- `installer.nsh`.
- `electron-builder.json` دون أسرار.
- هذا التقرير.

## 14. تجديد الشهادة

قبل انتهاء الشهادة:

1. أنشئ شهادة جديدة باسم مختلف أو بنفس الاسم مع thumbprint جديد.
2. صدّر `.cer` العام.
3. استبدل ملف `frontend/build/PartFlow-Internal-Code-Signing.cer`.
4. حدّث `certificateSubjectName` إذا تغير الاسم.
5. ثبّت الشهادة الجديدة على جهاز البناء.
6. ابنِ Installer جديداً.
7. تحقق من thumbprint والتوقيع.
8. وزّع النسخة الجديدة على الأجهزة.

## 15. حدود الأمان

هذه الشهادة تثق بها الأجهزة التي تثبت الشهادة العامة فقط. لا تمنح ثقة عامة في Windows، ولا تمنع SmartScreen على أجهزة غير مُدارة، ولا تحمي من تعديل التطبيق إذا امتلك المستخدم الجهاز بالكامل. للتوزيع العام تحتاج شهادة Authenticode من جهة إصدار موثوقة.

## 16. الحالة الحالية في المستودع

آخر إعداد منشور هو commit:

```text
c8651ca build: create one-click internal certificate installer
```

الملفات المتعلقة بالتنفيذ:

- `frontend/electron-builder.json`
- `frontend/build/installer.nsh`
- `frontend/build/PartFlow-Internal-Code-Signing.cer`
- `docs/SUBSCRIPTION-OFFLINE-REPAIR-REPORT-2026-09-12.md`
