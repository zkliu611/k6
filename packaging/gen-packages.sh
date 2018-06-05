#!/bin/bash
set -e

if [[ -z "${VERSION}" ]]; then
    echo "\$VERSION enviroment variable undefined"
    exit 2
fi

if [[ -z "${ARCH}" ]]; then
    echo "\$ARCH enviroment variable undefined"
    exit 2
fi

rm -rf pkg-build

echo "Building k6..."
CGO_ENABLED=0 GOARCH=$ARCH go build -a -ldflags '-s -w' -o /tmp/k6
echo "Done!"

mkdir -p dist
go-bin-deb generate --file packaging/deb.json -a $ARCH --version $VERSION -o dist/k6-v$VERSION-$ARCH.deb
go-bin-rpm generate --file packaging/rpm.json -a $ARCH --version $VERSION -o dist/k6-v$VERSION-$ARCH.rpm