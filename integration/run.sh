#!/bin/sh
# Integration checks for yup-ls, run inside a Debian (GNU coreutils) container.
#
# ls output is highly format-dependent, so comparisons are pinned to a
# deterministic, line-oriented shape: LC_ALL=C for stable byte sort and GNU
# `ls -1` (one entry per line, no columns/color) as the reference.
#
# parity ARGS  — yup-ls ARGS must be byte-identical to `ls -1 ARGS` over a
#                known fixture tree (used where yup-ls matches GNU exactly).
# assert WANT  — yup-ls must produce WANT exactly (used for the documented
#                divergences -a, -R, -l; see cmd-ls COMPATIBILITY.md).
set -eu

export LC_ALL=C

fails=0

# Build a known fixture tree so every listing is deterministic.
root=$(mktemp -d)
mkdir -p "$root/dir/sub"
: > "$root/dir/alpha.txt"
: > "$root/dir/bravo.txt"
: > "$root/dir/.hidden"
printf 'alpha' > "$root/dir/sized.txt"
: > "$root/dir/sub/b.txt"
# A directory holding exactly one regular file, for a fully deterministic -l
# assertion (only the file's mode and byte size, both stable, are emitted).
mkdir -p "$root/only"
printf 'alpha' > "$root/only/sized.txt"

parity() {
	ours=$(yup-ls "$@" 2>/dev/null || true)
	gnu=$(ls -1 "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  ls -1 %s\n' "$*"
	else
		printf 'FAIL  parity  ls -1 %s\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

assert() {
	want=$1
	shift
	got=$(yup-ls "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  ls %s\n' "$*"
	else
		printf 'FAIL  assert  ls %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# Default listing: one name per line, sorted, hidden entries excluded — matches
# GNU `ls -1` exactly.
parity "$root/dir"

# Default path is the current directory.
cd "$root/dir"
parity

# -a: yup-ls lists dotfiles but, unlike GNU, omits the synthetic "." and ".."
# entries (afero.ReadDir does not surface them). Documented divergence.
assert "$(printf '.hidden\nalpha.txt\nbravo.txt\nsized.txt\nsub')" -a "$root/dir"

# -R: yup-ls walks recursively but emits a single flat list of paths relative to
# the listing root, with no per-directory ".:" headers or blank-line grouping
# that GNU `ls -R` prints. Documented divergence.
assert "$(printf 'alpha.txt\nbravo.txt\nsized.txt\nsub\nsub/b.txt')" -R "$root/dir"

# -l: yup-ls long format is "<perm> <size> <name>" only — no owner, group, link
# count, or mtime, and no leading "total" line. Documented divergence. Listed
# against a single-file directory so mode and byte size are fully deterministic.
assert "-rw-r--r-- 5 sized.txt" -l "$root/only"

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
