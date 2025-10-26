#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status.

echo "Starting Snap build process..."

# Define source directories (relative to project root)
GORELEASER_DIST_DIR=$(find ./dist -maxdepth 1 -type d -name 'lettercli_linux_amd64*' | head -n 1)
if [ -z "$GORELEASER_DIST_DIR" ]; then
    echo "Error: Could not find GoReleaser Linux build directory in ./dist"
    echo "Please run 'goreleaser build --snapshot --clean' first."
    exit 1
fi
echo "Found GoReleaser build dir: $GORELEASER_DIST_DIR"

PYTHON_EXECS_DIR="./dist_py/linux_amd64"      # Where your pre-built Python execs are
SNAP_CONFIG_DIR="./snap"                      # Where snapcraft.yaml is

# Define the Snapcraft build directory (temporary)
SNAP_BUILD_DIR="./dist/snap_build"

# Clean previous build attempt
echo "Cleaning previous build..."
rm -rf "$SNAP_BUILD_DIR"
mkdir -p "$SNAP_BUILD_DIR/bin"    # Create bin dir
mkdir -p "$SNAP_BUILD_DIR/py_execs" # Create py_execs dir

# Stage necessary files for Snapcraft build context
echo "Staging files..."
# Copy Go binary INTO bin/
echo "Copying Go binary from $GORELEASER_DIST_DIR to $SNAP_BUILD_DIR/bin/ ..."
if [ ! -f "$GORELEASER_DIST_DIR/lettercli" ]; then
    echo "Error: lettercli binary not found in $GORELEASER_DIST_DIR"
    exit 1
fi
cp "$GORELEASER_DIST_DIR"/lettercli "$SNAP_BUILD_DIR/bin/"

# Copy Python executables INTO py_execs/
echo "Copying Python executables from $PYTHON_EXECS_DIR to $SNAP_BUILD_DIR/py_execs/ ..."
if [ ! -d "$PYTHON_EXECS_DIR" ] || [ -z "$(ls -A $PYTHON_EXECS_DIR)" ]; then
    echo "Error: Python executables directory ($PYTHON_EXECS_DIR) not found or empty."
    exit 1
fi
# Copy the *contents* of the python execs dir
cp "$PYTHON_EXECS_DIR"/* "$SNAP_BUILD_DIR/py_execs/"
chmod +x "$SNAP_BUILD_DIR/py_execs"/* # Ensure they are executable after copy

# Copy snapcraft.yaml to the root of the build dir
echo "Copying snapcraft.yaml from $SNAP_CONFIG_DIR..."
if [ ! -f "$SNAP_CONFIG_DIR/snapcraft.yaml" ]; then
    echo "Error: snapcraft.yaml not found in $SNAP_CONFIG_DIR"
    exit 1
fi
cp "$SNAP_CONFIG_DIR/snapcraft.yaml" "$SNAP_BUILD_DIR/"

# Run Snapcraft within the staged directory
echo "Running Snapcraft..."
cd "$SNAP_BUILD_DIR"
snapcraft pack --destructive-mode # Keep destructive mode for now

# Optional: Move the final .snap file back to the main dist directory
echo "Moving snap file..."
SNAP_FILE=$(find . -maxdepth 1 -name '*.snap' | head -n 1)
if [ -n "$SNAP_FILE" ]; then
    mv "$SNAP_FILE" ../
    echo "Snap file moved to dist/"
else
    echo "Warning: Could not find generated .snap file to move."
fi

echo "Snap build complete."
cd ../.. # Go back to project root