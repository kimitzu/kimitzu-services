@echo off
For /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c-%%a-%%b)
For /f "tokens=1-2 delims=/: " %%a in ('time /t') do (set mytime=%%a%%b)
echo Compiling to release\services.%mydate%_%mytime%.go
mkdir release
echo Building
packr2 build -ldflags="-s -w" -o release\services.%mydate%_%mytime%.exe
echo Cleaning packr
packr2 clean