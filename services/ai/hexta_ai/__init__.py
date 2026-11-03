"""Hexta AI service."""

import os
import sys
from pathlib import Path

# Automatically ensure generated protobuf bindings are in python path
_service_root = Path(__file__).resolve().parent.parent
_generated_dir = _service_root / "generated"
if _generated_dir.is_dir() and str(_generated_dir) not in sys.path:
    sys.path.insert(0, str(_generated_dir))


def load_env(env_path: Path | str | None = None) -> None:
    path = Path(env_path) if env_path else _service_root / ".env"
    if not path.is_file():
        return
    with open(path, encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, val = line.split("=", 1)
            key = key.strip()
            val = val.strip().strip("'\"")
            os.environ.setdefault(key, val)


load_env()
