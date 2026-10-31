#!/usr/bin/env python3
"""Shared local agent commands. Python 3.11+, no third-party dependencies."""

import argparse
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import sys
import tempfile
import tomllib

from policy import check_tool, edited_paths, file_reason, git

ROOT = Path(__file__).resolve().parents[2]
CHECK_TIMEOUT = 660


def current_task(root):
    pointer = root / ".harness/current-task"
    if not pointer.exists():
        return "unset (write tasks/T-NNN.md to .harness/current-task)"
    name = pointer.read_text().strip()
    candidate = (root / name).resolve()
    if not candidate.is_relative_to(root / "tasks") or not candidate.is_file():
        return "INVALID task pointer; expected an existing file under tasks/"
    return str(candidate.relative_to(root))


def context(root):
    branch = git(root, "branch", "--show-current") or "DETACHED"
    changes = git(root, "status", "--short") or "clean"
    active = []
    for task in sorted((root / "tasks").glob("T-*.md")):
        text = task.read_text()
        if any(marker in text for marker in ["**Result**: Open", "**Result**: Partial", "**Результат**: выбран", "**Результат**: В работе"]):
            active.append(task.stem)
    return (
        f"Repository: {root}\nBranch: {branch}\nCurrent task: {current_task(root)}\n"
        f"Open tasks: {', '.join(active) or 'none'}\nChanges:\n{changes[:4000]}\n"
        "Read: AGENT.md, .agents/HARNESS.md, PROJECT_CONTEXT.md, "
        "docs/IMPLEMENTATION_STATUS.md, docs/KNOWN_LIMITATIONS.md and the relevant docs/spec/.\n"
        "Code discovery: codebase-memory-mcp first. Verification: make harness-check."
    )


def doctor(root):
    failures = []
    print(f"Python {sys.version.split()[0]} (requires >=3.11)")
    for name in ["git", "make", "go", "gofmt", "node", "npm", "pg_isready", "codebase-memory-mcp", "codex", "claude"]:
        path = shutil.which(name)
        print(f"{'OK' if path else 'MISSING'} {name}: {path or 'not on PATH'}")
        if not path:
            failures.append(name)
    print(f"{'OK' if shutil.which('rtk') else 'OPTIONAL'} rtk")
    for name in [".claude/settings.json", ".codex/hooks.json"]:
        data = json.loads((root / name).read_text())
        if set(data["hooks"]) != {"SessionStart", "PreToolUse", "PostToolUse", "Stop"}:
            failures.append(name)
        print(f"OK JSON {name}")
    with (root / ".codex/config.toml").open("rb") as stream:
        tomllib.load(stream)
    print("OK TOML .codex/config.toml")
    for path in sorted((root / ".agents/skills").iterdir()):
        if not (path / "SKILL.md").is_file():
            failures.append(str(path))
    if not (root / "client/node_modules").is_dir():
        failures.append("client/node_modules: run npm ci in client/")
    if shutil.which("pg_isready"):
        database = os.environ.get("RF_TEST_ADMIN_URL", "postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable")
        result = subprocess.run(["pg_isready", "-d", database], capture_output=True, text=True, timeout=5)
        print(f"{'OK' if result.returncode == 0 else 'MISSING'} configured PostgreSQL listener")
        if result.returncode:
            failures.append("PostgreSQL: make db, or configure RF_TEST_ADMIN_URL")
    print("NOT VERIFIED: hook trust, account authorization, live MCP connectivity, E2E server.")
    for failure in failures:
        print(f"FAIL: {failure}")
    return 1 if failures else 0


def run_process(command, root, timeout=CHECK_TIMEOUT):
    # Use a new process group so a timeout also terminates make/go/node children.
    with tempfile.TemporaryFile(mode="w+b") as output:
        try:
            process = subprocess.Popen(command, cwd=root, stdout=output, stderr=subprocess.STDOUT, start_new_session=True)
        except OSError as error:
            return 127, str(error)
        try:
            code = process.wait(timeout=timeout)
        except subprocess.TimeoutExpired:
            os.killpg(process.pid, signal.SIGKILL)
            process.wait()
            code = 124
            output.write(b"\nHarness check timed out.\n")
        output.seek(0)
        return code, output.read().decode(errors="replace")


def run_checks(root):
    code, output = run_process(["make", "--no-print-directory", "harness-check"], root)
    state = root / ".harness"
    state.mkdir(exist_ok=True)
    (state / "check.log").write_text(output)
    return code, output


def format_files(root, cwd, tool, args):
    for name in edited_paths(tool, args):
        path = (cwd / name).resolve()
        if not path.is_file() or file_reason(root, cwd, name):
            continue
        if path.suffix == ".go":
            formatter = shutil.which("gofmt")
            if not formatter:
                return "gofmt is missing; formatting was not verified."
            result = subprocess.run([formatter, "-w", str(path)], capture_output=True, text=True, timeout=20)
            if result.returncode:
                return f"gofmt failed: {result.stderr[:2000]}"
        if path.suffix == ".md":
            for number, line in enumerate(path.read_text().splitlines(), 1):
                if line.rstrip() != line and not line.endswith("  "):
                    return f"{name}:{number}: trailing whitespace in Markdown."
    return None


def deny(reason):
    return {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": reason}}


def dispatch(root, payload):
    event = payload.get("hook_event_name")
    cwd = Path(payload.get("cwd", root)).resolve()
    if event == "SessionStart":
        return {"hookSpecificOutput": {"hookEventName": event, "additionalContext": context(root)}}
    if event == "PreToolUse":
        try:
            reason = check_tool(root, cwd, payload.get("tool_name"), payload.get("tool_input"))
        except (OSError, RuntimeError, ValueError, subprocess.TimeoutExpired) as error:
            reason = f"Harness policy could not verify the operation: {error}"
        return deny(reason) if reason else {}
    if event == "PostToolUse":
        args = payload.get("tool_input")
        if not isinstance(args, dict):
            return {"decision": "block", "reason": "Malformed tool_input during formatting."}
        reason = format_files(root, cwd, payload.get("tool_name"), args)
        return {"decision": "block", "reason": reason} if reason else {}
    if event == "Stop":
        if payload.get("stop_hook_active"):
            return {"systemMessage": "Repeated Stop: automatic checks skipped to prevent a loop. Completion is not verified; report failures or run make harness-check."}
        if not git(root, "status", "--porcelain"):
            return {}
        code, output = run_checks(root)
        if code:
            return {"decision": "block", "reason": f"make harness-check failed ({code}). See .harness/check.log. Fix failures or report the blocker explicitly.\n{output[-4000:]}"}
        return {}
    raise ValueError(f"Unsupported hook event: {event}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=["context", "doctor", "check", "hook"])
    args = parser.parse_args()
    try:
        if args.command == "context":
            print(context(ROOT))
        elif args.command == "doctor":
            return doctor(ROOT)
        elif args.command == "check":
            code, output = run_checks(ROOT)
            print(output, end="")
            return code
        else:
            payload = json.load(sys.stdin)
            if not isinstance(payload, dict):
                raise ValueError("Hook payload must be a JSON object")
            # Follow event cwd into worktrees; do not operate on the original checkout.
            root = Path(git(Path(payload.get("cwd", ROOT)), "rev-parse", "--show-toplevel")).resolve()
            print(json.dumps(dispatch(root, payload), ensure_ascii=False))
    except (OSError, RuntimeError, ValueError, subprocess.TimeoutExpired) as error:
        print(f"Harness error: {error}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())
