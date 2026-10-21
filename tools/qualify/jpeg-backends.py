#!/usr/bin/env python3
# SPDX-License-Identifier: MIT
"""Compare candidate jpegli with the current CLI; never select a new default."""

import argparse
import hashlib
import io
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import time
import tomllib
from typing import Any

from PIL import Image, ImageChops, ImageCms, ImageDraw, ImageOps, ImageStat, __version__, features


PHOTOS = ["jpeg/photographic/food-refocus.jpg", "jpeg/photographic/landscape-clouds.jpg"]
CONTROLS = ["jpeg/cmyk/jpeg-cmyk.jpg"] + [
    "jpeg/color-profile/" + name
    for name in ["jpeg-icc-gray.jpg", "jpeg-icc-cmyk.jpg", "jpeg-icc-display-p3.jpg"]
]
SRGB = ImageCms.ImageCmsProfile(ImageCms.createProfile("sRGB"))


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def describe(path: Path) -> dict[str, Any]:
    with Image.open(path) as image:
        profile = image.info.get("icc_profile", b"")
        return {"bytes": path.stat().st_size, "sha256": digest(path), "mode": image.mode,
                "size": list(image.size), "orientation": image.getexif().get(274, 1),
                "iccSHA256": hashlib.sha256(profile).hexdigest() if profile else None}


def viewed(path: Path) -> Image.Image:
    with Image.open(path) as source:
        profile = source.info.get("icc_profile")
        image = ImageOps.exif_transpose(source)
        if profile:
            return ImageCms.profileToProfile(image, ImageCms.ImageCmsProfile(io.BytesIO(profile)),
                                            SRGB, outputMode="RGB", renderingIntent=0)
        # Untagged RGB is assumed sRGB; untagged CMYK uses Pillow's unmanaged
        # conversion. Neither is a colorimetric guarantee about the original.
        return image.convert("RGB")


def differences(reference: Image.Image, candidate: Image.Image) -> dict[str, Any]:
    if reference.size != candidate.size:
        return {"omitted": "display dimensions differ"}
    stats = ImageStat.Stat(ImageChops.difference(reference, candidate))
    return {"meanAbsoluteChannelError": stats.mean, "channelRMSE": stats.rms,
            "maximumChannelError": [extent[1] for extent in stats.extrema]}


def crops(image: Image.Image) -> list[Image.Image]:
    width, height = image.size
    size = min(240, width, height)
    return [image.crop((x - size // 2, y - size // 2, x - size // 2 + size, y - size // 2 + size))
            for x, y in [(width // 2, height // 2), (width * 3 // 4, height // 4)]]


def sheet(path: Path, panels: list[tuple[str, list[Image.Image]]]) -> None:
    canvas = Image.new("RGB", (250 * len(panels), 530), "white")
    draw = ImageDraw.Draw(canvas)
    for column, (label, images) in enumerate(panels):
        draw.text((column * 250 + 4, 4), label, fill="black")
        for row, image in enumerate(images):
            canvas.paste(image, (column * 250 + 4, row * 255 + 24))
    canvas.save(path)


def run(options: argparse.Namespace) -> None:
    corpus = (options.source / "testdata/corpus").resolve()
    output = options.output.resolve()
    if output == corpus or corpus in output.parents:
        raise ValueError("comparison output must be outside the canonical corpus")
    manifest = tomllib.loads((corpus / "manifest.toml").read_text(encoding="utf-8"))
    entries = {entry["path"]: entry for entry in manifest["image"]}
    for relative in PHOTOS + CONTROLS:
        entry = entries[relative]
        if entry["license"] != "CC0-1.0" or not entry["redistributable"]:
            raise ValueError("comparison input lacks the selected legal-fixture grant")
        if digest(corpus / relative) != entry["sha256"]:
            raise ValueError("canonical input differs from the manifest: " + relative)
    tools = {name: path.resolve(strict=True) for name, path in
             [("smolpng", options.smolpng), ("jpegoptim", options.jpegoptim), ("jpegli", options.jpegli)]}
    output.parent.mkdir(parents=True, exist_ok=True)
    output.mkdir()  # Refuse an existing report or artifact directory.
    report: dict[str, Any] = {"schemaVersion": 1, "complete": False, "started": time.time(),
                              "jpegliDeclaredSource": options.jpegli_source,
                              "python": sys.version, "pillow": __version__,
                              "littleCMS": features.version("littlecms2"),
                              "tools": {name: {"path": str(path), "sha256": digest(path)}
                                        for name, path in tools.items()},
                              "commands": [], "photos": [], "controls": [],
                              "completionMeaning": "All comparison outcomes captured; failed candidate controls are not passing capabilities or backend acceptance.",
                              "interpretation": "Matched actual bytes and managed sRGB diagnostics; not perceptual acceptance. Untagged RGB assumes sRGB; untagged CMYK is unmanaged."}

    def save() -> None:
        staged = output / "report.json.tmp"
        staged.write_text(json.dumps(report, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        staged.replace(output / "report.json")

    save()
    environment = {key: value for key, value in os.environ.items()
                   if key.upper() not in {"TINIFY_API_KEY", "GH_TOKEN", "GITHUB_TOKEN"}}
    environment["OMP_NUM_THREADS"] = "1"

    def invoke(label: str, args: list[str], allow_failure: bool = False) -> dict[str, Any]:
        command: dict[str, Any] = {"label": label, "arguments": args, "started": time.time()}
        report["commands"].append(command)
        save()
        started = time.perf_counter()
        try:
            result = subprocess.run(args, capture_output=True, text=True, encoding="utf-8",
                                    errors="replace", timeout=300, env=environment,
                                    creationflags=subprocess.CREATE_NO_WINDOW if os.name == "nt" else 0)
            command.update(exitCode=result.returncode, stdout=result.stdout, stderr=result.stderr,
                           elapsedMS=(time.perf_counter() - started) * 1000)
            save()
        except Exception as error:
            command["error"] = str(error)
            save()
            raise
        if result.returncode and not allow_failure:
            raise RuntimeError(label + " failed: " + result.stderr + result.stdout)
        return command

    common = [str(tools["smolpng"]), "--home", str(output / "home"), "--no-project",
              "--jpegoptim-path", str(tools["jpegoptim"])]
    try:
        version = invoke("application-version", common + ["version", "--json"])
        report["applicationVersion"] = json.loads(version["stdout"])
        invoke("jpegli-help", [str(tools["jpegli"]), "--help"])
        shutil.copyfile(corpus / "LICENSE", output / "fixture-LICENSE")
        shutil.copyfile(corpus / "manifest.toml", output / "fixture-manifest.toml")
        for relative in PHOTOS:
            source = corpus / relative
            started = time.perf_counter()
            reference = viewed(source)
            stem = source.stem
            normalized = output / (stem + "-upright-srgb.png")
            profile = SRGB.tobytes() if describe(source)["iccSHA256"] else None
            reference.save(normalized, icc_profile=profile)
            photo: dict[str, Any] = {"input": relative, "original": describe(source),
                                     "prepared": describe(normalized),
                                     "preparationMS": (time.perf_counter() - started) * 1000,
                                     "rows": []}
            report["photos"].append(photo)
            panels = [("Original, upright", crops(reference))]
            for preset, quality in [("auto", 85), ("web", 80), ("smallest", 70)]:
                baseline = output / (stem + "-jpegoptim-" + preset + ".jpg")
                result = invoke(stem + "/jpegoptim/" + preset, common +
                                ["compress", str(source), "--preset", preset, "--jpeg-quality",
                                 f"{quality}-{quality}", "--output", str(baseline), "--copy", "none",
                                 "--offline", "--remote-policy", "never", "--jobs", "1", "--json"])
                parsed = json.loads(result["stdout"])
                if len(parsed["results"]) != 1 or parsed["results"][0].get("output") != str(baseline):
                    raise ValueError("baseline did not commit the requested output")
                budget = baseline.stat().st_size
                candidate = output / (stem + "-jpegli-" + preset + ".jpg")
                encoded = invoke(stem + "/jpegli/" + preset, [str(tools["jpegli"]), str(normalized),
                                                           str(candidate), "--target_size", str(budget)])
                first, second = viewed(baseline), viewed(candidate)
                row = {"preset": preset, "jpegoptimQuality": quality, "requestedJPEGliBytes": budget,
                       "jpegoptim": describe(baseline), "jpegli": describe(candidate),
                       "jpegliBudgetMet": candidate.stat().st_size <= budget,
                       "jpegoptimCommandMS": result["elapsedMS"], "jpegliCommandMS": encoded["elapsedMS"],
                       "jpegoptimDiagnostics": differences(reference, first),
                       "jpegliDiagnostics": differences(reference, second)}
                photo["rows"].append(row)
                panels.extend([(f"jpegoptim {quality}", crops(first)), ("jpegli @ " + preset, crops(second))])
                save()
            sheet(output / (stem + "-pixel-crops.png"), panels)
        for relative in CONTROLS:
            source = corpus / relative
            candidate = output / (source.stem + "-jpegli.jpg")
            result = invoke(source.stem + "/jpegli/control", [str(tools["jpegli"]), str(source),
                                                             str(candidate), "--quality", "85"], True)
            control: dict[str, Any] = {"input": relative, "original": describe(source),
                                       "observedExitCode": result["exitCode"],
                                       "status": "failed" if result["exitCode"] else "output"}
            report["controls"].append(control)
            if not result["exitCode"]:
                control["candidate"] = describe(candidate)
                control["exactProfilePreserved"] = control["original"]["iccSHA256"] == control["candidate"]["iccSHA256"]
                try:
                    control["diagnostics"] = differences(viewed(source), viewed(candidate))
                except Exception as error:
                    control["diagnosticsUnavailable"] = str(error)
            save()
        for relative in PHOTOS + CONTROLS:
            if digest(corpus / relative) != entries[relative]["sha256"]:
                raise ValueError("original changed: " + relative)
        report["complete"] = True
        report["ended"] = time.time()
        save()
    except Exception as error:
        report["error"] = str(error)
        save()
        raise


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ["source", "smolpng", "jpegoptim", "jpegli", "output"]:
        parser.add_argument("--" + name, type=Path, required=True)
    parser.add_argument("--jpegli-source", required=True)
    run(parser.parse_args())


if __name__ == "__main__":
    main()
