export function formatBytes(value: number | null | undefined): string {
  if (!value) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let size = value;
  let unit = 0;

  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit += 1;
  }

  return `${size.toLocaleString(undefined, {
    maximumFractionDigits: unit === 0 ? 0 : 1
  })} ${units[unit]}`;
}

export function formatDate(value: string): string {
  if (!value) {
    return "";
  }
  return new Date(value).toLocaleString();
}
