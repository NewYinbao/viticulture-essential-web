@echo off

chcp 65001 >nul

cd /d "%~dp0"

echo Viticulture EE All-Card RC1 - 2 to 6 human players.

echo Open http://localhost:3012 after server starts.

echo LAN: use this computer LAN IP and port 3012.

echo Close this window or press Ctrl+C to stop.

"%~dp0artifacts\legacy-builds\Viticulture-EE-AllCards-RC1.exe" -addr 0.0.0.0:3012 -data "%~dp0data-ee-allcards"

if errorlevel 1 pause
