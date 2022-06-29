#!/bin/sh

mkdir -p bin

# _build OS ARCH EXT
_build() {
	os=$1
	arch=$2
	if [ -n "$3" ]; then
		ext=".$3"
	else
		ext=""
	fi

	echo "Building: ${os}_${arch}"
	env GOOS="$1" GOARCH="$2" go build -o "bin/instx-${os}-${arch}${ext}"
}

_build "linux" "amd64"
_build "linux" "386"
_build "windows" "amd64" "exe"
_build "windows" "386" "exe"
