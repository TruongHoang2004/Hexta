#!/usr/bin/env python3
import json
import subprocess
import sys
import re

def slugify(text):
    text = text.lower()
    text = re.sub(r'[^a-z0-9]+', '-', text).strip('-')
    return text[:40]

def get_priority_weight(labels):
    label_names = [l.get("name", "").lower() for l in labels]
    
    # Exclude issues already in-progress or blocked
    if any(name in ["in-progress", "in progress", "wip", "blocked"] for name in label_names):
        return -1

    if any(name in ["priority:high", "high", "priority-high"] for name in label_names):
        return 1
    if any(name in ["priority:medium", "medium", "priority-medium"] for name in label_names):
        return 2
    if any(name in ["priority:low", "low", "priority-low"] for name in label_names):
        return 3
    return 4  # Default / no priority label

def main():
    cmd = ["gh", "issue", "list", "--state", "open", "--json", "number,title,body,labels,createdAt", "--limit", "100"]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
    except subprocess.CalledProcessError as e:
        print(json.dumps({"status": "error", "message": f"Failed to run gh command: {e.stderr}"}), file=sys.stderr)
        sys.exit(1)
    except FileNotFoundError:
        print(json.dumps({"status": "error", "message": "gh CLI not found on PATH."}), file=sys.stderr)
        sys.exit(1)

    issues = json.loads(result.stdout or "[]")
    if not issues:
        print(json.dumps({"status": "no_tasks", "message": "No open issues found in repository."}))
        sys.exit(0)

    # Filter and rank issues
    eligible = []
    for issue in issues:
        weight = get_priority_weight(issue.get("labels", []))
        if weight > 0:
            eligible.append((weight, issue.get("number", 0), issue))

    if not eligible:
        print(json.dumps({"status": "no_tasks", "message": "All open issues are either in-progress or blocked."}))
        sys.exit(0)

    # Sort by priority ascending (1 = high, 2 = medium, etc.), then by number ascending (oldest first)
    eligible.sort(key=lambda item: (item[0], item[1]))

    selected = eligible[0][2]
    branch_name = f"task/issue-{selected['number']}-{slugify(selected['title'])}"

    output = {
        "status": "ready",
        "issue": {
            "number": selected["number"],
            "title": selected["title"],
            "body": selected.get("body", ""),
            "labels": [l.get("name") for l in selected.get("labels", [])],
            "branch_name": branch_name
        }
    }
    print(json.dumps(output, indent=2))

if __name__ == "__main__":
    main()
