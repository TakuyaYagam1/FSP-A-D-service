import json
import unittest

from checker.protocol import (
    MAX_STATE_BYTES,
    PLACE_FIELDS,
    STATE_VERSION,
    Arguments,
    Failure,
    State,
    Status,
)


class ArgumentsTests(unittest.TestCase):
    def test_parse_check(self):
        request = Arguments.parse(["check", "127.0.0.1:8080"])

        self.assertEqual(request.action, "check")
        self.assertEqual(request.host, "127.0.0.1:8080")
        self.assertEqual(request.place, 1)

    def test_parse_put_with_private_id(self):
        request = Arguments.parse(["put", "service", "a" * 40, "A" * 31 + "=", "10"])

        self.assertEqual(request.action, "put")
        self.assertEqual(request.private, "a" * 40)
        self.assertEqual(request.flag, "A" * 31 + "=")
        self.assertEqual(request.place, 10)

    def test_parse_get_keeps_private_state(self):
        private = '{"version":1,"place":2}'
        request = Arguments.parse(["get", "service", private, "Z" * 31 + "=", "2"])

        self.assertEqual(request.private, private)
        self.assertEqual(request.place, 2)

    def test_parse_rejects_bad_put_id(self):
        for identifier in ("A" * 40, "a" * 39, "g" * 40):
            with self.subTest(identifier=identifier):
                self.assert_check_failed(
                    [
                        "put",
                        "service",
                        identifier,
                        "A" * 31 + "=",
                        "1",
                    ]
                )

    def test_parse_rejects_bad_flag(self):
        for flag in ("a" * 31 + "=", "A" * 30 + "=", "A" * 32):
            with self.subTest(flag=flag):
                self.assert_check_failed(["get", "service", "private", flag, "1"])

    def test_parse_rejects_unsupported_place(self):
        for place in ("0", "11", "01", "one"):
            with self.subTest(place=place):
                self.assert_check_failed(
                    [
                        "get",
                        "service",
                        "private",
                        "A" * 31 + "=",
                        place,
                    ]
                )

    def test_parse_rejects_invalid_command_shape(self):
        invalid_commands = (
            [],
            ["check"],
            ["put", "service"],
            ["delete", "service", "x", "A" * 31 + "=", "1"],
        )
        for args in invalid_commands:
            with self.subTest(args=args):
                self.assert_check_failed(args)

    def assert_check_failed(self, args):
        with self.assertRaises(Failure) as raised:
            Arguments.parse(args)
        self.assertEqual(raised.exception.status, Status.CHECK_FAILED)


class StateTests(unittest.TestCase):
    def setUp(self):
        self.state = State(
            place=1,
            username="checker-user",
            password="checker-password",
            object_number="MNF-12345678-1234-1234-1234-123456789abc",
        )

    def test_encode_and_decode_required_fields(self):
        raw = self.state.encode()
        decoded = State.decode(raw, 1)

        self.assertEqual(decoded, self.state)
        self.assertLessEqual(len(raw.encode("utf-8")), MAX_STATE_BYTES)
        self.assertEqual(
            set(json.loads(raw)),
            {"version", "place", "username", "password", "object_number"},
        )

    def test_state_version_and_place_constants(self):
        self.assertEqual(STATE_VERSION, 1)
        self.assertEqual(tuple(PLACE_FIELDS), tuple(range(1, 11)))

    def test_revision_id_is_optional_for_place_one(self):
        state = State(
            place=1,
            username="checker-user",
            password="checker-password",
            object_number="MNF-12345678-1234-1234-1234-123456789abc",
            revision_id="11111111-2222-3333-4444-555555555555",
        )

        self.assertEqual(State.decode(state.encode(), 1), state)

    def test_revision_id_is_rejected_for_other_places(self):
        state = State(
            place=2,
            username="checker-user",
            password="checker-password",
            object_number="INV-123",
            revision_id="11111111-2222-3333-4444-555555555555",
        )

        self.assert_check_failed(lambda: state.encode())

    def test_decode_rejects_place_mismatch(self):
        self.assert_check_failed(lambda: State.decode(self.state.encode(), 2))

    def test_decode_rejects_unknown_flag_field(self):
        data = json.loads(self.state.encode())
        data["flag"] = "A" * 31 + "="

        self.assert_check_failed(lambda: State.decode(json.dumps(data), 1))

    def test_decode_rejects_unknown_version_and_extra_field(self):
        data = json.loads(self.state.encode())
        data["version"] = 2
        self.assert_check_failed(lambda: State.decode(json.dumps(data), 1))

        data = json.loads(self.state.encode())
        data["extra"] = "value"
        self.assert_check_failed(lambda: State.decode(json.dumps(data), 1))

    def test_decode_rejects_missing_and_invalid_fields(self):
        original = json.loads(self.state.encode())
        invalid_states = []

        missing = dict(original)
        del missing["password"]
        invalid_states.append(missing)

        for name, value in (
            ("version", True),
            ("place", True),
            ("username", ""),
            ("password", 123),
            ("object_number", ""),
        ):
            invalid = dict(original)
            invalid[name] = value
            invalid_states.append(invalid)

        invalid_revision = dict(original)
        invalid_revision["revision_id"] = "bad-id"
        invalid_states.append(invalid_revision)

        null_revision = dict(original)
        null_revision["revision_id"] = None
        invalid_states.append(null_revision)

        for data in invalid_states:
            with self.subTest(data=data):
                self.assert_check_failed(
                    lambda data=data: State.decode(json.dumps(data), 1)
                )

    def test_decode_rejects_duplicate_keys(self):
        raw = self.state.encode().replace('"place":1', '"place":1,"place":1')

        self.assert_check_failed(lambda: State.decode(raw, 1))

    def test_decode_rejects_invalid_json_and_oversize_state(self):
        self.assert_check_failed(lambda: State.decode("{", 1))
        self.assert_check_failed(lambda: State.decode(" " * (MAX_STATE_BYTES + 1), 1))

    def test_encode_rejects_state_over_size_limit(self):
        state = State(
            place=1,
            username="checker-user",
            password="p" * MAX_STATE_BYTES,
            object_number="MNF-12345678-1234-1234-1234-123456789abc",
        )

        self.assert_check_failed(state.encode)

    def test_state_rejects_invalid_credentials_and_object_number(self):
        invalid_states = (
            State(1, "UPPERCASE", "checker-password", self.state.object_number),
            State(1, "ok", "checker-password", self.state.object_number),
            State(1, "checker-user", "short", self.state.object_number),
            State(1, "checker-user", "checker\npassword", self.state.object_number),
            State(2, "checker-user", "checker-password", "../certificates"),
            State(7, "checker-user", "checker-password", "MNF-123"),
        )

        for state in invalid_states:
            with self.subTest(state=state):
                self.assert_check_failed(state.encode)

    def assert_check_failed(self, callback):
        with self.assertRaises(Failure) as raised:
            callback()
        self.assertEqual(raised.exception.status, Status.CHECK_FAILED)


if __name__ == "__main__":
    unittest.main()
