def extract_json(text: str) -> str:
    """Extract the outermost JSON object from LLM response text."""
    cleaned = text.strip()
    # Strip code fences
    for prefix in ("```json", "```JSON", "```"):
        if cleaned.startswith(prefix):
            cleaned = cleaned[len(prefix):]
    if cleaned.endswith("```"):
        cleaned = cleaned[:-3]
    cleaned = cleaned.strip()

    # Find outermost JSON object
    start = cleaned.find("{")
    if start == -1:
        return ""
    depth = 0
    for i in range(start, len(cleaned)):
        if cleaned[i] == "{":
            depth += 1
        elif cleaned[i] == "}":
            depth -= 1
            if depth == 0:
                return cleaned[start:i + 1]
    return ""
