#!/bin/sh
set -eu

: "${PACKAGE_FORMAT:?PACKAGE_FORMAT is required}"
: "${PACKAGE_FILE:?PACKAGE_FILE is required}"
: "${SOURCE_COMMIT:?SOURCE_COMMIT is required}"
: "${EVIDENCE:=/workspace/artifacts/validation/linux-package-lifecycle}"
: "${PACKAGE_INPUTS:=/workspace/artifacts/package-fixture}"
: "${RUNNER_TEMP:=/runner-temp}"

package="$EVIDENCE/$PACKAGE_FILE"
home="$RUNNER_TEMP/smolpng-package-home"
toolroot="$RUNNER_TEMP/smolpng-optimizers"
package_installed=false

record_host() {
	cat /etc/os-release > "$EVIDENCE/container-os-release.txt"
	uname -srm > "$EVIDENCE/container-kernel.txt"
	test "$(uname -m)" = x86_64
	printf 'FORMAT=%s\n' "$PACKAGE_FORMAT" > "$EVIDENCE/package-format.txt"
	case "$PACKAGE_FORMAT" in
		rpm) grep -q '^ID=fedora$' /etc/os-release ;;
		apk) grep -q '^ID=alpine$' /etc/os-release ;;
		archlinux) grep -q '^ID=arch$' /etc/os-release ;;
		*) return 2 ;;
	esac
}

remove_package() {
	case "$PACKAGE_FORMAT" in
		rpm)
			dnf remove --assumeyes smolpng > "$EVIDENCE/remove.txt"
			if rpm -q smolpng > "$EVIDENCE/removed-package-status.txt" 2>&1; then
				return 1
			fi
			;;
		apk)
			apk del smolpng > "$EVIDENCE/remove.txt"
			if apk info -e smolpng > "$EVIDENCE/removed-package-status.txt" 2>&1; then
				return 1
			fi
			;;
		archlinux)
			pacman -R --noconfirm smolpng > "$EVIDENCE/remove.txt"
			if pacman -Q smolpng > "$EVIDENCE/removed-package-status.txt" 2>&1; then
				return 1
			fi
			;;
		*)
			printf 'Unsupported package format: %s\n' "$PACKAGE_FORMAT" >&2
			return 1
			;;
	esac
	test ! -e /usr/bin/smolpng
}

cleanup() {
	result=$?
	trap - EXIT
	if [ "$package_installed" = true ]; then
		set +e
		if [ ! -e "$EVIDENCE/user-state-before.sha256" ] && [ -d "$home" ]; then
			find "$home" -type f -print0 | sort -z | xargs -0 -r sha256sum > "$EVIDENCE/user-state-before.sha256"
		fi
		remove_package
		cleanup_result=$?
		if [ "$cleanup_result" -eq 0 ] && [ -s "$EVIDENCE/user-state-before.sha256" ]; then
			sha256sum --check "$EVIDENCE/user-state-before.sha256" > "$EVIDENCE/user-state-preserved.txt"
			cleanup_result=$?
		fi
		set -e
		if [ "$cleanup_result" -ne 0 ]; then
			printf 'Package removal did not complete; inspect remove.txt and package state.\n' >&2
			result=1
		fi
	fi
	exit "$result"
}
trap cleanup EXIT

record_host
test -f "$package"
sha256sum "$package" > "$EVIDENCE/container-package.sha256"
mkdir -p "$home" "$toolroot"

case "$PACKAGE_FORMAT" in
	rpm)
		dnf install --assumeyes ca-certificates coreutils curl diffutils findutils gzip jq pngquant tar > "$EVIDENCE/container-tools-install.txt"
		dnf --version > "$EVIDENCE/package-manager-version.txt"
		dnf list --installed pngquant > "$EVIDENCE/pngquant-package.txt"
		rpm -qp --queryformat '%{NAME} %{VERSION} %{ARCH}\n' "$package" > "$EVIDENCE/package-fields.txt"
		test "$(rpm -qp --queryformat '%{NAME}' "$package")" = smolpng
		test "$(rpm -qp --queryformat '%{ARCH}' "$package")" = x86_64
		rpm -qpl "$package" > "$EVIDENCE/package-contents.txt"
		package_installed=true
		dnf install --assumeyes --nogpgcheck "$package" > "$EVIDENCE/install.txt"
		;;
	apk)
		apk add --no-cache ca-certificates coreutils curl diffutils findutils gzip jq pngquant build-base libjpeg-turbo-dev tar > "$EVIDENCE/container-tools-install.txt"
		apk --version > "$EVIDENCE/package-manager-version.txt"
		apk info -v pngquant libjpeg-turbo-dev > "$EVIDENCE/pngquant-package.txt"
		cc --version > "$EVIDENCE/c-compiler-version.txt"
		make --version > "$EVIDENCE/make-version.txt"
		tar --extract --to-stdout --gzip --file "$package" .PKGINFO > "$EVIDENCE/package-fields.txt"
		grep -qx 'pkgname = smolpng' "$EVIDENCE/package-fields.txt"
		grep -qx 'arch = x86_64' "$EVIDENCE/package-fields.txt"
		tar --list --gzip --file "$package" > "$EVIDENCE/package-contents.txt"
		package_installed=true
		apk add --allow-untrusted "$package" > "$EVIDENCE/install.txt"
		;;
	archlinux)
		pacman -Syu --noconfirm --needed ca-certificates coreutils curl diffutils findutils gzip jq pngquant base-devel libarchive libjpeg-turbo tar > "$EVIDENCE/container-tools-install.txt"
		pacman --version > "$EVIDENCE/package-manager-version.txt"
		pacman -Q pngquant > "$EVIDENCE/pngquant-package.txt"
		bsdtar -xOf "$package" .PKGINFO > "$EVIDENCE/package-fields.txt"
		grep -qx 'pkgname = smolpng' "$EVIDENCE/package-fields.txt"
		grep -qx 'arch = x86_64' "$EVIDENCE/package-fields.txt"
		pacman -Qip "$package" > "$EVIDENCE/package-contents.txt"
		sed '/^SigLevel =/a LocalFileSigLevel = Optional' /etc/pacman.conf > "$RUNNER_TEMP/pacman.conf"
		package_installed=true
		pacman --config "$RUNNER_TEMP/pacman.conf" -U --noconfirm "$package" > "$EVIDENCE/install.txt"
		;;
	*)
		printf 'Unsupported package format: %s\n' "$PACKAGE_FORMAT" >&2
		exit 2
		;;
esac

pngquant --version > "$EVIDENCE/pngquant-version.txt"
sha256sum "$(command -v pngquant)" > "$EVIDENCE/pngquant.sha256"

if [ "$PACKAGE_FORMAT" = apk ]; then
	# The maintained Linux amd64 jpegoptim archive targets glibc; build the same
	# pinned upstream source natively for Alpine's musl userland instead.
	source_archive="$toolroot/jpegoptim-1.5.6.tar.gz"
	curl --fail --location --silent --show-error \
		https://github.com/tjko/jpegoptim/releases/download/v1.5.6/jpegoptim-1.5.6.tar.gz \
		--output "$source_archive"
	printf '%s  %s\n' \
		661a808dfffa933d78c6beb47a2937d572b9f03e94cbaaab3d4c0d72f410e9be \
		"$source_archive" | sha256sum --check > "$EVIDENCE/jpegoptim-source.sha256"
	mkdir "$toolroot/jpegoptim-source"
	tar --extract --gzip --file "$source_archive" --directory "$toolroot/jpegoptim-source"
	source_dir="$toolroot/jpegoptim-source/jpegoptim-1.5.6"
	(cd "$source_dir" && ./configure --prefix="$toolroot/jpegoptim-install" > "$EVIDENCE/jpegoptim-configure.txt")
	make --directory "$source_dir" --jobs=2 > "$EVIDENCE/jpegoptim-build.txt"
	make --directory "$source_dir" install > "$EVIDENCE/jpegoptim-install.txt"
	jpegoptim="$toolroot/jpegoptim-install/bin/jpegoptim"
	"$jpegoptim" --version > "$EVIDENCE/jpegoptim-version.txt"
	grep -q '1\.5\.6' "$EVIDENCE/jpegoptim-version.txt"
	sha256sum "$jpegoptim" > "$EVIDENCE/jpegoptim.sha256"
else
	"/usr/bin/smolpng" --home "$home" --no-project deps install jpegoptim --apply --json > "$EVIDENCE/jpegoptim-install.json"
	jpegoptim=$(jq -er '.managed | select(.state == "current") | .path' "$EVIDENCE/jpegoptim-install.json")
	"$jpegoptim" --version > "$EVIDENCE/jpegoptim-version.txt"
	sha256sum "$jpegoptim" > "$EVIDENCE/jpegoptim.sha256"
fi

"/usr/bin/smolpng" --home "$home" --no-project deps install oxipng --apply --json > "$EVIDENCE/oxipng-install.json"
oxipng=$(jq -er '.managed | select(.state == "current") | .path' "$EVIDENCE/oxipng-install.json")
"$oxipng" --version > "$EVIDENCE/oxipng-version.txt"
sha256sum "$oxipng" > "$EVIDENCE/oxipng.sha256"

"/usr/bin/smolpng" --home "$home" --no-project version --json > "$EVIDENCE/installed-version.json"
jq -e --arg source "$SOURCE_COMMIT" '.commit == $source and .version == "0.0.0"' "$EVIDENCE/installed-version.json"
"/usr/bin/smolpng" --home "$home" --no-project update status --json > "$EVIDENCE/update-status.json"
jq -e '.channel.owner == "system-package" and (.channel.canApply | not)' "$EVIDENCE/update-status.json"

case "$PACKAGE_FORMAT" in
	rpm)
		rpm -qf --queryformat '%{NAME} %{VERSION} %{ARCH}\n' /usr/bin/smolpng > "$EVIDENCE/executable-owner.txt"
		test "$(rpm -qf --queryformat '%{NAME}' /usr/bin/smolpng)" = smolpng
		rpm -V smolpng > "$EVIDENCE/package-integrity.txt"
		;;
	apk)
		apk info -W /usr/bin/smolpng > "$EVIDENCE/executable-owner.txt"
		grep -q 'smolpng' "$EVIDENCE/executable-owner.txt"
		;;
	archlinux)
		pacman -Qo /usr/bin/smolpng > "$EVIDENCE/executable-owner.txt"
		pacman --config "$RUNNER_TEMP/pacman.conf" -Qkk smolpng > "$EVIDENCE/package-integrity.txt"
		;;
esac
cmp "$PACKAGE_INPUTS/stage/linux-amd64/smolpng" /usr/bin/smolpng
cmp "$PACKAGE_INPUTS/stage/linux-amd64/LICENSE" /usr/share/licenses/smolpng/LICENSE
diff -r "$PACKAGE_INPUTS/stage/linux-amd64/LICENSES" /usr/share/licenses/smolpng/dependencies
cmp "$PACKAGE_INPUTS/stage/linux-amd64/third-party-notices.md" /usr/share/doc/smolpng/third-party-notices.md
sha256sum /usr/bin/smolpng > "$EVIDENCE/installed-executable.sha256"

"$PACKAGE_INPUTS/stage/linux-amd64/smolpng-qualify" \
	-executable /usr/bin/smolpng -expected-commit "$SOURCE_COMMIT" \
	-pngquant "$(command -v pngquant)" -oxipng "$oxipng" -jpegoptim "$jpegoptim" \
	-output "$EVIDENCE/journey"
jq -e '.complete and .passed' "$EVIDENCE/journey/report.json"

find "$home" -type f -print0 | sort -z | xargs -0 -r sha256sum > "$EVIDENCE/user-state-before.sha256"
remove_package
package_installed=false
if [ -s "$EVIDENCE/user-state-before.sha256" ]; then
	sha256sum --check "$EVIDENCE/user-state-before.sha256" > "$EVIDENCE/user-state-preserved.txt"
else
	printf 'No per-user files existed before removal\n' > "$EVIDENCE/user-state-preserved.txt"
fi
printf 'complete=true\npassed=true\n' > "$EVIDENCE/lifecycle-result.txt"
