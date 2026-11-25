#!/usr/bin/env python3
"""Build a USB kit from the locked official distributions. Python 3.11+; Go on PATH."""
import argparse
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import shutil
import subprocess
import tarfile
import tempfile
import urllib.request
import zipfile

ROOT = Path(__file__).resolve().parents[1]
LOCK = json.loads((ROOT / "tools/runtimes.json").read_text())
DIST = ROOT / "dist"
KIT = DIST / "Takealong"
CACHE = ROOT / ".cache"


def sha(path):
    with path.open("rb") as handle:
        return hashlib.file_digest(handle, "sha256").hexdigest()


def tree_sha(root):
    digest = hashlib.sha256()
    for path in sorted(root.rglob("*"), key=lambda p: p.relative_to(root).as_posix()):
        if path.is_symlink():
            raise ValueError(f"unexpected runtime symlink: {path}")
        if path.is_file():
            digest.update((path.relative_to(root).as_posix() + "\0" + sha(path) + "\n").encode())
    return digest.hexdigest()


def download(spec):
    CACHE.mkdir(exist_ok=True)
    dest = CACHE / spec["archive"]
    if dest.is_file() and sha(dest) == spec["sha256"]:
        return dest
    url = f"https://github.com/openai/codex/releases/download/rust-v{LOCK['version']}/{dest.name}"
    with tempfile.NamedTemporaryFile(dir=CACHE, delete=False) as handle:
        tmp = Path(handle.name)
    try:
        with urllib.request.urlopen(url, timeout=120) as response, tmp.open("wb") as handle:
            shutil.copyfileobj(response, handle)
        if sha(tmp) != spec["sha256"]:
            raise ValueError(f"upstream checksum mismatch: {dest.name}")
        tmp.replace(dest)
    finally:
        tmp.unlink(missing_ok=True)
    return dest


def unpack(archive, dest):
    dest.mkdir(parents=True)
    with tarfile.open(archive, "r:gz") as src:
        for item in src:
            name = PurePosixPath(item.name)
            if (name.is_absolute() or ".." in name.parts or "\\" in item.name or ":" in item.name
                    or not (item.isfile() or item.isdir())):
                raise ValueError(f"invalid runtime entry: {item.name}")
            if item.isdir():
                (dest / name).mkdir(parents=True, exist_ok=True)
            else:
                path = dest / name
                path.parent.mkdir(parents=True, exist_ok=True)
                with src.extractfile(item) as inp, path.open("xb") as out:
                    shutil.copyfileobj(inp, out)
                path.chmod(0o755 if item.mode & 0o111 else 0o644)


def package(targets, local, make_zip):
    if KIT.exists():
        shutil.rmtree(KIT)
    KIT.mkdir(parents=True)
    for target in targets:
        spec = LOCK["targets"][target]
        dest = KIT / "runtime" / (target + "-x64")
        if local:
            # Accept an offline mirror only when every pathname and file hash matches.
            shutil.copytree(local / dest.name, dest, symlinks=True)
        else:
            unpack(download(spec), dest)
        if tree_sha(dest) != spec["treeSha256"]:
            raise ValueError(f"runtime contents do not match lock: {target}")
        exe = "takealong.exe" if target == "windows" else "takealong"
        env = {**os.environ, "CGO_ENABLED": "0", "GOOS": target, "GOARCH": "amd64"}
        subprocess.run(["go", "build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w", "-o", KIT / exe, "."],
                       cwd=ROOT, env=env, check=True)
    shutil.copytree(ROOT / "recipes", KIT / "recipes")
    shutil.copytree(ROOT / "licenses", KIT / "licenses")
    shutil.copy2(ROOT / "LICENSE", KIT / "licenses/Takealong-MIT.txt")
    shutil.copy2(ROOT / "docs/START.txt", KIT / "Start here.txt")
    if make_zip:
        dest = DIST / "takealong-usb.zip"
        with zipfile.ZipFile(dest, "w", zipfile.ZIP_DEFLATED, compresslevel=3, allowZip64=True) as out:
            for p in sorted(KIT.rglob("*")):
                if p.is_file():
                    # Preserve Linux executable bits even when the end user unpacks on a Mac.
                    out.write(p, "Takealong/" + p.relative_to(KIT).as_posix())
        digest = sha(dest)
        (DIST / "SHA256SUMS").write_text(f"{digest}  {dest.name}\n")
        print(f"{dest.name}: {dest.stat().st_size:,} bytes; SHA-256 {digest}")


if __name__ == "__main__":
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--target", choices=["linux", "windows", "both"], default="both")
    ap.add_argument("--runtime-tree", type=Path, help="optional previously extracted, hash-verified runtime mirror")
    ap.add_argument("--no-zip", action="store_true", help="only populate dist/Takealong")
    opts = ap.parse_args()
    package(["linux", "windows"] if opts.target == "both" else [opts.target], opts.runtime_tree, not opts.no_zip)
