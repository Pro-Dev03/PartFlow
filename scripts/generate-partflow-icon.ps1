param(
  [string]$Source = (Join-Path $PSScriptRoot '..\frontend\public\partflow-logo.png'),
  [string]$Output = (Join-Path $PSScriptRoot '..\frontend\public\partflow-logo.ico')
)

Add-Type -AssemblyName System.Drawing

$sizes = @(16, 24, 32, 48, 64, 128, 256)
$sourceImage = [System.Drawing.Image]::FromFile((Resolve-Path $Source))
$tempFiles = @()

try {
  $images = @()
  foreach ($size in $sizes) {
    $bitmap = New-Object System.Drawing.Bitmap($size, $size)
    $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
    $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
    $graphics.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
    $graphics.Clear([System.Drawing.Color]::Transparent)
    $graphics.DrawImage($sourceImage, 0, 0, $size, $size)
    $graphics.Dispose()

    $temp = [System.IO.Path]::GetTempFileName()
    $bitmap.Save($temp, [System.Drawing.Imaging.ImageFormat]::Png)
    $bitmap.Dispose()
    $tempFiles += $temp
    $images += ,([System.IO.File]::ReadAllBytes($temp))
  }

  $header = New-Object System.Collections.Generic.List[byte]
  $header.AddRange([byte[]](0, 0, 1, 0, [byte]$sizes.Count, 0))
  $offset = 6 + (16 * $sizes.Count)
  $entries = New-Object System.Collections.Generic.List[byte]
  $payload = New-Object System.Collections.Generic.List[byte]

  for ($index = 0; $index -lt $sizes.Count; $index++) {
    $size = $sizes[$index]
    $bytes = $images[$index]
    $dimension = if ($size -eq 256) { 0 } else { $size }
    $entries.AddRange([byte[]]($dimension, $dimension, 0, 0, 1, 0, 32, 0))
    $length = $bytes.Length
    $entries.AddRange([BitConverter]::GetBytes([uint32]$length))
    $entries.AddRange([BitConverter]::GetBytes([uint32]$offset))
    $payload.AddRange([byte[]]$bytes)
    $offset += $length
  }

  $outputPath = [System.IO.Path]::GetFullPath($Output)
  [System.IO.Directory]::CreateDirectory([System.IO.Path]::GetDirectoryName($outputPath)) | Out-Null
  [System.IO.File]::WriteAllBytes($outputPath, $header.ToArray() + $entries.ToArray() + $payload.ToArray())
  Write-Output "Created $outputPath with sizes: $($sizes -join ', ')"
}
finally {
  $sourceImage.Dispose()
  foreach ($temp in $tempFiles) {
    Remove-Item $temp -Force -ErrorAction SilentlyContinue
  }
}
