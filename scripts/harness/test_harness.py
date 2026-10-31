"""Behavioral checks run without an agent account or a database."""

import json
import os
import shutil
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

import harness
from policy import check_tool, edited_paths


class RepositoryTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="agent harness ")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name).resolve()
        self.git("init", "-b", "feature/harness")
        self.git("config", "user.email", "harness@example.invalid")
        self.git("config", "user.name", "Harness test")
        self.git("config", "commit.gpgsign", "false")
        self.git("config", "core.hooksPath", "/dev/null")
        self.write("AGENT.md", "Rules\n")
        self.write("server/migrations/0001_init.sql", "SELECT 1;\n")
        self.write("tasks/T-001.md", "- **Result**: Open\n")
        self.git("add", ".")
        self.git("commit", "-qm", "fixture")

    def git(self, *args):
        return subprocess.check_output(
            ["git", "-C", str(self.root), *args], stderr=subprocess.STDOUT, text=True
        ).strip()

    def write(self, name, text):
        path = self.root / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text)
        return path

    def policy(self, name, args, cwd=None):
        return check_tool(self.root, cwd or self.root, name, args)

    def event(self, event, **fields):
        return {"hook_event_name": event, "cwd": str(self.root), **fields}

    def test_context_reports_branch_task_and_dirty_state(self):
        self.write(".harness/current-task", "tasks/T-001.md\n")
        self.write("notes.md", "changed\n")
        text = harness.context(self.root)
        for expected in ["feature/harness", "T-001", "notes.md", "AGENT.md"]:
            self.assertIn(expected, text)

    def test_invalid_task_pointer_is_explicit(self):
        self.write(".harness/current-task", "../../outside.md\n")
        self.assertIn("INVALID", harness.context(self.root))

    def test_file_policy_allows_source_new_migration_and_env_example(self):
        for name in ["client/src/file.ts", "server/migrations/0002_new.sql", ".env.example"]:
            with self.subTest(name=name):
                self.assertIsNone(self.policy("Write", {"file_path": name}))

    def test_file_policy_blocks_secrets_outputs_and_committed_migrations(self):
        for name in [".env", "server/.env.local", "api/openapi.yaml",
                     "client/dist/main.js", "server/migrations/0001_init.sql"]:
            with self.subTest(name=name):
                self.assertIsNotNone(self.policy("Edit", {"file_path": name}))

    def test_staged_new_migration_remains_editable(self):
        self.write("server/migrations/0002_new.sql", "SELECT 2;\n")
        self.git("add", ".")
        self.assertIsNone(self.policy("Edit", {"file_path": "server/migrations/0002_new.sql"}))

    def test_nested_cwd_and_symlink_cannot_hide_secret(self):
        self.write(".env", "EXAMPLE=fixture\n")
        (self.root / "alias").symlink_to(self.root)
        self.assertIsNotNone(self.policy("Edit", {"file_path": "alias/.env"}))
        self.assertIsNotNone(self.policy("Write", {"file_path": "../.env"}, self.root / "server"))

    def test_patch_checks_deleted_and_moved_paths(self):
        for body in ["*** Delete File: .env", "*** Update File: ordinary.md\n*** Move to: .env",
                     "*** Update File: server/migrations/0001_init.sql\n*** Move to: new.sql"]:
            with self.subTest(body=body):
                self.assertIsNotNone(self.policy("apply_patch", {"command": "*** Begin Patch\n" + body + "\n*** End Patch"}))

    def test_patch_source_and_move_destination_are_returned(self):
        paths = edited_paths("apply_patch", {"command": "*** Begin Patch\n*** Update File: a.go\n*** Move to: b.go\n*** End Patch"})
        self.assertEqual(paths, ["a.go", "b.go"])

    def test_crlf_patch_cannot_hide_protected_paths(self):
        for name in [".env", "api/openapi.yaml", "server/migrations/0001_init.sql"]:
            command = f"*** Begin Patch\r\n*** Update File: {name}\r\n@@\r\n-a\r\n+b\r\n*** End Patch\r\n"
            with self.subTest(name=name):
                self.assertIsNotNone(self.policy("apply_patch", {"command": command}))

    def test_good_patch_and_ordinary_command_are_allowed(self):
        self.assertIsNone(self.policy("apply_patch", {"command": "*** Begin Patch\n*** Add File: source.py\n+x=1\n*** End Patch"}))
        self.assertIsNone(self.policy("Bash", {"command": "rtk git status --short"}))
        self.assertIsNone(self.policy("Bash", {"command": "python3 server/api/gen_openapi.py"}))

    def test_destructive_commands_are_denied(self):
        for command in ["rm -rf .", "rtk git reset --hard", "git push --force origin feature/harness",
                        "git push origin +HEAD:main", "git clean -fdx", "git status && rm -rf server",
                        "git -C server reset --hard", "env EXAMPLE=1 git push --force-with-lease"]:
            with self.subTest(command=command):
                self.assertIsNotNone(self.policy("Bash", {"command": command}))

    def test_multiline_command_does_not_hide_deletion(self):
        self.assertIsNotNone(self.policy("Bash", {"command": "git status\nrm -rf server"}))

    def test_relative_workdir_is_resolved_from_event_cwd(self):
        self.assertIsNotNone(self.policy("Bash", {"command": "rm -rf .", "workdir": "server"}))

    def test_removing_external_temporary_directory_is_allowed(self):
        with tempfile.TemporaryDirectory() as other:
            self.assertIsNone(self.policy("Bash", {"command": f"rm -rf '{other}/scratch'"}))

    def test_removing_repository_parent_is_denied(self):
        self.assertIsNotNone(self.policy("Bash", {"command": f"rm -rf '{self.root.parent}'"}))

    def test_commit_allowed_on_feature_and_denied_on_main(self):
        self.assertIsNone(self.policy("Bash", {"command": "git commit -m 'test'"}))
        self.git("branch", "-m", "main")
        self.assertIsNotNone(self.policy("Bash", {"command": "rtk git commit -m 'test'"}))

    def test_session_start_returns_context_json(self):
        result = harness.dispatch(self.root, self.event("SessionStart"))
        self.assertEqual(result["hookSpecificOutput"]["hookEventName"], "SessionStart")
        self.assertIn("feature/harness", result["hookSpecificOutput"]["additionalContext"])

    def test_pretool_returns_deny_without_auto_allow(self):
        result = harness.dispatch(self.root, self.event("PreToolUse", tool_name="Bash", tool_input={"command": "rm -rf ."}))
        self.assertEqual(result["hookSpecificOutput"]["permissionDecision"], "deny")
        result = harness.dispatch(self.root, self.event("PreToolUse", tool_name="Bash", tool_input={"command": "git status"}))
        self.assertEqual(result, {})

    def test_malformed_tool_input_is_denied(self):
        result = harness.dispatch(self.root, self.event("PreToolUse", tool_name="Edit", tool_input=[]))
        self.assertEqual(result["hookSpecificOutput"]["permissionDecision"], "deny")

    def test_stop_with_failure_blocks_and_includes_log(self):
        self.write("changed.md", "dirty\n")
        with patch.object(harness, "run_checks", return_value=(1, "fixture failure")):
            result = harness.dispatch(self.root, self.event("Stop", stop_hook_active=False))
        self.assertEqual(result["decision"], "block")
        self.assertIn("fixture failure", result["reason"])

    def test_repeated_stop_does_not_run_checks_or_claim_success(self):
        with patch.object(harness, "run_checks") as run:
            result = harness.dispatch(self.root, self.event("Stop", stop_hook_active=True))
        run.assert_not_called()
        self.assertIn("not verified", result["systemMessage"])

    def test_clean_stop_does_not_run_full_suite(self):
        with patch.object(harness, "run_checks") as run:
            result = harness.dispatch(self.root, self.event("Stop", stop_hook_active=False))
        run.assert_not_called()
        self.assertEqual(result, {})

    def test_dirty_stop_success_is_allowed(self):
        self.write("changed.md", "dirty\n")
        with patch.object(harness, "run_checks", return_value=(0, "checks passed")) as run:
            result = harness.dispatch(self.root, self.event("Stop"))
        run.assert_called_once_with(self.root)
        self.assertEqual(result, {})

    def test_posttool_formats_go_file(self):
        path = self.write("server/test.go", "package fixture\nfunc f(){ }\n")
        with patch.object(harness.shutil, "which", return_value="/fixture/gofmt"), patch.object(harness.subprocess, "run") as run:
            run.return_value.returncode = 0
            harness.dispatch(self.root, self.event("PostToolUse", tool_name="Write", tool_input={"file_path": str(path)}))
        self.assertEqual(run.call_args.args[0], ["/fixture/gofmt", "-w", str(path)])

    def test_missing_formatter_reports_failure(self):
        self.write("server/test.go", "package fixture\n")
        with patch.object(harness.shutil, "which", return_value=None):
            result = harness.dispatch(self.root, self.event("PostToolUse", tool_name="Write", tool_input={"file_path": "server/test.go"}))
        self.assertEqual(result["decision"], "block")
        self.assertIn("gofmt", result["reason"])

    def test_process_timeout_is_reported(self):
        code, output = harness.run_process(
            [sys.executable, "-c", "import time; time.sleep(10)"], self.root, timeout=0.05
        )
        self.assertEqual(code, 124)
        self.assertIn("timed out", output)

    def test_missing_check_executable_fails(self):
        code, output = harness.run_process(["/missing/harness-command"], self.root, timeout=1)
        self.assertNotEqual(code, 0)
        self.assertIn("missing", output)

    def test_doctor_uses_configured_database_without_printing_credentials(self):
        self.write(".claude/settings.json", json.dumps({"hooks": {name: [] for name in ["SessionStart", "PreToolUse", "PostToolUse", "Stop"]}}))
        self.write(".codex/hooks.json", (self.root / ".claude/settings.json").read_text())
        self.write(".codex/config.toml", "[features]\nhooks = true\n")
        (self.root / ".agents/skills").mkdir(parents=True)
        (self.root / "client/node_modules").mkdir(parents=True)
        url = "postgres://fixture:never-print-this@127.0.0.1:55432/postgres"
        with patch.dict(os.environ, {"RF_TEST_ADMIN_URL": url}), patch.object(harness.shutil, "which", return_value="/fixture/bin"), patch.object(harness.subprocess, "run") as run, patch("builtins.print") as output:
            run.return_value.returncode = 0
            self.assertEqual(harness.doctor(self.root), 0)
        self.assertEqual(run.call_args.args[0], ["pg_isready", "-d", url])
        self.assertNotIn("never-print-this", str(output.call_args_list))


class ConfigurationTest(unittest.TestCase):
    def test_both_clients_register_all_four_events(self):
        root = Path(__file__).resolve().parents[2]
        for name in [".claude/settings.json", ".codex/hooks.json"]:
            with self.subTest(name=name):
                hooks = json.loads((root / name).read_text())["hooks"]
                self.assertEqual(set(hooks), {"SessionStart", "PreToolUse", "PostToolUse", "Stop"})
                self.assertEqual(hooks["Stop"][0]["hooks"][0]["timeout"], 720)

    def test_real_hook_commands_from_nested_directory_with_spaces(self):
        fixture = RepositoryTest()
        fixture.setUp()
        self.addCleanup(fixture.doCleanups)
        root = Path(__file__).resolve().parents[2]
        for name in ["harness.py", "policy.py"]:
            target = fixture.root / "scripts/harness" / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(root / "scripts/harness" / name, target)
        nested = fixture.root / "nested directory"
        nested.mkdir()
        fixture.write("Makefile", "harness-check:\n\t@echo fixture-check-failure\n\t@exit 7\n")
        for config in [".claude/settings.json", ".codex/hooks.json"]:
            hooks = json.loads((root / config).read_text())["hooks"]
            events = [
                ("SessionStart", {}, "hookSpecificOutput"),
                ("PreToolUse", {"tool_name": "Bash", "tool_input": {"command": "rm -rf ."}}, "hookSpecificOutput"),
                ("PostToolUse", {"tool_name": "Write", "tool_input": {"file_path": "note.md"}}, None),
                ("Stop", {"stop_hook_active": True}, "systemMessage"),
                ("Stop", {"stop_hook_active": False}, "decision"),
            ]
            for event, fields, key in events:
                with self.subTest(config=config, event=event):
                    command = hooks[event][0]["hooks"][0]["command"]
                    payload = {"hook_event_name": event, "cwd": str(nested), **fields}
                    result = subprocess.run(command, shell=True, cwd=nested, input=json.dumps(payload), capture_output=True, text=True, timeout=10)
                    self.assertEqual(result.returncode, 0, result.stderr)
                    data = json.loads(result.stdout)
                    if key:
                        self.assertIn(key, data)
                        if event == "Stop" and not fields.get("stop_hook_active"):
                            self.assertEqual(data["decision"], "block")
                            self.assertIn("fixture-check-failure", data["reason"])
                    else:
                        self.assertEqual(data, {})


if __name__ == "__main__":
    unittest.main()
