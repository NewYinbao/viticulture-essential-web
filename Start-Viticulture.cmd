@echo off
chcp 65001 >nul
cd /d "%~dp0"
if not exist "%~dp0dist\Viticulture.exe" (
  echo Build first: powershell -File scripts\build.ps1
  pause
  exit /b 1
)
echo Open http://localhost:3013 . Friends use this computer's LAN IP and port 3013.
echo This version uses runtime\ee-data. Existing game saves stay in their original folders.
"%~dp0dist\Viticulture.exe" -addr 0.0.0.0:3013 -data "%~dp0runtime\ee-data"
if errorlevel 1 pause
