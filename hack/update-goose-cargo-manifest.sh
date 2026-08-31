#!/usr/bin/env bash
# Regenerate images/agent-base/goose/ for cachi2 cargo prefetch.
#
# The full goose source is fetched at image build time from a prefetched
# tarball (see ocp-build-data images/mta-agent-base.yml). This directory
# only needs enough of the workspace for hermeto to resolve Cargo.lock.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GOOSE_VERSION="${GOOSE_VERSION:-v1.45.0}"
OUT="${ROOT}/images/agent-base/goose"
WORKDIR="$(mktemp -d)"

cleanup() { rm -rf "${WORKDIR}"; }
trap cleanup EXIT

git clone --depth 1 --branch "${GOOSE_VERSION}" https://github.com/aaif-goose/goose.git "${WORKDIR}/goose"

rm -rf "${OUT}"
mkdir -p "${OUT}"
cp "${WORKDIR}/goose/Cargo.toml" "${WORKDIR}/goose/Cargo.lock" "${OUT}/"

while IFS= read -r manifest; do
	rel="${manifest#${WORKDIR}/goose/}"
	dest_dir="${OUT}/$(dirname "${rel}")"
	mkdir -p "${dest_dir}/src"
	cp "${manifest}" "${dest_dir}/"
	touch "${dest_dir}/src/lib.rs"
	if grep -q '^\[\[bin\]\]' "${manifest}"; then
		touch "${dest_dir}/src/main.rs"
	fi
done < <(find "${WORKDIR}/goose/crates" -name Cargo.toml)

mkdir -p "${OUT}/vendor/v8/src"
cp "${WORKDIR}/goose/vendor/v8/Cargo.toml" "${OUT}/vendor/v8/"
touch "${OUT}/vendor/v8/src/lib.rs"

echo "Updated ${OUT} from aaif-goose/goose ${GOOSE_VERSION}"
