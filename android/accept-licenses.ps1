# 自动接受Android SDK许可证
Write-Host "正在接受Android SDK许可证..." -ForegroundColor Green

# 创建必要的目录
$licensesDir = "D:\Android\Sdk\licenses"
if (!(Test-Path $licensesDir)) {
    New-Item -ItemType Directory -Path $licensesDir -Force
}

# 添加许可证文件
$licenseContent = @"
8933bad161af4178b1185d1a37fbf41ea5269c55
"@ 

$licenseContent | Out-File -FilePath "$licensesDir\android-sdk-license" -Encoding ASCII

$intelLicense = @"
d975f751698a77b662f1254ddbeed3901e976f5a
"@

$intelLicense | Out-File -FilePath "$licensesDir\intel-android-extra-license" -Encoding ASCII

Write-Host "许可证已接受!" -ForegroundColor Green
Write-Host "现在可以重新构建项目了" -ForegroundColor Yellow