#!/usr/bin/env bash
# Regenerate images/agent-base/goose/ for cachi2 cargo prefetch.
#
# The full goose source is fetched at image build time from a prefetched
# tarball (see ocp-build-data images/mta-agent-base.yml). This directory
# only needs enough of the workspace for hermeto to resolve Cargo.lock.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTAINERFILE="${ROOT}/images/agent-base/Containerfile"
# Single source of truth: read the version declared in the Containerfile ARG.
GOOSE_VERSION="${GOOSE_VERSION:-$(grep -m1 '^ARG GOOSE_VERSION=' "${CONTAINERFILE}" | cut -d= -f2)}"
OUT="${ROOT}/images/agent-base/goose"
WORKDIR="$(mktemp -d)"
OUT_TMP="$(mktemp -d "${ROOT}/images/agent-base/goose.XXXXXX")"

cleanup() { rm -rf "${WORKDIR}" "${OUT_TMP}"; }
trap cleanup EXIT

git clone --depth 1 --branch "${GOOSE_VERSION}" https://github.com/aaif-goose/goose.git "${WORKDIR}/goose"

cp "${WORKDIR}/goose/Cargo.toml" "${WORKDIR}/goose/Cargo.lock" "${OUT_TMP}/"

while IFS= read -r manifest; do
	rel="${manifest#"${WORKDIR}/goose/"}"
	dest_dir="${OUT_TMP}/$(dirname "${rel}")"
	mkdir -p "${dest_dir}/src"
	cp "${manifest}" "${dest_dir}/"
	touch "${dest_dir}/src/lib.rs"
	if grep -q '^\[\[bin\]\]' "${manifest}"; then
		touch "${dest_dir}/src/main.rs"
	fi
done < <(find "${WORKDIR}/goose/crates" -name Cargo.toml)

mkdir -p "${OUT_TMP}/vendor/v8/src"
cp "${WORKDIR}/goose/vendor/v8/Cargo.toml" "${OUT_TMP}/vendor/v8/"
touch "${OUT_TMP}/vendor/v8/src/lib.rs"

# Swap atomically: only replace OUT once the new tree is fully written.
rm -rf "${OUT}"
mv "${OUT_TMP}" "${OUT}"

echo "Updated ${OUT} from aaif-goose/goose ${GOOSE_VERSION}"
