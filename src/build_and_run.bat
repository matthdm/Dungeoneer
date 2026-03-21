:: To use, run this command from project's root directory: .\build_and_run.bat
@echo off

:: Build the project and run .exe file without an additional terminal window
go build -ldflags -H=windowsgui -o dungeoneer.exe

if %errorlevel% equ 0 (
    dungeoneer.exe
) else (
    echo Build failed!
    pause
)