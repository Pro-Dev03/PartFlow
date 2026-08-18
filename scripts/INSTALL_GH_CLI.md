# تثبيت GitHub CLI

لقد قمت بإنشاء عدة سكريبتات لتثبيت GitHub CLI، ولكن نظراً لقيود بيئة Git Bash، يفضل التثبيت اليدوي.

## الطريقة الأسهل (الموصى بها)

### 1. تثبيت باستخدام Winget (موجود في Windows 10/11)
افتح PowerShell أو CMD كـ Administrator وقم بتشغيل:

```powershell
winget install --id GitHub.cli
```

### 2. تثبيت يدوياً
1. قم بزيارة: https://github.com/cli/cli/releases/latest
2. حمل الإصدار: `gh_2.51.0_windows_amd64.zip` (أو أحدث)
3. فك الضغط إلى مجلد مثلاً: `C:\Tools\gh`
4. أضف المجلد إلى PATH:
   - ابحث عن "Environment Variables" في Windows
   - حرر "Path" وأضف `C:\Tools\gh`

## بعد التثبيت

1. أعد تشغيل الطرفية (Terminal)
2. تحقق من التثبيت:
   ```bash
   gh --version
   ```

3. سجل الدخول بحساب GitHub:
   ```bash
   gh auth login
   ```

4. اتبع التعليمات:
   - اختر GitHub.com
   - اختر HTTPS
   - اختر Login with a web browser
   - أدخل الكود الذي سيظهر
   - اضغط Authorize في المتصفح

## السكريبتات المتاحة

يمكنك تجربة السكريبتات التالية إذا كنت تريد التثبيت التلقائي:

- `scripts/install-gh-cli.sh` - لـ Git Bash
- `scripts/install-gh-cli.ps1` - لـ PowerShell
- `scripts/install-gh-cli-simple.ps1` - لـ PowerShell (بسيط)
- `scripts/install-gh-cli.bat` - لـ CMD

لتشغيل أي سكريبت PowerShell كـ Administrator:
```powershell
powershell.exe -ExecutionPolicy Bypass -File scripts/install-gh-cli-simple.ps1
```

## الاستخدام بعد التثبيت

```bash
# تسجيل الدخول
gh auth login

# التحقق من الحالة
gh auth status

# إنشاء repository جديد
gh repo create my-repo --public

# استنساخ repository
gh repo clone owner/repo

# عرض issues
gh issue list

# إنشاء issue جديد
gh issue create --title "Bug fix" --body "Description"
```
