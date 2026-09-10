@echo off
chcp 65001 >nul
cd /d "%~dp0"
echo Viticulture EE Rules Fix - NOT a complete rules signoff.
echo Open http://localhost:3012 after startup.
echo Do not run simultaneously with the old server on the same save directory.
echo Old active games are not repaired or migrated. Read RULE_FIX_REPORT.md.
"%~dp0Viticulture-EE-Rules-Fix.exe" -addr 0.0.0.0:3012 -data "%~dp0data-ee-allcards"
if errorlevel 1 pause
