#!/usr/bin/env bash

: "${BINARY_NAME:="egctl"}"
: "${USE_SUDO:="true"}"
: "${EGCTL_INSTALL_DIR:="/usr/local/bin"}"

HAS_CURL="$(type "curl" &> /dev/null && echo true || echo false)"
HAS_WGET="$(type "wget" &> /dev/null && echo true || echo false)"
HAS_GIT="$(type "git" &> /dev/null && echo true || echo false)"

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS="$(uname|tr '[:upper:]' '[:lower:]')"

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*|cygwin*) OS='windows';;
  esac
}

# runs the given command as root (detects if we are root already)
runAsRoot() {
  if [ $EUID -ne 0 ] && [ "$USE_SUDO" == "true" ]; then
    sudo "${@}"
  else
    "${@}"
  fi
}

# verifySupported checks that the os/arch combination is supported for
# binary builds, as well whether or not necessary tools are present.
verifySupported() {
  local supported="darwin-amd64\ndarwin-arm64\nlinux-amd64\nlinux-arm64\n"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    echo "To build from source, go to https://github.com/envoyproxy/gateway"
    exit 1
  fi

  if [ "${HAS_CURL}" != "true" ] && [ "${HAS_WGET}" != "true" ]; then
    echo "Either curl or wget is required"
    exit 1
  fi

  if [ "${HAS_GIT}" != "true" ]; then
    echo "[WARNING] Could not find git. It is required for plugin installation."
  fi
}

# checkDesiredVersion checks if the desired version is available.
checkDesiredVersion() {
  if [ "$DESIRED_VERSION" == "" ]; then
    # Get tag from release URL
    local latest_release_url="https://github.com/envoyproxy/gateway/releases"
    if [ "${HAS_CURL}" == "true" ]; then
      TAG=$(curl -Ls $latest_release_url | grep 'href="/envoyproxy/gateway/releases/tag/v[0-9]*.[0-9]*.[0-9]*\"' | sed -E 's/.*\/envoyproxy\/gateway\/releases\/tag\/(v[0-9\.]+)".*/\1/g' | head -1)
    elif [ "${HAS_WGET}" == "true" ]; then
      TAG=$(wget $latest_release_url -O - 2>&1 | grep 'href="/envoyproxy/gateway/releases/tag/v[0-9]*.[0-9]*.[0-9]*\"' | sed -E 's/.*\/envoyproxy\/gateway\/releases\/tag\/(v[0-9\.]+)".*/\1/g' | head -1)
    fi
  else
    TAG=$DESIRED_VERSION
  fi
}

# checkEGCTLInstalledVersion checks which version of egctl is installed and
# if it needs to be changed.
checkEGCTLInstalledVersion() {
  if [[ -f "${EGCTL_INSTALL_DIR}/${BINARY_NAME}" ]]; then
    version=$("${EGCTL_INSTALL_DIR}/${BINARY_NAME}" version | grep -Eo "v[0-9]+\.[0-9]+.*" )
    if [[ "$version" == "$TAG" ]]; then
      echo "egctl ${version} is already ${DESIRED_VERSION:-latest}"
      return 0
    else
      echo "egctl ${TAG} is available. Changing from version ${version}."
      return 1
    fi
  else
    return 1
  fi
}

# downloadFile downloads the latest binary package
# for that binary.
downloadFile() {
  EGCTL_DIST="egctl_${TAG}_${OS}_${ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/envoyproxy/gateway/releases/download/$TAG/$EGCTL_DIST"
  EGCTL_TMP_ROOT="$(mktemp -dt egctl-installer-XXXXXX)"
  EGCTL_TMP_FILE="$EGCTL_TMP_ROOT/$EGCTL_DIST"
  echo "Downloading $DOWNLOAD_URL"
  if [ "${HAS_CURL}" == "true" ]; then
    curl -SsL "$DOWNLOAD_URL" -o "$EGCTL_TMP_FILE"
  elif [ "${HAS_WGET}" == "true" ]; then
    wget -q -O "$EGCTL_TMP_FILE" "$DOWNLOAD_URL"
  fi
}

# installFile installs the egctl binary.
installFile() {
  EGCTL_TMP="$EGCTL_TMP_ROOT/$BINARY_NAME"
  mkdir -p "$EGCTL_TMP"
  tar xf "$EGCTL_TMP_FILE" -C "$EGCTL_TMP"
  EGCTL_TMP_BIN="$EGCTL_TMP/bin/$OS/$ARCH/egctl"
  echo "Preparing to install $BINARY_NAME into ${EGCTL_INSTALL_DIR}"
  runAsRoot cp "$EGCTL_TMP_BIN" "$EGCTL_INSTALL_DIR/$BINARY_NAME"
  echo "$BINARY_NAME installed into $EGCTL_INSTALL_DIR/$BINARY_NAME"
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    if [[ -n "$INPUT_ARGUMENTS" ]]; then
      echo "Failed to install $BINARY_NAME with the arguments provided: $INPUT_ARGUMENTS"
      help
    else
      echo "Failed to install $BINARY_NAME"
    fi
    echo -e "\tFor support, go to https://github.com/envoyproxy/gateway."
  fi
  cleanup
  exit $result
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  set +e
  if ! [ "$(command -v $BINARY_NAME)" ]; then
    echo "$BINARY_NAME not found. Is $EGCTL_INSTALL_DIR on your PATH?"
    exit 1
  fi
  set -e
}

# help provides possible cli installation arguments
help () {
  echo "Accepted cli arguments are:"
  echo -e "\t[--help|-h ] ->> prints this help"
  echo -e "\t[--version|-v <desired_version>] . When not defined it fetches the latest release from GitHub"
  echo -e "\t e.g. --version v3.0.0 or -v latest"
  echo -e "\t[--no-sudo]  ->> install without sudo"
}

# cleanup temporary files.
cleanup() {
  if [[ -d "${EGCTL_TMP_ROOT:-}" ]]; then
    rm -rf "$EGCTL_TMP_ROOT"
  fi
}

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e

# Parsing input arguments (if any)
export INPUT_ARGUMENTS="${*}"
set -u
while [[ $# -gt 0 ]]; do
  case $1 in
    '--version'|-v)
       shift
       if [[ $# -ne 0 ]]; then
           export DESIRED_VERSION="${1}"
       else
           echo -e "Please provide the desired version. e.g. --version v3.0.0 or -v canary"
           exit 0
       fi
       ;;
    '--no-sudo')
       USE_SUDO="false"
       ;;
    '--help'|-h)
       help
       exit 0
       ;;
    *) exit 1
       ;;
  esac
  shift
done
set +u

initArch
initOS
verifySupported
checkDesiredVersion
if ! checkEGCTLInstalledVersion; then
  downloadFile
  installFile
fi
testVersion
cleanup
