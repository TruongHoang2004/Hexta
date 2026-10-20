"""Atlas adoption runs only in disposable databases, never the application DB."""

import os
from pathlib import Path
import shutil
import subprocess
from urllib.parse import urlsplit, urlunsplit
from uuid import uuid4

import psycopg
from psycopg import sql
import pytest


@pytest.fixture
def atlas_database():
    dsn = os.getenv("AI_TEST_DATABASE_URL")
    if not dsn:
        pytest.skip("AI_TEST_DATABASE_URL is required")
    if not shutil.which("atlas"):
        pytest.fail("Atlas CLI is required for configured migration adoption tests")
    configured_root = os.getenv("ATLAS_PROJECT_ROOT")
    project = Path(configured_root) if configured_root else Path(__file__).resolve().parents[3]
    if not (project / "atlas.hcl").exists():
        pytest.fail("Atlas project sources must be mounted for migration adoption tests")
    name = "hexta_ai_test_" + uuid4().hex
    with psycopg.connect(dsn, autocommit=True) as admin:
        admin.execute(sql.SQL("CREATE DATABASE {}").format(sql.Identifier(name)))
        parsed = urlsplit(dsn)
        database_url = urlunsplit(parsed._replace(path="/" + name))
        try:
            yield project, database_url
        finally:
            admin.execute(sql.SQL("DROP DATABASE {} WITH (FORCE)").format(sql.Identifier(name)))


def legacy_fixture(project, database_url):
    with psycopg.connect(database_url) as db:
        db.execute((project / "migrations/schema/ai.sql").read_text())
        db.execute("""INSERT INTO ai_runtime.runs
            (tenant_id,run_id,request,digest,result,status)
            VALUES ('atlas-adoption','retained',%s,%s,%s,1)""", (b"request", b"digest", b"result"))


def apply_ai(project, database_url):
    environment = {**os.environ, "AI_DATABASE_URL": database_url}
    return subprocess.run(
        ["atlas", "migrate", "apply", "--env", "ai", "--allow-dirty"],
        cwd=project, env=environment, capture_output=True, text=True, timeout=45,
    )


def test_adopts_compatible_legacy_table_without_losing_runs(atlas_database):
    project, database_url = atlas_database
    legacy_fixture(project, database_url)
    assert apply_ai(project, database_url).returncode == 0
    assert apply_ai(project, database_url).returncode == 0
    with psycopg.connect(database_url) as db:
        row = db.execute("SELECT request,digest,result,status FROM ai_runtime.runs WHERE tenant_id='atlas-adoption'").fetchone()
        assert tuple(bytes(value) for value in row[:3]) == (b"request", b"digest", b"result")
        assert row[3] == 1
        assert db.execute("SELECT count(*) FROM atlas_ai_revisions.atlas_schema_revisions WHERE version='20261020100000' AND COALESCE(error,'')='' ").fetchone()[0] == 1


def test_rejects_incompatible_legacy_schema_without_changing_rows(atlas_database):
    project, database_url = atlas_database
    legacy_fixture(project, database_url)
    with psycopg.connect(database_url) as db:
        db.execute("ALTER TABLE ai_runtime.runs ADD COLUMN incompatible text")
    result = apply_ai(project, database_url)
    assert result.returncode != 0
    assert "incompatible" in result.stderr
    with psycopg.connect(database_url) as db:
        assert db.execute("SELECT count(*) FROM ai_runtime.runs").fetchone()[0] == 1


def test_api_and_ai_histories_share_database_without_collision(atlas_database):
    project, database_url = atlas_database
    environment = {**os.environ, "AI_DATABASE_URL": database_url,
                   "API_DATABASE_URL": database_url + ("&" if "?" in database_url else "?") + "search_path=public",
                   "MIGRATE_TARGET": "all", "MIGRATION_ROOT": str(project)}
    for _ in range(2):
        result = subprocess.run(["sh", str(project / "scripts/migrate.sh"), "apply"],
                                cwd=project, env=environment, capture_output=True, text=True, timeout=45)
        assert result.returncode == 0
    with psycopg.connect(database_url) as db:
        assert db.execute("SELECT count(*) FROM atlas_schema_revisions.atlas_schema_revisions WHERE COALESCE(error,'')=''").fetchone()[0] == 7
        assert db.execute("SELECT count(*) FROM atlas_ai_revisions.atlas_schema_revisions WHERE COALESCE(error,'')=''").fetchone()[0] == 1
