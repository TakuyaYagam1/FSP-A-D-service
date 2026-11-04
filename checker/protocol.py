"""ForcAD arguments and private retrieval state for Wahrwelt."""

import json
import re
from dataclasses import dataclass
from enum import IntEnum

PRIVATE_ID = re.compile(r"[0-9a-f]{40}")
FLAG = re.compile(r"[A-Z0-9]{31}=")
USERNAME = re.compile(r"[a-z0-9_.-]{3,32}")
UUID_TEXT = (
    r"[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-"
    r"[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
)
REVISION_ID = re.compile(UUID_TEXT)
STATE_VERSION = 1
MAX_STATE_BYTES = 4096
PLACE_FIELDS = {
    1: "sealed_route",
    2: "origin_phrase",
    3: "payment_ref",
    4: "inspector_note",
    5: "release_pin",
    6: "crew_phrase",
    7: "slot_code",
    8: "quota_code",
    9: "archive_code",
    10: "gate_word",
}
PLACE_ARGS = frozenset(str(place) for place in PLACE_FIELDS)
OBJECT_NUMBER = {
    1: re.compile(rf"MNF-{UUID_TEXT}"),
    2: re.compile(rf"CRT-{UUID_TEXT}"),
    3: re.compile(rf"INV-{UUID_TEXT}"),
    4: re.compile(rf"INSP-{UUID_TEXT}"),
    5: re.compile(rf"LIC-{UUID_TEXT}"),
    6: re.compile(rf"CTR-{UUID_TEXT}"),
    7: re.compile(UUID_TEXT),
    8: re.compile(rf"QUO-{UUID_TEXT}"),
    9: re.compile(rf"AWB-{UUID_TEXT}"),
    10: re.compile(rf"GATE-{UUID_TEXT}"),
}
STATE_REQUIRED_FIELDS = frozenset(
    {
        "version",
        "place",
        "username",
        "password",
        "object_number",
    }
)
STATE_OPTIONAL_FIELDS = frozenset({"revision_id"})


class Status(IntEnum):
    UP = 101
    CORRUPT = 102
    MUMBLE = 103
    DOWN = 104
    CHECK_FAILED = 110


class Failure(Exception):
    def __init__(self, status: Status, public: str, private: str):
        super().__init__(private)
        self.status = status
        self.public = public
        self.private = private


@dataclass(frozen=True)
class Arguments:
    action: str
    host: str
    private: str = ""
    flag: str = ""
    place: int = 1

    @classmethod
    def parse(cls, args: list[str]) -> "Arguments":
        if len(args) == 2 and args[0] == "check" and args[1]:
            return cls("check", args[1])
        if len(args) != 5 or args[0] not in {"put", "get"} or not args[1]:
            raise Failure(
                Status.CHECK_FAILED,
                "Checker invocation failed",
                "Arguments.parse - expected check HOST or "
                "put/get HOST PRIVATE FLAG PLACE",
            )
        action, host, private, flag, raw_place = args
        if raw_place not in PLACE_ARGS or not FLAG.fullmatch(flag):
            raise Failure(
                Status.CHECK_FAILED,
                "Checker invocation failed",
                "Arguments.parse - invalid flag format or unsupported place",
            )
        if action == "put" and not PRIVATE_ID.fullmatch(private):
            raise Failure(
                Status.CHECK_FAILED,
                "Checker invocation failed",
                "Arguments.parse - PUT private ID must contain "
                "40 lowercase hex characters",
            )
        return cls(action, host, private, flag, int(raw_place))


@dataclass(frozen=True)
class State:
    place: int
    username: str
    password: str
    object_number: str
    revision_id: str | None = None

    def encode(self) -> str:
        try:
            _validate_state(
                self.place,
                self.username,
                self.password,
                self.object_number,
                self.revision_id,
            )
            data = {
                "version": STATE_VERSION,
                "place": self.place,
                "username": self.username,
                "password": self.password,
                "object_number": self.object_number,
            }
            if self.revision_id is not None:
                data["revision_id"] = self.revision_id
            raw = json.dumps(data, separators=(",", ":"))
            if len(raw.encode("utf-8")) > MAX_STATE_BYTES:
                raise ValueError("state too large")
        except (TypeError, ValueError, UnicodeError) as exc:
            raise Failure(
                Status.CHECK_FAILED,
                "Checker state is invalid",
                "State.encode - invalid or oversized retrieval data",
            ) from exc
        return raw

    @classmethod
    def decode(cls, raw: str, place: int) -> "State":
        try:
            if not isinstance(raw, str) or len(raw) > MAX_STATE_BYTES:
                raise ValueError("invalid state size")
            if type(place) is not int or place not in PLACE_FIELDS:
                raise ValueError("invalid requested place")
            if len(raw.encode("utf-8")) > MAX_STATE_BYTES:
                raise ValueError("state too large")
            data = json.loads(
                raw,
                object_pairs_hook=_unique_object,
                parse_constant=_reject_constant,
            )
            if not isinstance(data, dict):
                raise TypeError("state must be an object")
            allowed = STATE_REQUIRED_FIELDS | STATE_OPTIONAL_FIELDS
            if not STATE_REQUIRED_FIELDS.issubset(data) or not set(data).issubset(
                allowed
            ):
                raise ValueError("invalid state fields")
            if type(data["version"]) is not int or data["version"] != STATE_VERSION:
                raise ValueError("unsupported state version")
            state_place = data["place"]
            username = data["username"]
            password = data["password"]
            object_number = data["object_number"]
            revision_id = data.get("revision_id")
            if "revision_id" in data and type(revision_id) is not str:
                raise ValueError("invalid revision ID")
            _validate_state(
                state_place,
                username,
                password,
                object_number,
                revision_id,
            )
            if state_place != place:
                raise ValueError("place mismatch")
        except (TypeError, ValueError, UnicodeError, RecursionError) as exc:
            raise Failure(
                Status.CHECK_FAILED,
                "Checker state is invalid",
                "State.decode - invalid retrieval data or place mismatch",
            ) from exc
        return cls(state_place, username, password, object_number, revision_id)


def _validate_state(
    place: int,
    username: str,
    password: str,
    object_number: str,
    revision_id: str | None,
) -> None:
    if type(place) is not int or place not in PLACE_FIELDS:
        raise ValueError("invalid place")
    if type(username) is not str or not USERNAME.fullmatch(username):
        raise ValueError("invalid username")
    if (
        type(password) is not str
        or not 10 <= len(password) <= 72
        or any(ord(char) < 32 or ord(char) == 127 for char in password)
    ):
        raise ValueError("invalid password")
    if type(object_number) is not str or not OBJECT_NUMBER[place].fullmatch(
        object_number
    ):
        raise ValueError("invalid object number")
    if revision_id is not None and (
        place != 1
        or type(revision_id) is not str
        or not REVISION_ID.fullmatch(revision_id)
    ):
        raise ValueError("invalid revision ID")


def _unique_object(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate state field")
        result[key] = value
    return result


def _reject_constant(value: str) -> None:
    raise ValueError(f"invalid JSON constant: {value}")
