"""Offline tests must never inherit developer credentials or paid configuration."""

import os

os.environ["AI_PROVIDER"] = "fake"
for key in ("AI_API_KEY", "GEMINI_API_KEY", "GOOGLE_API_KEY", "OPENAI_API_KEY", "ANTHROPIC_API_KEY"):
    os.environ[key] = ""
os.environ.update({"AI_MODEL": "gemini-2.5-flash", "AI_TIMEOUT": "20", "AI_MAX_TOKENS": "2048",
                   "AI_MAX_INPUT_TOKENS": "16000", "AI_TEMPERATURE": "0.7", "AI_CONFIDENCE_THRESHOLD": "0.8"})
# Only a caller-exported test URL can opt into database tests, never .env defaults.
os.environ["AI_TEST_DATABASE_URL"] = os.environ.get("AI_TEST_DATABASE_URL", "")
