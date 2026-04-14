param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$SourcePath,

    [Parameter(Mandatory=$true, Position=1)]
    [string]$TargetHost,

    [Parameter(Position=2)]
    [string]$TargetUser = "arduino"
)

$AllowSSHAuthWithPassword = @("-o", "PreferredAuthentications=password", "-o", "PasswordAuthentication=yes")
$Target = "$TargetUser@$TargetHost"

# 1. Use tinygo to build the firmware and generate a .hex file
$HexOutputPath = [System.IO.Path]::ChangeExtension($SourcePath, ".hex")
Write-Host "Building firmware from source '$SourcePath'..."
& tinygo build -o $HexOutputPath -target=arduino-uno-q $SourcePath
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to build firmware from $SourcePath"
    exit 1
}

$HexFilename = Split-Path $HexOutputPath -Leaf
$TargetHexPath = "/home/arduino/$HexFilename"

# 2. Upload the .hex file
Write-Host "Uploading firmware '$HexFilename'..."
& scp @AllowSSHAuthWithPassword $HexOutputPath "${Target}:${TargetHexPath}"
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to upload $HexOutputPath"
    exit 1
}

# 3. Run the OpenOCD command on the target to program the firmware
Write-Host "Flashing firmware..."
& ssh $Target @AllowSSHAuthWithPassword "/opt/openocd/bin/openocd -s /opt/openocd/share/openocd/scripts -s /opt/openocd -c `"adapter driver linuxgpiod`" -c `"adapter gpio swclk -chip 1 26`" -c `"adapter gpio swdio -chip 1 25`" -c `"adapter gpio srst -chip 1 38`" -c `"transport select swd`" -c `"adapter speed 1000`" -c `"reset_config srst_only srst_push_pull`" -f /opt/openocd/stm32u5x.cfg -c `"program $TargetHexPath verify reset exit`""

Write-Host "Done!"
