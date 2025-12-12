# 快速修复Android构建问题

## 立即解决方案

### 1. 快速修复版本号问题

已经修改了 `android/build.gradle`，现在会使用当前日期作为默认版本号（如 2024.12.12），你也可以手动指定：

```bash
# Windows
cd android
gradlew.bat assembleDebug -PAPP_VERSION=1.2.3

# Linux/Mac
cd android
./gradlew assembleDebug -PAPP_VERSION=1.2.3
```

### 2. 快速修复AAR缺失问题

**方案A: 使用构建脚本（推荐）**

我已经创建了构建脚本：
- Windows: `build-android.bat`
- Linux/Mac: `build-android.sh`

使用方法：
```bash
# Windows
build-android.bat 1.2.3

# Linux/Mac
chmod +x build-android.sh
./build-android.sh 1.2.3
```

**方案B: 手动构建AAR**

如果你已经安装了Go和gomobile：

```bash
# 1. 安装gomobile（如果没有）
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
gomobile init

# 2. 准备Go模块
go work init
go work use . ./tunnel ./mobile

# 3. 构建AAR
cd mobile
go mod tidy
gomobile bind -target=android -androidapi=24 -o ../android/ech-workers.aar .
cd ..

# 4. 构建APK
cd android
./gradlew assembleDebug -PAPP_VERSION=1.2.3
```

**方案C: 临时跳过AAR依赖（仅用于测试UI）**

如果你只想测试UI是否正常，可以临时注释掉AAR依赖：

在 `android/build.gradle` 中找到：
```gradle
dependencies {
    implementation 'androidx.appcompat:appcompat:1.4.1'
    implementation 'androidx.cardview:cardview:1.0.0'
    implementation files('ech-workers.aar')  // 临时注释这行
}
```

改为：
```gradle
dependencies {
    implementation 'androidx.appcompat:appcompat:1.4.1'
    implementation 'androidx.cardview:cardview:1.0.0'
    // implementation files('ech-workers.aar')  // 临时注释
}
```

然后在 `MainActivity.java` 中注释掉相关检查：
```java
// 临时注释掉这段检查
/*
try {
    Log.d(TAG, "onCreate: Testing mobile.Mobile class availability");
    Class.forName("mobile.Mobile");
    Log.d(TAG, "onCreate: mobile.Mobile class found");
} catch (ClassNotFoundException e) {
    Log.e(TAG, "onCreate: mobile.Mobile class NOT found - AAR missing or not loaded!", e);
    Toast.makeText(this, "错误: 缺少核心库文件,请重新安装应用", Toast.LENGTH_LONG).show();
}
*/
```

这样可以先构建出APK测试UI，但功能不会正常工作。

## 推荐步骤

1. 使用我提供的构建脚本 `build-android.bat` 或 `build-android.sh`
2. 如果遇到Go或gomobile问题，先安装这些依赖
3. 构建成功后，APK应该可以正常安装和显示UI

## 常见问题

**Q: gomobile 安装失败**
A: 确保Go版本 >= 1.19，网络连接正常，可能需要设置GOPROXY

**Q: AAR构建失败**
A: 检查mobile目录下的go.mod文件是否正确，确保所有依赖都能下载

**Q: 仍然白屏**
A: 检查logcat输出，看具体错误信息