# Windows 控制台 UTF-8 输出，避免 dev-all 等脚本 Write-Host 中文乱码。
# 用法：在脚本开头 . "$PSScriptRoot\ps-console-utf8.ps1"
if ($env:OS -ne 'Windows_NT') { return }

try {
    $utf8 = [System.Text.UTF8Encoding]::new($false)
    [Console]::OutputEncoding = $utf8
    $OutputEncoding = $utf8
    if ($Host.Name -eq 'ConsoleHost') {
        $null = cmd /c chcp 65001 >$null 2>&1
    }
} catch {
    # 非交互环境忽略
}
