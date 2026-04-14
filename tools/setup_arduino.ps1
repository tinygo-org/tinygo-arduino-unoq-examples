<#
.SYNOPSIS
    Sets up the Arduino UNO Q for use with TinyGo this session.

.DESCRIPTION
    This script stops the arduino-router service, and installs the screen utility.

#>

# 1. Stop the arduino-router service
Write-Host "Stopping arduino-router..." -ForegroundColor Cyan
adb shell -t "sudo systemctl stop arduino-router"

# 2. Install screen utility
Write-Host "Install screen utility..." -ForegroundColor Cyan
adb shell -t "sudo apt-get update && sudo apt-get install -y screen"

Write-Host "Done!" -ForegroundColor Green
