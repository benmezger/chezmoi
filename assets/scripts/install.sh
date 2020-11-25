#!/bin/sh
set -e

usage() {
	this=$1
	cat <<EOF
$this: download chezmoi and optionally run chezmoi

Usage: $this [-b bindir] [-d] [-t tag] [chezmoi-args...]
  -b sets the installation directory, default is ./bin.
  -d turns on debug logging.
  -t sets the tag from https://github.com/twpayne/chezmoi/releases, default is latest.
If chezmoi-args is given, after install chezmoi is executed with chezmoi-args.
EOF
	exit 2
}

parse_args() {
	BINDIR=${BINDIR:-./bin}
	while getopts "b:dh?t:" arg; do
		case "${arg}" in
		b) BINDIR="${OPTARG}" ;;
		d) log_set_priority 10 ;;
		h | \?) usage "$0" ;;
		t) TAG="${OPTARG}" ;;
		esac
	done
	shift $((OPTIND - 1))
	EXECARGS="$*"
}
# this function wraps all the destructive operations
# if a curl|bash cuts off the end of the script due to
# network, either nothing will happen or will syntax error
# out preventing half-done work
execute() {
	tmpdir=$(mktemp -d)
	log_debug "downloading files into ${tmpdir}"
	http_download "${tmpdir}/${TARBALL}" "${TARBALL_URL}"
	http_download "${tmpdir}/${CHECKSUM}" "${CHECKSUM_URL}"
	hash_sha256_verify "${tmpdir}/${TARBALL}" "${tmpdir}/${CHECKSUM}"
	(cd "${tmpdir}" && untar "${TARBALL}")
	test ! -d "${BINDIR}" && install -d "${BINDIR}"
	binexe="chezmoi"
	if [ "${OS}" = "windows" ]; then
		binexe="${binexe}.exe"
	fi
	install "${tmpdir}/${binexe}" "${BINDIR}/"
	log_info "installed ${BINDIR}/${binexe}"
	rm -rf "${tmpdir}"
	# shellcheck disable=SC2086
	test -n "${EXECARGS}" && exec "${BINDIR}/${binexe}" ${EXECARGS}
}
platform_check() {
	case "${PLATFORM}" in
	darwin/amd64) return 0 ;;
	freebsd/386) return 0 ;;
	freebsd/amd64) return 0 ;;
	freebsd/arm) return 0 ;;
	freebsd/arm64) return 0 ;;
	linux/386) return 0 ;;
	linux/amd64) return 0 ;;
	linux/arm) return 0 ;;
	linux/arm64) return 0 ;;
	linux/ppc64) return 0 ;;
	linux/ppc64le) return 0 ;;
	openbsd/386) return 0 ;;
	openbsd/amd64) return 0 ;;
	openbsd/arm) return 0 ;;
	openbsd/arm64) return 0 ;;
	windows/386) return 0 ;;
	windows/amd64) return 0 ;;
	*)
		log_crit "platform ${PLATFORM} is not supported. Make sure this script is up-to-date and file request at https://github.com/twpayne/chezmoi/issues/new."
		return 1
		;;
	esac
}
tag_to_version() {
	if [ -z "${TAG}" ]; then
		log_info "checking GitHub for latest tag"
	else
		log_info "checking GitHub for tag '${TAG}'"
	fi
	REALTAG=$(github_release "twpayne/chezmoi" "${TAG}") && true
	if test -z "${REALTAG}"; then
		log_crit "unable to find tag '${TAG}'. Use 'latest' or a tag from https://github.com/twpayne/chezmoi/releases."
		exit 1
	fi
	# if version starts with 'v', remove it
	TAG="${REALTAG}"
	VERSION=${TAG#v}
}
adjust_format() {
	# change format (tar.gz or zip) based on OS
	case "${OS}" in
	windows) FORMAT=zip ;;
	esac
	true
}
adjust_os() {
	# adjust archive name based on OS
	case "${OS}" in
	386) OS=i386 ;;
	esac
	true
}
adjust_arch() {
	# adjust archive name based on ARCH
	case "${ARCH}" in
	386) ARCH=i386 ;;
	esac
	true
}

cat /dev/null <<EOF
------------------------------------------------------------------------
https://github.com/client9/shlib - portable posix shell functions
Public domain - http://unlicense.org
https://github.com/client9/shlib/blob/master/LICENSE.md
but credit (and pull requests) appreciated.
------------------------------------------------------------------------
EOF
is_command() {
	command -v "$1" >/dev/null
}
echoerr() {
	echo "$@" 1>&2
}
log_prefix() {
	echo "$0"
}
_logp=6
log_set_priority() {
	_logp="$1"
}
log_priority() {
	if test -z "$1"; then
		echo "${_logp}"
		return
	fi
	[ "$1" -le "${_logp}" ]
}
log_tag() {
	case "$1" in
	0) echo "emerg" ;;
	1) echo "alert" ;;
	2) echo "crit" ;;
	3) echo "err" ;;
	4) echo "warning" ;;
	5) echo "notice" ;;
	6) echo "info" ;;
	7) echo "debug" ;;
	*) echo "$1" ;;
	esac
}
log_debug() {
	log_priority 7 || return 0
	echoerr "$(log_prefix)" "$(log_tag 7)" "$@"
}
log_info() {
	log_priority 6 || return 0
	echoerr "$(log_prefix)" "$(log_tag 6)" "$@"
}
log_err() {
	log_priority 3 || return 0
	echoerr "$(log_prefix)" "$(log_tag 3)" "$@"
}
log_crit() {
	log_priority 2 || return 0
	echoerr "$(log_prefix)" "$(log_tag 2)" "$@"
}
uname_os() {
	os=$(uname -s | tr '[:upper:]' '[:lower:]')
	case "${os}" in
	cygwin_nt*) os="windows" ;;
	mingw*) os="windows" ;;
	msys_nt*) os="windows" ;;
	esac
	echo "${os}"
}
uname_arch() {
	arch=$(uname -m)
	case "${arch}" in
	x86_64) arch="amd64" ;;
	x86) arch="386" ;;
	i686) arch="386" ;;
	i386) arch="386" ;;
	aarch64) arch="arm64" ;;
	armv5*) arch="arm" ;;
	armv6*) arch="arm" ;;
	armv7*) arch="arm" ;;
	esac
	echo "${arch}"
}
uname_os_check() {
	os=$(uname_os)
	case "${os}" in
	darwin) return 0 ;;
	dragonfly) return 0 ;;
	freebsd) return 0 ;;
	linux) return 0 ;;
	android) return 0 ;;
	nacl) return 0 ;;
	netbsd) return 0 ;;
	openbsd) return 0 ;;
	plan9) return 0 ;;
	solaris) return 0 ;;
	windows) return 0 ;;
	esac
	log_crit "uname_os_check '$(uname -s)' got converted to '${os}' which is not a GOOS value. Please file bug at https://github.com/client9/shlib."
	return 1
}
uname_arch_check() {
	arch=$(uname_arch)
	case "${arch}" in
	386) return 0 ;;
	amd64) return 0 ;;
	arm64) return 0 ;;
	arm) return 0 ;;
	ppc64) return 0 ;;
	ppc64le) return 0 ;;
	mips) return 0 ;;
	mipsle) return 0 ;;
	mips64) return 0 ;;
	mips64le) return 0 ;;
	s390x) return 0 ;;
	amd64p32) return 0 ;;
	esac
	log_crit "uname_arch_check '$(uname -m)' got converted to '${arch}' which is not a GOARCH value.  Please file bug report at https://github.com/client9/shlib."
	return 1
}
untar() {
	tarball=$1
	case "${tarball}" in
	*.tar.gz | *.tgz) tar -xzf "${tarball}" ;;
	*.tar) tar -xf "${tarball}" ;;
	*.zip) unzip "${tarball}" ;;
	*)
		log_err "untar unknown archive format for ${tarball}"
		return 1
		;;
	esac
}
http_download_curl() {
	local_file=$1
	source_url=$2
	header=$3
	if [ -z "${header}" ]; then
		code=$(curl -w '%{http_code}' -sL -o "${local_file}" "${source_url}")
	else
		code=$(curl -w '%{http_code}' -sL -H "${header}" -o "${local_file}" "${source_url}")
	fi
	if [ "${code}" != "200" ]; then
		log_debug "http_download_curl received HTTP status ${code}"
		return 1
	fi
	return 0
}
http_download_wget() {
	local_file=$1
	source_url=$2
	header=$3
	if [ -z "${header}" ]; then
		wget -q -O "${local_file}" "${source_url}"
	else
		wget -q --header "${header}" -O "${local_file}" "${source_url}"
	fi
}
http_download() {
	log_debug "http_download $2"
	if is_command curl; then
		http_download_curl "$@"
		return
	elif is_command wget; then
		http_download_wget "$@"
		return
	fi
	log_crit "http_download unable to find wget or curl"
	return 1
}
http_copy() {
	tmp=$(mktemp)
	http_download "${tmp}" "$1" "$2" || return 1
	body=$(cat "${tmp}")
	rm -f "${tmp}"
	echo "${body}"
}
github_release() {
	owner_repo=$1
	version=$2
	test -z "${version}" && version="latest"
	giturl="https://github.com/${owner_repo}/releases/${version}"
	json=$(http_copy "${giturl}" "Accept:application/json")
	test -z "${json}" && return 1
	version=$(echo "${json}" | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//')
	test -z "${version}" && return 1
	echo "${version}"
}
hash_sha256() {
	target=${1:-/dev/stdin}
	if is_command gsha256sum; then
		hash=$(gsha256sum "${target}") || return 1
		echo "${hash}" | cut -d ' ' -f 1
	elif is_command sha256sum; then
		hash=$(sha256sum "${target}") || return 1
		echo "${hash}" | cut -d ' ' -f 1
	elif is_command shasum; then
		hash=$(shasum -a 256 "${target}" 2>/dev/null) || return 1
		echo "${hash}" | cut -d ' ' -f 1
	elif is_command sha256; then
		hash=$(sha256 -q "${target}" 2>/dev/null) || return 1
		echo "${hash}" | cut -d ' ' -f 1
	elif is_command openssl; then
		hash=$(openssl -dst openssl dgst -sha256 "${target}") || return 1
		echo "${hash}" | cut -d ' ' -f a
	else
		log_crit "hash_sha256 unable to find command to compute sha-256 hash"
		return 1
	fi
}
hash_sha256_verify() {
	target=$1
	checksums=$2
	if [ -z "${checksums}" ]; then
		log_err "hash_sha256_verify checksum file not specified in arg2"
		return 1
	fi
	basename=${target##*/}
	want=$(grep "${basename}" "${checksums}" 2>/dev/null | tr '\t' ' ' | cut -d ' ' -f 1)
	if [ -z "${want}" ]; then
		log_err "hash_sha256_verify unable to find checksum for '${target}' in '${checksums}'"
		return 1
	fi
	got=$(hash_sha256 "${target}")
	if [ "${want}" != "${got}" ]; then
		log_err "hash_sha256_verify checksum for '$target' did not verify ${want} vs $got"
		return 1
	fi
}
cat /dev/null <<EOF
------------------------------------------------------------------------
End of functions from https://github.com/client9/shlib
------------------------------------------------------------------------
EOF

FORMAT=tar.gz
OS=$(uname_os)
ARCH=$(uname_arch)

# use in logging routines
log_prefix() {
	echo "twpayne/chezmoi"
}
PLATFORM="${OS}/${ARCH}"
GITHUB_DOWNLOAD=https://github.com/twpayne/chezmoi/releases/download

uname_os_check "${OS}"
uname_arch_check "${ARCH}"

parse_args "$@"

platform_check

tag_to_version

adjust_format

adjust_os

adjust_arch

log_info "found version: ${VERSION} for ${TAG}/${OS}/${ARCH}"

NAME=chezmoi_${VERSION}_${OS}_${ARCH}
TARBALL=${NAME}.${FORMAT}
TARBALL_URL=${GITHUB_DOWNLOAD}/${TAG}/${TARBALL}
CHECKSUM=chezmoi_${VERSION}_checksums.txt
CHECKSUM_URL=${GITHUB_DOWNLOAD}/${TAG}/${CHECKSUM}

execute
