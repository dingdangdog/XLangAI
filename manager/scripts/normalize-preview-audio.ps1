# 将 storage/audio 下音色试听 MP3 归一化到约 -14 LUFS，便于手机与 PC 在相同媒体音量下听感接近。
# 用法（仓库根或本目录）: .\servers\manager\scripts\normalize-preview-audio.ps1

$ErrorActionPreference = "Stop"
$ffmpeg = $env:XLANGAI_FFMPEG_PATH
if (-not $ffmpeg -or -not (Test-Path $ffmpeg)) {
  $ffmpeg = (Get-Command ffmpeg -ErrorAction SilentlyContinue)?.Source
}
if (-not $ffmpeg) {
  Write-Error "未找到 ffmpeg。请安装并加入 PATH，或设置 XLANGAI_FFMPEG_PATH。"
}

$dir = Join-Path $PSScriptRoot "..\storage\audio" | Resolve-Path
$filter = "highpass=f=80,volume=4dB,loudnorm=I=-14:TP=-1:LRA=8:print_format=none,alimiter=limit=0.97"

Get-ChildItem "$dir\*.mp3" | ForEach-Object {
  $tmp = "$($_.FullName).norm.mp3"
  & $ffmpeg -hide_banner -loglevel error -y -i $_.FullName -af $filter -c:a libmp3lame -q:a 2 $tmp
  if ($LASTEXITCODE -ne 0) { Remove-Item $tmp -ErrorAction SilentlyContinue; throw "ffmpeg failed: $($_.Name)" }
  Move-Item -Force $tmp $_.FullName
  Write-Host "normalized $($_.Name)"
}
