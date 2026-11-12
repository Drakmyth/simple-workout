#!/usr/bin/env python3

import argparse
import re
from pathlib import Path

START_MARKER = "<!-- container-image:start -->"
END_MARKER = "<!-- container-image:end -->"
TOKEN_PATTERN = re.compile(r"{{[A-Z_]+}}")


def prepare_release(notes_path: Path, include_container: bool) -> str:
    notes = notes_path.read_text() if notes_path.is_file() else ""
    parts = [notes.rstrip()] if notes.strip() else []
    if include_container:
        parts.append(f"{START_MARKER}\n{END_MARKER}")
    return "\n\n".join(parts) + ("\n" if parts else "")


def render_container(template: str, version: str, digest: str) -> str:
    digest_note = f"Digest: `{digest}`" if digest else ""
    rendered = template.replace("{{VERSION}}", version).replace("{{DIGEST}}", digest_note)
    unresolved = TOKEN_PATTERN.findall(rendered)
    if unresolved:
        raise ValueError(f"unresolved template tokens: {', '.join(unresolved)}")
    return rendered.rstrip() + "\n"


def insert_container(release_notes: str, container_notes: str) -> str:
    if release_notes.count(START_MARKER) != 1 or release_notes.count(END_MARKER) != 1:
        raise ValueError("release notes must contain exactly one container marker pair")
    before, remainder = release_notes.split(START_MARKER, 1)
    _, after = remainder.split(END_MARKER, 1)
    container = container_notes.strip()
    return f"{before}{START_MARKER}\n\n{container}\n\n{END_MARKER}{after}"


def write(path: Path, content: str) -> None:
    path.write_text(content)


def main() -> None:
    parser = argparse.ArgumentParser(description="Compose Simple Workout release notes")
    commands = parser.add_subparsers(dest="command", required=True)

    prepare = commands.add_parser("prepare-release")
    prepare.add_argument("--notes", type=Path, required=True)
    prepare.add_argument("--include-container", choices=("true", "false"), required=True)
    prepare.add_argument("--output", type=Path, required=True)

    render = commands.add_parser("render-container")
    render.add_argument("--template", type=Path, required=True)
    render.add_argument("--version", required=True)
    render.add_argument("--digest", default="")
    render.add_argument("--output", type=Path, required=True)

    insert = commands.add_parser("insert-container")
    insert.add_argument("--release", type=Path, required=True)
    insert.add_argument("--container", type=Path, required=True)

    args = parser.parse_args()
    if args.command == "prepare-release":
        content = prepare_release(args.notes, args.include_container == "true")
        write(args.output, content)
    elif args.command == "render-container":
        content = render_container(args.template.read_text(), args.version, args.digest)
        write(args.output, content)
    else:
        content = insert_container(args.release.read_text(), args.container.read_text())
        write(args.release, content)


if __name__ == "__main__":
    main()
