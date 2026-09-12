; Trust the public internal certificate for this Windows user before the app is launched.
!macro customInstall
  ; The private key never ships with the installer; only the public certificate is imported.
  ExecWait '"$SYSDIR\certutil.exe" -user -addstore "Root" "$INSTDIR\resources\PartFlow-Internal-Code-Signing.cer"'
  ExecWait '"$SYSDIR\certutil.exe" -user -addstore "TrustedPublisher" "$INSTDIR\resources\PartFlow-Internal-Code-Signing.cer"'
!macroend
