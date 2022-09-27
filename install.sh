#!/bin/sh

# Explanation https://gist.github.com/mohanpedala/1e2ff5661761d3abd0385e8223e16425

set -e

owner="benchkram"
name="bob"
base="https://github.com/${owner}"

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
_logp=7
log_set_priority() {
  _logp="$1"
}
log_priority() {
  if test -z "$1"; then
    echo "$_logp"
    return
  fi
  [ "$1" -le "$_logp" ]
}
log_tag() {
  case $1 in
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
  case "$os" in
  cygwin_nt*) os="windows" ;;
  mingw*) os="windows" ;;
  msys_nt*) os="windows" ;;
  esac
  echo "$os"
}
uname_arch() {
  arch=$(uname -m)
  case $arch in
  x86_64) arch="amd64" ;;
  x86) arch="386" ;;
  i686) arch="386" ;;
  i386) arch="386" ;;
  aarch64) arch="arm64" ;;
  armv5*) arch="armv5" ;;
  armv6*) arch="armv6" ;;
  armv7*) arch="armv7" ;;
  esac
  echo ${arch}
}
uname_os_check() {
  os=$(uname_os)
  case "$os" in
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
  log_crit "uname_os_check '$(uname -s)' got converted to '$os' which is not a GOOS value. Please file bug at https://github.com/client9/shlib"
  return 1
}
uname_arch_check() {
  arch=$(uname_arch)
  case "$arch" in
  386) return 0 ;;
  amd64) return 0 ;;
  arm64) return 0 ;;
  armv5) return 0 ;;
  armv6) return 0 ;;
  armv7) return 0 ;;
  ppc64) return 0 ;;
  ppc64le) return 0 ;;
  mips) return 0 ;;
  mipsle) return 0 ;;
  mips64) return 0 ;;
  mips64le) return 0 ;;
  s390x) return 0 ;;
  amd64p32) return 0 ;;
  esac
  log_crit "uname_arch_check '$(uname -m)' got converted to '$arch' which is not a GOARCH value.  Please file bug report at https://github.com/client9/shlib"
  return 1
}
untar() {
  tarball=$1
  case "${tarball}" in
  *.tar.gz | *.tgz) tar --no-same-owner -xzf "${tarball}" ;;
  *.tar) tar --no-same-owner -xf "${tarball}" ;;
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
  if [ -z "$header" ]; then
    code=$(curl -w '%{http_code}' -sL -o "$local_file" "$source_url")
  else
    code=$(curl -w '%{http_code}' -sL -H "$header" -o "$local_file" "$source_url")
  fi
  if [ "$code" != "200" ]; then
    log_debug "http_download_curl received HTTP status $code"
    return 1
  fi
  return 0
}
http_download_wget() {
  local_file=$1
  source_url=$2
  header=$3
  if [ -z "$header" ]; then
    wget -q -O "$local_file" "$source_url"
  else
    wget -q --header "$header" -O "$local_file" "$source_url"
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
  body=$(cat "$tmp")
  rm -f "${tmp}"
  echo "$body"
}
github_release() {
  owner_repo=$1
  version=$2
  test -z "$version" && version="latest"
  giturl="https://github.com/${owner_repo}/releases/${version}"
  json=$(http_copy "$giturl" "Accept:application/json")
  test -z "$json" && return 1
  version=$(echo "$json" | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//')
  test -z "$version" && return 1
  echo "$version"
}
hash_sha256() {
  TARGET=${1:-/dev/stdin}
  if is_command gsha256sum; then
    hash=$(gsha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command sha256sum; then
    hash=$(sha256sum "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command shasum; then
    hash=$(shasum -a 256 "$TARGET" 2>/dev/null) || return 1
    echo "$hash" | cut -d ' ' -f 1
  elif is_command openssl; then
    hash=$(openssl -dst openssl dgst -sha256 "$TARGET") || return 1
    echo "$hash" | cut -d ' ' -f a
  else
    log_crit "hash_sha256 unable to find command to compute sha-256 hash"
    return 1
  fi
}
hash_sha256_verify() {
  TARGET=$1
  checksums=$2
  if [ -z "$checksums" ]; then
    log_err "hash_sha256_verify checksum file not specified in arg2"
    return 1
  fi
  BASENAME=${TARGET##*/}
  want=$(grep "${BASENAME}" "${checksums}" 2>/dev/null | tr '\t' ' ' | cut -d ' ' -f 1)
  if [ -z "$want" ]; then
    log_err "hash_sha256_verify unable to find checksum for '${TARGET}' in '${checksums}'"
    return 1
  fi
  got=$(hash_sha256 "$TARGET")
  if [ "$want" != "$got" ]; then
    log_err "hash_sha256_verify checksum for '$TARGET' did not verify ${want} vs $got"
    return 1
  fi
}
cat /dev/null <<EOF
------------------------------------------------------------------------
End of functions from https://github.com/client9/shlib
------------------------------------------------------------------------
EOF

# Get OS
uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
  cygwin_nt*) os="windows" ;;
  mingw*) os="windows" ;;
  msys_nt*) os="windows" ;;
  esac
  echo "$os"
}

# Get architecture info
uname_arch() {
  arch=$(uname -m)
  case $arch in
  x86_64) arch="amd64" ;;
  x86) arch="386" ;;
  i686) arch="386" ;;
  i386) arch="386" ;;
  aarch64) arch="arm64" ;;
  armv5*) arch="armv5" ;;
  armv6*) arch="armv6" ;;
  armv7*) arch="armv7" ;;
  esac
  echo "$arch"
}

# Gets binary download URL from Github
# ex. https://github.com/benchkram/bob/releases/download/0.5.3/bob_0.5.3_linux_amd64
download_url() {
  version=$1
  os="$(uname_os)"
  arch="$(uname_arch)"
  url="${base}/${name}/releases/download/${version}/${name}_${version}_${os}_${arch}"
  echo "$url"
}

binary_name() {
  version=$1
  os="$(uname_os)"
  arch="$(uname_arch)"
  echo "${name}_${version}_${os}_${arch}"
}

# Get binaries checksums URL
# ex. https://github.com/benchkram/bob/releases/download/0.5.3/checksums.txt
checksums_url() {
  version=$1
  os="$(uname_os)"
  arch="$(uname_arch)"
  url="${base}/${name}/releases/download/${version}/checksums.txt"
  echo "$url"
}

# Validate supported OS
validate_os() {
  os="$(uname_os)"
  if [ "$os" != "linux" ] && [ "$os" != "darwin" ]; then
    log_debug "your system is not supported"
    return 1
  fi
}

guidelines_autocompletion() {
  echo "
To add auto-completion:

  BASH:
    * Add source <(bob completion) to your .bashrc

  ZSH:
    * Add source <(bob completion -z) to your .zshrc
"
}

validate_os

tmpdir=$(mktemp -d)

log_debug "downloading files into ${tmpdir}"

version="$(github_release "${owner}/${name}" "${BOB_VERSION}")"
bin_dir="${BIN_DIR:-./}"
bin_name="$(binary_name "$version")"

download_url="$(download_url "${version}")"
http_download "${tmpdir}/$bin_name" "${download_url}"

checksum_download_url="$(checksums_url "${version}")"
http_download "${tmpdir}/checksums.txt" "${checksum_download_url}"
hash_sha256_verify "${tmpdir}/$bin_name" "${tmpdir}/checksums.txt"

mv "${tmpdir}/$bin_name" "${tmpdir}/${name}"

sudo test ! -d "${bin_dir}" && install -d "${bin_dir}"
sudo install "${tmpdir}/${name}" "${bin_dir}"

rm -rf "${tmpdir}"

echo ""
echo ""

echo "the bin dir is $bin_dir"

if [ "$bin_dir" = './' ]; then
  echo "Successfully installed bob in $(pwd)"
else
  echo "Successfully installed bob in $(pwd)/$bin_dir"
fi

guidelines_autocompletion
