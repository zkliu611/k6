#!/bin/bash
set -e

rm -rf pkg-build

echo "Building k6..."
GOARCH=$ARCH go build -o /tmp/k6
echo "Done!"

go-bin-$1 generate --file packaging/$1.json -a $ARCH --version $VERSION -o dist/k6-v$VERSION-$ARCH.deb