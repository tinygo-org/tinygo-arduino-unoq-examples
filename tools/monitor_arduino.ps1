# Opens a serial console to the Arduino Q board via adb.
# Uses screen on the remote device at 115200 baud.

Write-Host "Connecting to /dev/ttyHS1 at 115200 baud..."

$screenrc = 'hardstatus alwayslastline "Serial: /dev/ttyHS1 @ 115200 | Exit: Ctrl+A then K"'
& adb shell -t "echo '$screenrc' > /tmp/.screenrc_serial && screen -c /tmp/.screenrc_serial /dev/ttyHS1 115200"
