#!/bin/bash

set -e

# --- Pre-flight checks ---
if ! command -v flutter &> /dev/null
then
    echo "flutter could not be found. Please ensure the Flutter SDK is installed and in your PATH."
    exit 1
fi

echo "========================================================================"
echo "Reddit Client Build Script"
echo "========================================================================"
echo "NOTE: This script is configured to build for Linux by default."
echo "The current development environment has limitations that prevent"
echo "generating build files for other platforms."
echo "To build for other platforms, first add support for them by running:"
echo "flutter create . --platforms=<platform-name>"
echo "Then, uncomment the corresponding build section in this script."
echo "========================================================================"

# --- Clean and Setup ---
echo "Cleaning up previous builds..."
rm -rf build/
mkdir -p build

# --- Build for Linux ---
echo "Building for Linux..."
flutter build linux
echo "Packaging Linux build..."
mkdir -p build/linux
mv build/linux/x64/release/bundle/* build/linux/
echo "Linux build complete. Artifacts are in build/linux/"

# ========================================================================
# --- Build commands for other platforms (uncomment to use) ---
# ========================================================================

# --- Android ---
# echo "Building for Android (APK)..."
# flutter build apk
# echo "Packaging Android build..."
# mkdir -p build/android
# mv build/app/outputs/flutter-apk/app-release.apk build/android/
# echo "Android build complete. Artifacts are in build/android/"

# --- iOS ---
# echo "Building for iOS..."
# flutter build ios --no-codesign
# echo "Packaging iOS build..."
# mkdir -p build/ios
# mv build/ios/iphoneos/Runner.app build/ios/
# echo "iOS build complete. Artifacts are in build/ios/"

# --- macOS ---
# echo "Building for macOS..."
# flutter build macos
# echo "Packaging macOS build..."
# mkdir -p build/macos
# mv "build/macos/Build/Products/Release/reddit_client.app" build/macos/
# echo "macOS build complete. Artifacts are in build/macos/"

# --- Windows ---
# echo "Building for Windows..."
# flutter build windows
# echo "Packaging Windows build..."
# mkdir -p build/windows
# mv build/windows/runner/Release/* build/windows/
# echo "Windows build complete. Artifacts are in build/windows/"

echo "========================================================================"
echo "Build script finished successfully."
echo "========================================================================"
