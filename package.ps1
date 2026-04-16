$root = $PSScriptRoot
$src  = "$root\src"
$out  = "$root\dungeoneer.zip"

Remove-Item -Force $out -ErrorAction SilentlyContinue

$files = @(
  @{ src = "$src\dungeoneer.exe";     dst = "dungeoneer.exe" },
  @{ src = "$src\meta.json";          dst = "meta.json" },
  @{ src = "$src\controls.json";      dst = "controls.json" },
  @{ src = "$src\levels\hub.json";    dst = "levels/hub.json" }
)

Add-Type -Assembly "System.IO.Compression.FileSystem"
$zip = [System.IO.Compression.ZipFile]::Open($out, "Create")

foreach ($f in $files) {
  $entry  = $zip.CreateEntry($f.dst)
  $writer = $entry.Open()
  $bytes  = [System.IO.File]::ReadAllBytes($f.src)
  $writer.Write($bytes, 0, $bytes.Length)
  $writer.Close()
}

$zip.Dispose()
Write-Host "Created dungeoneer.zip"
