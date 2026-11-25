#!/usr/bin/env python3
"""Exercise the real packaged CLI with a private test home and a loopback responder."""
import gzip
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from threading import Thread

MARK = "KITWAKE_SMOKE_OK"


class Responses(BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def do_GET(self):
        self.send_error(404)

    def do_POST(self):
        if self.path != "/v1/responses":
            self.send_error(404)
            return
        raw = self.rfile.read(int(self.headers.get("Content-Length", 0)))
        if self.headers.get("Content-Encoding") == "gzip":
            raw = gzip.decompress(raw)
        if b"Give me a short briefing on this folder." not in raw:
            self.send_error(422, "Expected the packaged recipe")
            return
        msg = {"id": "msg_1", "type": "message", "status": "completed", "role": "assistant",
               "content": [{"type": "output_text", "text": MARK, "annotations": []}]}
        resp = {"id": "resp_1", "object": "response", "created_at": 1, "status": "completed",
                "model": "kitwake-test", "output": [msg], "usage": {"input_tokens": 1, "output_tokens": 1, "total_tokens": 2}}
        events = [
            {"type": "response.created", "response": {**resp, "status": "in_progress", "output": []}},
            {"type": "response.output_item.added", "output_index": 0, "item": {**msg, "status": "in_progress", "content": []}},
            {"type": "response.output_text.delta", "item_id": "msg_1", "output_index": 0, "content_index": 0, "delta": MARK},
            {"type": "response.output_item.done", "output_index": 0, "item": msg},
            {"type": "response.completed", "response": resp},
        ]
        body = "".join(f"event: {e['type']}\ndata: {json.dumps(e)}\n\n" for e in events).encode()
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


def smoke(root):
    exe = root.resolve() / ("kitwake.exe" if os.name == "nt" else "kitwake")
    with tempfile.TemporaryDirectory() as tmp, ThreadingHTTPServer(("127.0.0.1", 0), Responses) as srv:
        Thread(target=srv.serve_forever, daemon=True).start()
        try:
            home = Path(tmp) / "home"
            codex = home / ".codex"
            project = Path(tmp) / "folder with spaces"
            project.mkdir()
            codex.mkdir(parents=True)
            (codex / "config.toml").write_text(
                'model = "kitwake-test"\nmodel_provider = "localtest"\ncheck_for_update_on_startup = false\n'
                '[model_providers.localtest]\nname = "Loopback"\nwire_api = "responses"\nrequires_openai_auth = false\n'
                f'base_url = "http://127.0.0.1:{srv.server_port}/v1"\n')
            # Start with an allowlist; never pass a developer's provider credentials into this test.
            allowed = {"PATH","SYSTEMROOT","WINDIR","SYSTEMDRIVE","TEMP","TMP","TMPDIR","COMSPEC","PATHEXT",
                       "PROCESSOR_ARCHITECTURE","NUMBER_OF_PROCESSORS","LANG","LC_ALL"}
            env = {k:v for k,v in os.environ.items() if k.upper() in allowed}
            env.update(HOME=str(home), USERPROFILE=str(home), CODEX_HOME=str(codex), NO_PROXY="127.0.0.1",
                       LOCALAPPDATA=str(home/"AppData"/"Local"), APPDATA=str(home/"AppData"/"Roaming"))
            def run(*args):
                p = subprocess.run([str(exe), *args], cwd=project, env=env, capture_output=True, text=True, timeout=100)
                if p.returncode:
                    raise SystemExit(p.stderr[-4000:] or p.stdout[-1000:])
                return p.stdout
            if "USB bundle" not in run("doctor"):
                raise SystemExit("doctor did not select the bundle")
            if MARK not in run("brief", str(project)):
                raise SystemExit("the model response did not reach the terminal")
            if not list((codex / "sessions").rglob("*.jsonl")):
                raise SystemExit("the agent did not save a session")
            print("PASS: bundled Codex received the recipe, printed its answer, and saved a session.")
        finally:
            srv.shutdown()


if __name__ == "__main__":
    smoke(Path(sys.argv[1]))
