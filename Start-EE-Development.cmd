@echo off
chcp 65001 >nul
cd /d "%~dp0"
echo EE DEVELOPMENT ONLY - 76 visitor effects NOT implemented.
echo Open http://localhost:3011 - separate from the previous preview.
"%~dp0Viticulture-EE-Development.exe" -addr 0.0.0.0:3011 -data "%~dp0data-ee"
if errorlevel 1 pause
