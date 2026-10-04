import { useCallback, useState } from "react";
import { api } from "../../lib/api";
import type { Tag } from "../../types";

const TAG_COLOR_POOL = [
  "#ef4444",
  "#f97316",
  "#eab308",
  "#22c55e",
  "#3b82f6",
  "#6366f1",
  "#8b5cf6",
  "#ec4899",
  "#06b6d4",
  "#84cc16",
  "#94a3b8",
];

const hashString = (value: string): number => {
  let hash = 0;
  for (let i = 0; i < value.length; i += 1) {
    hash = (hash << 5) - hash + value.charCodeAt(i);
    hash |= 0;
  }
  return hash;
};

const pickTagColor = (name: string): string => {
  if (!name) {
    return TAG_COLOR_POOL[0];
  }
  const hash = hashString(name.toLowerCase());
  const index = Math.abs(hash) % TAG_COLOR_POOL.length;
  return TAG_COLOR_POOL[index];
};

export function useTags() {
  const [tags, setTags] = useState<Tag[]>([]);
  const ensureTagsForNames = useCallback(
    async (tagNames: string[]) => {
      const normalized = Array.from(new Set(tagNames.map((tag) => tag.trim()).filter(Boolean)));
      if (!normalized.length) {
        return [];
      }
      const existingMap = new Map(tags.map((tag) => [tag.name.toLowerCase(), tag]));
      const resolved: Tag[] = [];
      const created: Tag[] = [];
      for (const name of normalized) {
        const key = name.toLowerCase();
        const existing = existingMap.get(key);
        if (existing) {
          resolved.push(existing);
          continue;
        }
        const createdTag = await api.createTag({
          name,
          color: pickTagColor(name),
        });
        existingMap.set(key, createdTag);
        created.push(createdTag);
        resolved.push(createdTag);
      }
      if (created.length) {
        setTags((prev) => [...prev, ...created]);
      }
      return resolved;
    },
    [tags],
  );

  return { setTags, ensureTagsForNames };
}
