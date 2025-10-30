#!/bin/bash
set -e

echo "Starting Snap build process..."

GORELEASER_DIST_DIR=$(find ./dist -maxdepth 1 -type d -name 'lettercli_linux_amd64*' | head -n 1)
if [ -z "$GORELEASER_DIST_DIR" ]; then
    echo "❌ Error: Could not find GoReleaser Linux build directory in ./dist"
    echo "Please run 'goreleaser release --clean' first."
    exit 1
fi
echo "✅ Found GoReleaser build dir: $GORELEASER_DIST_DIR"

PYTHON_EXECS_DIR="./dist_py/linux_amd64"
SNAP_CONFIG_DIR="./snap"
ASSETS_DIR="./assets"
SNAP_BUILD_DIR="./dist/snap_build"

echo "🧹 Cleaning previous build..."
rm -rf "$SNAP_BUILD_DIR"
mkdir -p "$SNAP_BUILD_DIR/bin" "$SNAP_BUILD_DIR/py_execs" "$SNAP_BUILD_DIR/assets"

echo "📦 Staging files..."

if [ ! -f "$GORELEASER_DIST_DIR/lettercli" ]; then
    echo "❌ Error: lettercli binary not found in $GORELEASER_DIST_DIR"
    exit 1
fi
cp "$GORELEASER_DIST_DIR/lettercli" "$SNAP_BUILD_DIR/bin/"
echo "✅ Copied Go binary."

if [ ! -d "$PYTHON_EXECS_DIR" ] || [ -z "$(ls -A "$PYTHON_EXECS_DIR")" ]; then
    echo "❌ Error: Python executables directory ($PYTHON_EXECS_DIR) not found or empty."
    exit 1
fi
cp "$PYTHON_EXECS_DIR"/* "$SNAP_BUILD_DIR/py_execs/"
chmod +x "$SNAP_BUILD_DIR/py_execs"/*
echo "✅ Copied Python executables."

if [ ! -f "$SNAP_CONFIG_DIR/snapcraft.yaml" ]; then
    echo "❌ Error: snapcraft.yaml not found in $SNAP_CONFIG_DIR"
    exit 1
fi
cp "$SNAP_CONFIG_DIR/snapcraft.yaml" "$SNAP_BUILD_DIR/"
echo "✅ Copied snapcraft.yaml."

if [ ! -d "$ASSETS_DIR" ] || [ -z "$(ls -A "$ASSETS_DIR")" ]; then
    echo "⚠️ Warning: Assets directory ($ASSETS_DIR) not found or empty. Continuing anyway."
else
    cp -r "$ASSETS_DIR"/* "$SNAP_BUILD_DIR/assets/" 
    echo "✅ Copied assets."
fi

echo "🚀 Running Snapcraft..."
cd "$SNAP_BUILD_DIR"

snapcraft pack

echo "📁 Moving snap file..."
SNAP_FILE=$(find . -maxdepth 1 -name '*.snap' | head -n 1)
if [ -n "$SNAP_FILE" ]; then
    mv "$SNAP_FILE" ../
    echo "✅ Snap file moved to dist/."
else
    echo "⚠️ Warning: Could not find generated .snap file to move."
fi

echo "✅ Snap build complete."
cd ../..

