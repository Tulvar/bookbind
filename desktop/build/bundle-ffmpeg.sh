#!/usr/bin/env bash
set -euo pipefail

version=9.1
sha256=1a4d394ae63d190a526cd052eb4415c156b5c5e2df37efab41a7de0b8fde2433
archive="ffmpeg-${version}.tar.xz"
destination=${1:?Usage: bash desktop/build/bundle-ffmpeg.sh DESTINATION}
mkdir -p "$destination"
destination=$(cd "$destination" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cd "$work"

curl --fail --silent --show-error --location --retry 3 --proto '=https' \
    --output "$archive" "https://ffmpeg.org/releases/$archive"
if command -v shasum >/dev/null; then
    printf '%s  %s\n' "$sha256" "$archive" | shasum -a 256 -c -
else
    printf '%s  %s\n' "$sha256" "$archive" | sha256sum -c -
fi
tar -xf "$archive"
cd "ffmpeg-$version"

config=(
    --disable-autodetect
    --disable-gpl --disable-nonfree --disable-version3
    --disable-shared --enable-static
    --disable-ffplay --disable-doc --disable-debug
)
extension=
case "$(uname -s)" in
    Darwin)
        export MACOSX_DEPLOYMENT_TARGET=11.0
        ;;
    MSYS*|MINGW*|UCRT*)
        config+=(--target-os=mingw32 --extra-ldflags=-static)
        extension=.exe
        ;;
esac
./configure "${config[@]}"
make -j "${FFMPEG_BUILD_JOBS:-$(getconf _NPROCESSORS_ONLN)}" ffmpeg"$extension" ffprobe"$extension"

mkdir -p "$destination/bin" "$destination/ffmpeg-source"
install -m 755 "ffmpeg$extension" "ffprobe$extension" "$destination/bin/"
cp "../$archive" COPYING.LGPLv2.1 LICENSE.md "$destination/ffmpeg-source/"
printf './configure' > "$destination/ffmpeg-source/configure.txt"
printf ' %s' "${config[@]}" >> "$destination/ffmpeg-source/configure.txt"
printf '\n' >> "$destination/ffmpeg-source/configure.txt"
if [[ $(uname -s) == Darwin ]]; then
    printf 'MACOSX_DEPLOYMENT_TARGET=%s\n' "$MACOSX_DEPLOYMENT_TARGET" >> "$destination/ffmpeg-source/configure.txt"
fi

# Check the installed pair without a system FFmpeg or toolchain DLLs on PATH.
ffmpeg="$destination/bin/ffmpeg$extension"
ffprobe="$destination/bin/ffprobe$extension"
ffmpeg_version=$(PATH= "$ffmpeg" -version)
ffprobe_version=$(PATH= "$ffprobe" -version)
[[ $ffmpeg_version == "ffmpeg version $version "* ]]
[[ $ffprobe_version == "ffprobe version $version "* ]]
PATH= "$ffmpeg" -v error -f lavfi -i sine=duration=0.2 -c:a aac -movflags +faststart "$work/test.m4b"
codec=$(PATH= "$ffprobe" -v error -select_streams a:0 -show_entries stream=codec_name -of csv=p=0 "$work/test.m4b")
[[ $codec == aac ]]
