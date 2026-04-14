#!/bin/bash

# 1. Stop the arduino-router service
echo "Stopping arduino-router..."
adb shell -t "sudo systemctl stop arduino-router"

# 2. Install screen utility
echo "Installing screen..."
adb shell -t "sudo apt-get update && sudo apt-get install -y screen"

echo "Done!"
