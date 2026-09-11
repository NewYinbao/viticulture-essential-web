@echo off
setlocal
cd /d "%~dp0"
if not exist "dist\Viticulture-Expansions.exe" (
  echo Expansion development build is missing. See docs\expansions-release.md.
  pause
  exit /b 1
)
echo Legacy expansion development build. This EXE is not upgraded automatically and may not include the latest modules.
echo For the current continuation build, use Start-Viticulture-Continuation.cmd after running scripts\build-continuation.ps1.
echo URL: http://localhost:3014
echo Data: runtime\expansions-data
"dist\Viticulture-Expansions.exe" -addr 0.0.0.0:3014 -data runtime/expansions-data
pause
