# Android 构建问题修复指南

## 问题1: 版本号缺失

### 解决方案1: 简化版本号获取逻辑

修改 `android/build.gradle` 中的 `getVersionName()` 函数，添加更可靠的版本号获取方式：

```gradle
def getVersionName() {
    // 如果已经缓存,直接返回
    if (ext.cachedVersionName != null) {
        return ext.cachedVersionName
    }
    
    // 1. 优先使用 gradle property
    if (project.hasProperty('APP_VERSION')) {
        def prop = project.property('APP_VERSION')?.toString()?.trim()
        if (prop && prop != '' && prop.toLowerCase() != 'null') {
            def version = prop.replaceAll('^v', '')
            if (version.matches('[0-9]+\\.[0-9]+\\.[0-9]+')) {
                ext.cachedVersionName = version
                return version
            }
        }
    }
    
    // 2. 从环境变量获取
    def envVersion = System.getenv('GITHUB_REF_NAME') ?: System.getenv('APP_VERSION')
    if (envVersion) {
        def version = envVersion.toString().replaceAll('^(refs/tags/)?v', '').trim()
        if (version.matches('[0-9]+\\.[0-9]+\\.[0-9]+')) {
            ext.cachedVersionName = version
            return version
        }
    }
    
    // 3. 从 git tag 获取
    try {
        def process = ['git', 'describe', '--tags', '--abbrev=0'].execute()
        process.waitFor()
        if (process.exitValue() == 0) {
            def gitVersion = process.text.trim()
            if (gitVersion.startsWith('v')) {
                def version = gitVersion.substring(1)
                if (version.matches('[0-9]+\\.[0-9]+\\.[0-9]+')) {
                    ext.cachedVersionName = version
                    return version
                }
            }
        }
    } catch (Exception e) {
        // 忽略git错误
    }
    
    // 4. 使用当前日期作为版本号
    def dateVersion = new Date().format('yyyy.MM.dd')
    ext.cachedVersionName = dateVersion
    return dateVersion
}
```

### 解决方案2: 手动设置版本号

在构建时手动指定版本号：

```bash
# 方法1: 使用gradle属性
./gradlew assembleDebug -PAPP_VERSION=1.2.3

# 方法2: 设置环境变量
export APP_VERSION=1.2.3
./gradlew assembleDebug
```

## 问题2: APK白屏问题

### 原因分析
缺少 `ech-workers.aar` 文件，这个文件包含了Go编译的native库。

### 解决方案1: 构建AAR文件

1. **安装依赖**:
```bash
# 安装Go (如果没有)
# 安装gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
gomobile init
```

2. **构建AAR**:
```bash
cd mobile
gomobile bind -target=android -androidapi=24 -o ../android/ech-workers.aar .
```

### 解决方案2: 修改构建脚本

创建自动构建脚本：