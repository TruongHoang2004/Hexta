// Ho Chi Minh City uses UTC+07:00 without daylight saving time.
export function creationRange(start: string, end: string) {
  const validDate = (value: string) => {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
    const date = new Date(`${value}T00:00:00Z`);
    return (
      Number.isFinite(date.getTime()) &&
      date.toISOString().slice(0, 10) === value
    );
  };
  if (!validDate(start) || !validDate(end) || start > end)
    throw new Error(
      "Choose a valid start and end date, with the end on or after the start.",
    );
  const nextDay = new Date(`${end}T00:00:00Z`);
  nextDay.setUTCDate(nextDay.getUTCDate() + 1);
  return {
    from: `${start}T00:00:00+07:00`,
    to: `${nextDay.toISOString().slice(0, 10)}T00:00:00+07:00`,
  };
}
export function localDate(instant: string | null, exclusiveEnd = false) {
  if (!instant) return "";
  const time = new Date(instant).getTime();
  if (!Number.isFinite(time)) return "";
  return new Date(time + 7 * 60 * 60 * 1000 - (exclusiveEnd ? 1 : 0))
    .toISOString()
    .slice(0, 10);
}
