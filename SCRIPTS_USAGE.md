# Upload Scripts Usage

## Available Scripts

### 1. `upload.ps1` (English)
English version upload script with no encoding issues.

**Usage:**
```powershell
.\upload.ps1
```

### 2. `upload-cn.ps1` (Chinese Pinyin)
Chinese version with pinyin prompts to avoid encoding issues.

**Usage:**
```powershell
.\upload-cn.ps1
```

### 3. `upload.sh` (Linux/macOS)
Bash script for Unix-like systems.

**Usage:**
```bash
chmod +x upload.sh
./upload.sh
```

## Quick Upload Steps

### Option 1: Use Script (Recommended)

```powershell
# Run the upload script
.\upload.ps1

# Follow the prompts:
# 1. Enter commit message (or press Enter for default)
# 2. Enter GitHub repository URL (if not configured)
```

### Option 2: Manual Upload

```bash
# Initialize git
git init
git branch -M main

# Add and commit files
git add .
git commit -m "Initial commit"

# Add remote repository (replace with your URL)
git remote add origin https://github.com/YOUR_USERNAME/ech-workers.git

# Push to GitHub
git push -u origin main
```

## Release New Version

```bash
# Create and push tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# GitHub Actions will automatically:
# - Build Android APK (all architectures)
# - Build OpenWrt IPK (x86_64, aarch64, mipsel)
# - Create GitHub Release
# - Upload all build artifacts
```

## Troubleshooting

### Script Encoding Issues

If you see garbled text, use the English version:
```powershell
.\upload.ps1
```

Or run commands manually:
```powershell
git add .
git commit -m "Update"
git push origin main
```

### Permission Denied

```powershell
# Enable script execution (run as Administrator)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Remote Already Exists

```bash
# Update remote URL
git remote set-url origin https://github.com/YOUR_USERNAME/ech-workers.git

# Or remove and re-add
git remote remove origin
git remote add origin https://github.com/YOUR_USERNAME/ech-workers.git
```

## Next Steps

After uploading:

1. **Update README**: Replace `YOUR_USERNAME` with your GitHub username
2. **Configure Secrets**: Add keystore secrets for APK signing (optional)
3. **Create Release**: Push a tag to trigger automatic builds
4. **Star Project**: Give your project a star!

For more details, see [UPLOAD_GUIDE.md](UPLOAD_GUIDE.md)
