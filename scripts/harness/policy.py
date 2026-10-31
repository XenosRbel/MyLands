"""Small guardrails for direct edits and recognizable shell commands, not a sandbox."""

from pathlib import Path
import re
import shlex
import subprocess


def git(root, *args):
    result = subprocess.run(
        ["git", "-C", str(root), *args], capture_output=True, text=True, timeout=10
    )
    if result.returncode:
        raise RuntimeError(result.stderr.strip() or "Git command failed")
    return result.stdout.strip()


def edited_paths(tool, args):
    if tool == "apply_patch":
        command = args.get("command", args.get("patch", ""))
        if not isinstance(command, str):
            return []
        normalized = "\n".join(command.splitlines())
        return re.findall(r"^\*\*\* (?:Add File|Update File|Delete File|Move to): (.+)$", normalized, re.M)
    path = args.get("file_path", args.get("notebook_path"))
    return [path] if isinstance(path, str) and path else []


def file_reason(root, cwd, name):
    path = (cwd / name).resolve()
    try:
        relative = path.relative_to(root)
    except ValueError:
        return "Direct edits outside this repository are blocked."
    parts = relative.parts
    if not parts:
        return "A file path is required."
    if any(part == ".env" or (part.startswith(".env.") and part != ".env.example") for part in parts):
        return "Secret environment files must not be edited by an agent."
    if path.suffix in {".pem", ".key", ".p12", ".pfx"} or parts[0] == ".git":
        return "Secret keys and Git internals are protected."
    if relative.as_posix() == "api/openapi.yaml":
        return "Generate api/openapi.yaml with server/api/gen_openapi.py."
    if any(part in {"node_modules", "dist"} for part in parts) or parts[:2] == ("server", "bin"):
        return "Build outputs and installed dependencies are protected."
    if parts[:2] == ("server", "migrations"):
        # HEAD is the conservative baseline; staged new migrations remain editable.
        tracked = git(root, "ls-tree", "--name-only", "HEAD", "--", relative.as_posix())
        if tracked:
            return "Committed migrations are append-only. Add a new migration."
    return None


def unwrapped(tokens):
    tokens = list(tokens)
    while tokens:
        base = Path(tokens[0]).name
        if base in {"rtk", "env", "command", "sudo"}:
            tokens.pop(0)
            if tokens and tokens[0] == "proxy":
                tokens.pop(0)
        elif re.match(r"^[A-Za-z_][A-Za-z_0-9]*=", tokens[0]):
            tokens.pop(0)
        else:
            break
    return tokens


def shell_reason(root, cwd, command):
    lexer = shlex.shlex(command, posix=True, punctuation_chars=";&|()\n")
    lexer.whitespace_split = True
    lexer.whitespace = " \t\r"
    segments = []
    segment = []
    for token in lexer:
        if token and all(ch in ";&|()\n" for ch in token):
            segments.append(segment)
            segment = []
        else:
            segment.append(token)
    segments.append(segment)
    for segment in segments:
        tokens = unwrapped(segment)
        if not tokens:
            continue
        executable = Path(tokens[0]).name
        args = tokens[1:]
        if executable == "cd" and args:
            cwd = (cwd / args[0]).resolve()
            continue
        if executable == "git":
            target = cwd
            while args and args[0].startswith("-"):
                option = args.pop(0)
                if option == "-C" and args:
                    target = (target / args.pop(0)).resolve()
                elif option in {"-c", "--git-dir", "--work-tree"} and args:
                    args.pop(0)
            if not args:
                continue
            operation, *flags = args
            if operation == "push" and any(
                arg in {"-f", "--force"} or arg.startswith(("--force-with-lease", "+")) for arg in flags
            ):
                return "Force pushes are blocked."
            if operation == "reset" and "--hard" in flags:
                return "git reset --hard is blocked."
            if operation == "clean" and any(arg.startswith("-") and "f" in arg for arg in flags):
                return "Forced git clean is blocked."
            if operation == "commit":
                branch = git(target, "branch", "--show-current")
                if not branch or branch in {"main", "master"} or branch.startswith("release/"):
                    return "Commit on a feature branch, not a protected or detached branch."
        if executable == "rm":
            options = [arg for arg in args if arg.startswith("-") and arg != "--"]
            recursive = any(arg in {"--recursive"} or (arg.startswith("-") and not arg.startswith("--") and "r" in arg.lower()) for arg in options)
            if not recursive:
                continue
            for arg in args:
                if arg.startswith("-"):
                    continue
                # Treat an unexpanded glob conservatively within the repository.
                prefix = re.split(r"[*?\[]", arg)[0]
                target = (cwd / prefix).resolve()
                if target.is_relative_to(root) or root.is_relative_to(target):
                    return "Recursive deletion in or above this repository is blocked."
    return None


def check_tool(root, cwd, tool, args):
    if not isinstance(args, dict):
        return "Malformed tool_input: expected an object."
    if tool in {"Bash", "exec_command"}:
        command = args.get("command", args.get("cmd"))
        if not isinstance(command, str):
            return "Malformed shell command."
        actual_cwd = (cwd / args.get("cwd", args.get("workdir", cwd))).resolve()
        return shell_reason(root, actual_cwd, command)
    if tool in {"Edit", "Write", "NotebookEdit", "apply_patch"}:
        paths = edited_paths(tool, args)
        if not paths:
            return "No supported file path in edit input."
        for name in paths:
            reason = file_reason(root, cwd, name)
            if reason:
                return f"{name}: {reason}"
    return None
