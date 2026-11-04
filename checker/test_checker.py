import base64
import contextlib
import io
import json
import os
import unittest
import uuid
from pathlib import Path
from unittest.mock import patch

from checker import checker
from checker.protocol import Failure, State, Status

FLAGS = {place: chr(64 + place) * 31 + "=" for place in range(1, 11)}
PNG = b"\x89PNG\r\n\x1a\nchecker"


class FakeClient:
    """Small in-memory Wahrwelt API used to exercise checker flows."""

    def __init__(self):
        self.calls = []
        self.users = {}
        self.sessions = {}
        self.cases = {}
        self.manifests = {}
        self.certificates = {}
        self.invoices = {}
        self.inspections = {}
        self.licenses = {}
        self.tokens = {}
        self.contracts = {}
        self.quotas = {}
        self.waybills = {}
        self.gate_passes = {}
        self.route_slots = {}
        self.approvals = {}
        self.previews = {}
        self.sequence = 0
        self.wrong_secret = False
        self.unavailable_stored = False
        self.bad_preview = False

        for username in ("carrier", "inspector"):
            self.users[username] = {
                "password": username,
                "user_id": self.next_uuid(),
            }

    def next_uuid(self):
        self.sequence += 1
        return str(uuid.UUID(int=self.sequence))

    def number(self, prefix):
        return f"{prefix}-{self.next_uuid()}"

    def user_body(self, username):
        return {"user_id": self.users[username]["user_id"], "username": username}

    def register(self, username, password):
        self.calls.append(
            {"method": "POST", "path": "/api/v1/register", "username": username}
        )
        if username in self.users:
            raise Failure(
                Status.MUMBLE, "Service rejected checker account", "register failed"
            )
        self.users[username] = {"password": password, "user_id": self.next_uuid()}
        return self.user_body(username)

    def login(self, username, password):
        self.calls.append({"method": "POST", "path": "/api/v1/login"})
        user = self.users.get(username)
        if user is None or user["password"] != password:
            raise Failure(
                Status.MUMBLE, "Service rejected checker login", "login failed"
            )
        cookie = f"session={self.next_uuid()}"
        self.sessions[cookie] = username
        return self.user_body(username), cookie

    def request_json(
        self,
        method,
        path,
        expected,
        payload=None,
        cookie=None,
        bearer=None,
        headers=None,
        stored=False,
    ):
        username = self.sessions.get(cookie)
        call = {
            "method": method,
            "path": path,
            "expected": expected,
            "payload": payload,
            "cookie": cookie,
            "username": username,
            "bearer": bearer,
            "headers": headers,
            "stored": stored,
        }
        self.calls.append(call)

        if stored and self.unavailable_stored:
            raise Failure(
                Status.CORRUPT, "Stored flag is unavailable", "stored object is absent"
            )
        if method == "GET" and path == "/healthz":
            return {"status": "ok"}

        if method == "POST" and path == "/api/v1/cases":
            case_id = self.next_uuid()
            self.cases[case_id] = {"owner": username, "case_id": case_id, **payload}
            return {
                "case_id": case_id,
                "owner_id": self.users[username]["user_id"],
                "kind": payload["kind"],
                "origin": payload["origin"],
                "destination": payload["destination"],
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "GET" and path.startswith("/api/v1/cases/"):
            case_id = path.rsplit("/", 1)[1]
            item = self.cases[case_id]
            if item["owner"] != username:
                raise Failure(
                    Status.CORRUPT, "Stored flag is unavailable", "case is absent"
                )
            return {
                "case_id": case_id,
                "owner_id": self.users[username]["user_id"],
                "kind": item["kind"],
                "origin": item["origin"],
                "destination": item["destination"],
                "created_at": "2026-01-01T00:00:00Z",
            }
        if (
            method == "POST"
            and path.startswith("/api/v1/cases/")
            and path.endswith("/route-slots")
        ):
            case_id = path.split("/")[4]
            slot_id = self.next_uuid()
            self.route_slots[slot_id] = {
                "owner": username,
                "case_id": case_id,
                **payload,
            }
            self.approvals[slot_id] = set()
            return {
                "route_slot_id": slot_id,
                "case_id": case_id,
                "carrier_username": payload["carrier_username"],
                "inspector_username": payload["inspector_username"],
                "created_at": "2026-01-01T00:00:00Z",
            }
        if (
            method == "POST"
            and path.startswith("/api/v1/route-slots/")
            and path.endswith("/approvals")
        ):
            slot_id = path.split("/")[4]
            self.approvals[slot_id].add((username, payload["role"]))
            return {
                "carrier_approved": any(
                    role == "carrier" for _, role in self.approvals[slot_id]
                ),
                "inspector_approved": any(
                    role == "inspector" for _, role in self.approvals[slot_id]
                ),
                "ready": len(self.approvals[slot_id]) >= 2,
            }
        if (
            method == "GET"
            and path.startswith("/api/v1/route-slots/")
            and path.endswith("/code")
        ):
            slot_id = path.split("/")[4]
            item = self.route_slots[slot_id]
            if username != item["owner"] and len(self.approvals[slot_id]) < 2:
                raise Failure(
                    Status.CORRUPT, "Stored flag is unavailable", "slot code is absent"
                )
            return {"slot_code": self.secret(item["slot_code"])}

        if method == "POST" and path == "/api/v1/manifests":
            number, revision = self.number("MNF"), self.next_uuid()
            self.manifests[number] = {**payload, "revision_id": revision}
            return {
                "manifest_number": number,
                "revision_id": revision,
                "status": "received",
            }
        if method == "GET" and "/revisions/" in path:
            _, number, _, revision = path.split("/")[-4:]
            item = self.manifests[number]
            if item["revision_id"] != revision:
                raise Failure(
                    Status.CORRUPT, "Stored flag is unavailable", "revision is absent"
                )
            return {
                "manifest_number": number,
                "revision_id": revision,
                "status": "received",
                "sealed_route": self.secret(item["sealed_route"]),
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "POST" and path == "/api/v1/certificates":
            number = self.number("CRT")
            self.certificates[number] = payload["origin_phrase"]
            return {"serial": number, "created_at": "2026-01-01T00:00:00Z"}
        if method == "GET" and path.startswith("/api/v1/certificates/"):
            number = path.rsplit("/", 1)[1]
            return {
                "serial": number,
                "origin_phrase": self.secret(self.certificates[number]),
            }
        if method == "POST" and path == "/api/v1/invoices":
            number = self.number("INV")
            self.invoices[number] = payload["payment_ref"]
            return {
                "invoice_number": number,
                "amount": payload["amount"],
                "status": "pending",
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "GET" and path.startswith("/api/v1/invoices/"):
            number = path.rsplit("/", 1)[1]
            return {
                "invoice_number": number,
                "amount": 1,
                "status": "pending",
                "payment_ref": self.secret(self.invoices[number]),
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "POST" and path == "/api/v1/inspections":
            number = self.number("INSP")
            self.inspections[number] = {
                "owner": username,
                "inspector": payload["inspector_username"],
                "note": payload["inspector_note"],
            }
            return {"inspection_number": number, "created_at": "2026-01-01T00:00:00Z"}
        if (
            method == "GET"
            and path.startswith("/api/v1/inspections/")
            and path.endswith("/note")
        ):
            number = path.split("/")[-2]
            item = self.inspections[number]
            if username not in {item["owner"], item["inspector"]}:
                raise Failure(
                    Status.CORRUPT, "Stored flag is unavailable", "inspection is absent"
                )
            return {"inspector_note": self.secret(item["note"])}
        if method == "POST" and path == "/api/v1/licenses":
            number = self.number("LIC")
            self.licenses[number] = payload["release_pin"]
            return {
                "license_number": number,
                "status": "active",
                "created_at": "2026-01-01T00:00:00Z",
            }
        if (
            method == "POST"
            and path.startswith("/api/v1/licenses/")
            and path.endswith("/tokens")
        ):
            number = path.split("/")[-2]
            token = f"token-{self.next_uuid()}"
            self.tokens[token] = number
            return {"token": token, "expires_at": "2027-01-01T00:00:00Z"}
        if (
            method == "GET"
            and path.startswith("/api/v1/licenses/")
            and path.endswith("/release-pin")
        ):
            number = path.split("/")[-2]
            if self.tokens.get(bearer) != number:
                raise Failure(
                    Status.CORRUPT, "Stored flag is unavailable", "license is absent"
                )
            return {
                "license_number": number,
                "release_pin": self.secret(self.licenses[number]),
            }
        if method == "POST" and path == "/api/v1/contracts":
            number = self.number("CTR")
            self.contracts[number] = payload
            return {"contract_number": number, "created_at": "2026-01-01T00:00:00Z"}
        if method == "GET" and path.startswith("/api/v1/contracts/"):
            number = path.rsplit("/", 1)[1]
            return {
                "contract_number": number,
                **self.contracts[number],
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "POST" and path == "/api/v1/quotas":
            number = self.number("QUO")
            self.quotas[number] = payload
            return {"quota_number": number, "created_at": "2026-01-01T00:00:00Z"}
        if method == "GET" and path.startswith("/api/v1/quotas/"):
            number = path.rsplit("/", 1)[1]
            return {
                "quota_number": number,
                **self.quotas[number],
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "POST" and path == "/api/v1/archive-waybills":
            number = self.number("AWB")
            self.waybills[number] = payload["archive_code"]
            return {
                "waybill_number": number,
                "state": "ready",
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "GET" and path.startswith("/api/v1/archive-waybills/"):
            number = path.split("/")[-1]
            return {
                "waybill_number": number,
                "archive_code": self.secret(self.waybills[number]),
                "state": "ready",
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "POST" and path == "/api/v1/gate-passes":
            number = self.number("GATE")
            self.gate_passes[number] = {
                "owner": username,
                "operator": payload["gate_operator_username"],
                "gate_word": payload["gate_word"],
                "state": payload["state"],
            }
            return {
                "gate_pass_number": number,
                "state": payload["state"],
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "GET" and path.startswith("/api/v1/gate-passes/"):
            number = path.split("/")[-1]
            item = self.gate_passes[number]
            if username not in {item["owner"], item["operator"]}:
                raise Failure(
                    Status.CORRUPT, "Stored flag is unavailable", "gate pass is absent"
                )
            return {
                "gate_pass_number": number,
                "state": item["state"],
                "gate_word": self.secret(item["gate_word"]),
                "created_at": "2026-01-01T00:00:00Z",
            }
        if method == "POST" and path == "/api/v1/previews":
            preview_id = self.next_uuid()
            self.previews[preview_id] = "created"
            return {"preview_id": preview_id, "state": "created"}
        if method == "PUT" and path.startswith("/api/v1/previews/"):
            preview_id = path.rsplit("/", 1)[1]
            self.previews[preview_id] = "complete"
            return {"preview_id": preview_id, "state": "queued"}
        if method == "GET" and path.startswith("/api/v1/previews/"):
            preview_id = path.rsplit("/", 1)[1]
            result = {"preview_id": preview_id, "state": self.previews[preview_id]}
            if result["state"] == "complete":
                result["png_base64"] = (
                    "not-base64"
                    if self.bad_preview
                    else base64.b64encode(PNG).decode("ascii")
                )
            return result
        raise AssertionError(f"unexpected request: {method} {path}")

    def secret(self, value):
        if self.wrong_secret:
            return "wrong-value"
        return value


class CheckerTests(unittest.TestCase):
    def test_checker_is_an_executable_python_entrypoint(self):
        path = Path(checker.__file__)

        self.assertEqual(path.read_text().splitlines()[0], "#!/usr/bin/env python3")
        self.assertTrue(os.access(path, os.X_OK))

    def test_all_flag_places_put_then_get_the_exact_field(self):
        api = FakeClient()

        for place, flag in FLAGS.items():
            with self.subTest(place=place):
                state = checker.run_put(api, "a" * 40, flag, place)
                self.assertIsInstance(state, State)
                self.assertEqual(state.place, place)
                self.assertRegex(state.username, r"^[a-z0-9_.-]{3,32}$")
                self.assertGreaterEqual(len(state.password), 10)
                self.assertNotIn(flag, state.encode())
                self.assertNotIn("session", state.encode())
                self.assertNotIn("token", state.encode())
                self.assertNotIn("carrier", state.encode())
                self.assertNotIn("inspector", state.encode())
                self.assertEqual(
                    checker.run_get(api, state.encode(), flag, place),
                    state.object_number,
                )

        manifest = next(
            call for call in api.calls if call["path"] == "/api/v1/manifests"
        )
        self.assertEqual(manifest["expected"], 201)
        self.assertIsNotNone(manifest["cookie"])
        self.assertTrue(any(call.get("stored") is True for call in api.calls))

    def test_manifest_private_state_contains_revision_id(self):
        api = FakeClient()
        state = checker.run_put(api, "a" * 40, FLAGS[1], 1)

        self.assertRegex(state.revision_id, r"^[0-9a-f-]{36}$")
        self.assertEqual(State.decode(state.encode(), 1).revision_id, state.revision_id)

    def test_get_marks_flag_mismatch_as_corrupt_without_echoing_values(self):
        api = FakeClient()
        state = checker.run_put(api, "a" * 40, FLAGS[3], 3)
        api.wrong_secret = True

        with self.assertRaises(Failure) as raised:
            checker.run_get(api, state.encode(), FLAGS[3], 3)

        self.assertEqual(raised.exception.status, Status.CORRUPT)
        self.assertNotIn(FLAGS[3], raised.exception.public + raised.exception.private)
        self.assertNotIn(
            "wrong-value", raised.exception.public + raised.exception.private
        )

    def test_get_marks_missing_stored_object_as_corrupt(self):
        api = FakeClient()
        state = checker.run_put(api, "a" * 40, FLAGS[2], 2)
        api.unavailable_stored = True

        with self.assertRaises(Failure) as raised:
            checker.run_get(api, state.encode(), FLAGS[2], 2)

        self.assertEqual(raised.exception.status, Status.CORRUPT)

    def test_get_rejects_untrusted_state_identifiers_before_network_use(self):
        api = FakeClient()
        invalid = State(2, "owner", "password-long", "/api/v1/cases/other", None)

        with self.assertRaises(Failure) as raised:
            checker.run_get(api, invalid.encode(), FLAGS[2], 2)

        self.assertEqual(raised.exception.status, Status.CHECK_FAILED)
        self.assertEqual(api.calls, [])

    def test_check_uses_seed_roles_and_exercises_approval_and_eps_flows(self):
        api = FakeClient()

        checker.run_check(api)

        self.assertTrue(
            any(
                call["method"] == "GET" and call["path"] == "/healthz"
                for call in api.calls
            )
        )
        approvals = [call for call in api.calls if call["path"].endswith("/approvals")]
        self.assertEqual(
            [(call["username"], call["payload"]["role"]) for call in approvals],
            [
                ("carrier", "carrier"),
                ("inspector", "inspector"),
            ],
        )
        note_reads = [call for call in api.calls if call["path"].endswith("/note")]
        gate_reads = [
            call
            for call in api.calls
            if call["path"].startswith("/api/v1/gate-passes/")
        ]
        self.assertTrue(note_reads)
        self.assertTrue(gate_reads)
        self.assertEqual(note_reads[-1]["username"], "inspector")
        self.assertEqual(gate_reads[-1]["username"], "inspector")
        self.assertTrue(
            any(
                call["method"] == "PUT" and call["path"].startswith("/api/v1/previews/")
                for call in api.calls
            )
        )
        self.assertTrue(
            any(
                call["path"].startswith("/api/v1/previews/") and call["method"] == "GET"
                for call in api.calls
            )
        )
        self.assertEqual(
            {"carrier", "inspector"} & set(api.users), {"carrier", "inspector"}
        )
        self.assertFalse(
            any(
                call["path"] == "/api/v1/register"
                and call.get("username") in {"carrier", "inspector"}
                for call in api.calls
            )
        )

    def test_check_issues_fresh_license_token_for_release_pin(self):
        api = FakeClient()

        checker.run_check(api)

        token_call = next(
            call for call in api.calls if call["path"].endswith("/tokens")
        )
        pin_call = next(
            call for call in api.calls if call["path"].endswith("/release-pin")
        )
        self.assertEqual(token_call["payload"], {"scope": "license:read"})
        self.assertIn(pin_call["bearer"], api.tokens)
        self.assertEqual(
            api.tokens[pin_call["bearer"]], pin_call["path"].split("/")[-2]
        )
        self.assertIsNone(pin_call["cookie"])

    def test_check_rejects_invalid_preview_image(self):
        api = FakeClient()
        api.bad_preview = True

        with self.assertRaises(Failure) as raised:
            checker.run_check(api)

        self.assertEqual(raised.exception.status, Status.MUMBLE)

    def test_cli_uses_separate_public_and_private_put_channels(self):
        api = FakeClient()
        stdout, stderr = io.StringIO(), io.StringIO()
        args = ["put", "service", "a" * 40, FLAGS[1], "1"]

        with (
            patch.object(checker, "Client", return_value=api),
            contextlib.redirect_stdout(stdout),
            contextlib.redirect_stderr(stderr),
        ):
            result = checker.main(args)

        state = json.loads(stderr.getvalue())
        self.assertEqual(result, Status.UP)
        self.assertRegex(stdout.getvalue().strip(), r"^MNF-[0-9a-f-]{36}$")
        self.assertEqual(state["place"], 1)
        self.assertNotIn(FLAGS[1], stderr.getvalue())

    def test_cli_get_and_check_print_ok(self):
        api = FakeClient()
        state = checker.run_put(api, "a" * 40, FLAGS[2], 2)

        for args in (
            ["get", "service", state.encode(), FLAGS[2], "2"],
            ["check", "service"],
        ):
            stdout, stderr = io.StringIO(), io.StringIO()
            with (
                patch.object(checker, "Client", return_value=api),
                contextlib.redirect_stdout(stdout),
                contextlib.redirect_stderr(stderr),
            ):
                result = checker.main(args)
            self.assertEqual(result, Status.UP)
            self.assertEqual(stdout.getvalue(), "OK\n")
            self.assertEqual(stderr.getvalue(), "")

    def test_cli_maps_protocol_verdicts_and_hides_unexpected_values(self):
        for status in (Status.CORRUPT, Status.MUMBLE, Status.DOWN, Status.CHECK_FAILED):
            stdout, stderr = io.StringIO(), io.StringIO()
            failure = Failure(status, "safe public", "safe private")
            with (
                patch.object(checker, "Client", side_effect=failure),
                contextlib.redirect_stdout(stdout),
                contextlib.redirect_stderr(stderr),
            ):
                result = checker.main(["check", "service"])
            self.assertEqual(result, status)
            self.assertEqual(stdout.getvalue(), "safe public\n")
            self.assertEqual(stderr.getvalue(), "safe private\n")

        stdout, stderr = io.StringIO(), io.StringIO()
        with (
            patch.object(checker, "Client", side_effect=RuntimeError("secret-value")),
            contextlib.redirect_stdout(stdout),
            contextlib.redirect_stderr(stderr),
        ):
            result = checker.main(["check", "service"])
        self.assertEqual(result, Status.CHECK_FAILED)
        self.assertNotIn("secret-value", stdout.getvalue() + stderr.getvalue())
        self.assertNotIn("Traceback", stdout.getvalue() + stderr.getvalue())

    def test_public_hint_is_enabled(self):
        self.assertIs(checker.PUBLIC_HINT, True)


if __name__ == "__main__":
    unittest.main()
