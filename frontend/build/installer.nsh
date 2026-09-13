; Trust the public internal certificate for this Windows user before the app is launched.
!macro customInstall
  ; The private key never ships with the installer; only the public certificate is imported.
  ExecWait '"$SYSDIR\certutil.exe" -user -addstore "Root" "$INSTDIR\resources\PartFlow-Internal-Code-Signing.cer"'
  ExecWait '"$SYSDIR\certutil.exe" -user -addstore "TrustedPublisher" "$INSTDIR\resources\PartFlow-Internal-Code-Signing.cer"'

  ; Point shortcuts directly at the external multi-size ICO to avoid stale Windows icon cache entries.
  Delete "$DESKTOP\PartFlow.lnk"
  Delete "$DESKTOP\PartFlow - Shortcut.lnk"
  CreateShortCut "$DESKTOP\PartFlow.lnk" "$INSTDIR\PartFlow.exe" "" "$INSTDIR\resources\partflow-logo.ico" 0 SW_SHOWNORMAL "" "PartFlow"
  CreateDirectory "$SMPROGRAMS\PartFlow"
  Delete "$SMPROGRAMS\PartFlow\PartFlow.lnk"
  CreateShortCut "$SMPROGRAMS\PartFlow\PartFlow.lnk" "$INSTDIR\PartFlow.exe" "" "$INSTDIR\resources\partflow-logo.ico" 0 SW_SHOWNORMAL "" "PartFlow"
!macroend
