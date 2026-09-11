@echo off
setlocal
cd /d "%~dp0"
if not exist "dist\Viticulture-Continuation.exe" (
  echo Build first: powershell -File scripts\build-continuation.ps1
  pause
  exit /b 1
)
echo URL: http://localhost:3015
echo Data: runtime\continuation-data
echo Each player sets a seat password when creating or joining a room.
"dist\Viticulture-Continuation.exe" -addr 0.0.0.0:3015 -data runtime/continuation-data
pause
