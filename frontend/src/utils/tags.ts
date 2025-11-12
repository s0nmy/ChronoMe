export function parseTagInput(input: string): string[] {
  return input
    .split(/[,、\s]+/)
    .map((tag) => tag.trim())
    .filter(Boolean);
}

export function formatTags(tags: string[] = []): string {
  return tags.join(", ");
}
