@echo off
setlocal enabledelayedexpansion

REM Android APK 构建脚本 (Windows版本)
REM 解决版本号和AAR缺失问题

echo =========================================
echo Android APK 构建脚本 (Windows)
echo =========================================

REM 1. 检查依赖
echo 检查构建依赖...

REM 检查Go
go version >nul 2>&1
if errorlevel 1 (
    echo 错误: 未找到Go，请先安装Go
    exit /b 1
)

REM 检查Java
java -version >nul 2>&1
if errorlevel 1 (
    echo 错误: 未找到Java，请先安装JDK 17
    exit /b 1
)

REM 检查Android SDK
if "%ANDROID_HOME%"=="" (
    echo 错误: 未设置ANDROID_HOME环境变量
    exit /b 1
)

REM 2. 设置版本号
if "%1"=="" (
    for /f "tokens=1-3 delims=/ " %%a in ('date /t') do set VERSION=%%c.%%a.%%b
) else (
    set VERSION=%1
)
echo 使用版本号: %VERSION%
set APP_VERSION=%VERSION%
set GITHUB_REF_NAME=%VERSION%

REM 3. 安装gomobile (如果需要)
echo 检查gomobile...
gomobile version >nul 2>&1
if errorlevel 1 (
    echo 安装gomobile...
    go install golang.org/x/mobile/cmd/gomobile@latest
    go install golang.org/x/mobile/cmd/gobind@latest
    gomobile init
)

REM 4. 准备Go模块
echo 准备Go模块...
if not exist go.work (
    go work init
    go work use . ./tunnel ./mobile
)

cd mobile
go mod tidy
cd ..

REM 5. 构建AAR
echo 构建AAR文件...
cd mobile
gomobile bind -v -target=android -androidapi=24 -o ../android/ech-workers.aar .
cd ..

REM 验证AAR文件
if not exist android\ech-workers.aar (
    echo 错误: AAR文件构建失败
    exit /b 1
)

echo AAR文件构建成功
dir android\ech-workers.aar

REM 6. 构建APK
echo 构建APK...
cd android
gradlew.bat clean
gradlew.bat assembleDebug -PAPP_VERSION=%VERSION%

REM 7. 检查输出
echo 构建完成！
echo APK文件位置:
dir /s build\outputs\apk\*.apk

echo =========================================
echo 构建成功完成！
echo 版本号: %VERSION%
echo =========================================