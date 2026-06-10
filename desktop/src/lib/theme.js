export const THEME_STORAGE_KEY = "mmtunnel.theme";

export function resolveThemeMode(value) {
  return ["system", "light", "dark"].includes(value) ? value : "system";
}

export function themeClassForMode(mode, systemDark) {
  const resolved = resolveThemeMode(mode);
  if (resolved === "dark") return "dark";
  if (resolved === "system" && systemDark) return "dark";
  return "";
}

export function getStoredTheme(storage = globalThis.localStorage) {
  try {
    return resolveThemeMode(storage?.getItem(THEME_STORAGE_KEY));
  } catch {
    return "system";
  }
}

export function setStoredTheme(mode, storage = globalThis.localStorage) {
  const resolved = resolveThemeMode(mode);
  storage?.setItem(THEME_STORAGE_KEY, resolved);
  return resolved;
}

export function applyTheme(mode, systemDark, root = globalThis.document?.documentElement) {
  const className = themeClassForMode(mode, systemDark);
  root?.classList.toggle("dark", className === "dark");
  return className;
}
