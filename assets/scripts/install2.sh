#!/bin/sh

# chezmoi install script
# contains code from and inspired by
# https://github.com/client9/shlib
# https://github.com/goreleaser/godownloader

set -e

_logp=9
tmpdir=$(mktemp -d)
trap 'rm -rf ${tmpdir}' EXIT

usage() {
	this="$1"
	cat <<EOF
$this: download chezmoi and optionally run chezmoi

Usage: $this [-b bindir] [-d] [chezmoi-args]
  -b sets the installation directory, default is ${BINDIR}.
  -d enables debug logging.
If chezmoi-args are given, after install chezmoi is executed with chezmoi-args.
EOF
	exit 2
}

main() {
	BINDIR=${BINDIR:-./bin}
	EXECARGS=
	TAG=latest

	GOOS=$(get_os)
	GOARCH=$(get_arch)

	parse_args "$@"

	check_goos_goarch

	case "${OS}" in
	windows)
		BINSUFFIX=.exe
		FORMAT=zip
		;;
	*)
		BINSUFFIX=
		FORMAT=tar.gz
		;;
	esac

	realtag="$(real_tag ${TAG})"
	version
	name=chezmoi_${VERSION}
	log_info "downloading "

	test ! -d "${BINDIR}" && install -d "${BINDIR}"

	binary="chezmoi${BINSUFFIX}"
	install "${tmpdir}/${binary}" "${BINDIR}/"

	log_info "installed ${BINDIR}/${binexe}"

	# shellcheck disable=SC2086
	test -n "${EXECARGS}" && exec "${BINDIR}/${binary}" $EXECARGS
}

echo_error() {
	echo "$@" 1>&2
}

get_os() {
	os=$(uname -s | tr '[:upper:]' '[:lower:]')
	case "${os}" in
	cygwin_nt*) os="windows" ;;
	mingw*) os="windows" ;;
	msys_nt*) os="windows" ;;
	esac
	if [ "${os}" == "windows" ]; then
		BINSUFFIX=.exe
		FORMAT=zip
	fi
	echo "${os}"
}

get_arch() {
	arch=$(uname -m)
	case "${arch}" in
	386) arch="i386" ;;
	aarch64) arch="arm64" ;;
	armv*) arch="arm" ;;
	i386) arch="i386" ;;
	i686) arch="i386" ;;
	x86) arch="i386" ;;
	x86_64) arch="amd64" ;;
	esac
	echo "${arch}"
}

real_tag() {
	tag=$1
	log_debug "checking GitHub for tag ${tag}"
	giturl="https://github.com/twpayne/chezmoi/releases/${tag}"
	json=$(http_copy "${giturl}" "Accept: application/json")
	test -z "${json}" && return 1
	real_tag=$(echo "${json}" | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//')
	test -z "${real_tag}" && return 1
	log_info "found tag ${real_tag} for ${tag}"
	echo "${real_tag}"
}

http_copy() {
	tmp=$(mktemp)
	http_download "${tmp}" "$1" "$2" || return 1
	body=$(cat "${tmp}")
	rm -f "${tmp}"
	echo "${body}"
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

is_command() {
	command -v "$1" >/dev/null
}

log_debug() {
	[ 3 -le "${_logp}" ] || return 0
	echo debug "$@" 1>&2
}

log_info() {
	[ 2 -le "${_logp}" ] || return 0
	echo info "$@" 1>&2
}

log_err() {
	[ 1 -le "${_logp}" ] || return 0
	echo error "$@" 1>&2
}

log_crit() {
	[ 0 -le "${_logp}" ] || return 0
	echo critical "$@" 1>&2
}

parse_args() {
	while getopts "b:dh?" arg; do
		case "${arg}" in
		b) BINDIR="${OPTARG}" ;;
		d) _logp=0 ;;
		h | \?) usage "$0" ;;
		*) return 1 ;;
		esac
	done
	shift $((OPTIND - 1))
	EXECARGS="$*"
}

check_goos_goarch() {
	case "${GOOS}/${GOARCH}" in
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
		echo "${GOOS}/${GOARCH}: unsupported platform" 1>&2
		return 1
		;;
	esac
}

main "$@"
