export const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
];

export const getTodayISO = () => {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
};

export const buildDateStr = (year, month, day) =>
  `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;

// The backend sends Go's zero-value time ("0001-01-01T00:00:00Z") instead of
// null when check-in/check-out hasn't happened yet, so treat pre-1970 dates as absent
export const isValidTimestamp = (timestamp) => {
  if (!timestamp) return false;
  return new Date(timestamp).getFullYear() > 1970;
};

export const formatDate = (dateStr) => {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  const day = date.getDate();
  const suffix =
    day % 10 === 1 && day !== 11
      ? 'st'
      : day % 10 === 2 && day !== 12
      ? 'nd'
      : day % 10 === 3 && day !== 13
      ? 'rd'
      : 'th';
  const month = date.toLocaleString('en-US', { month: 'long' });
  return `${day}${suffix} ${month} ${date.getFullYear()}`;
};

export const formatDateWithWeekday = (dateStr) => {
  if (!dateStr) return '-';
  const weekday = new Date(dateStr).toLocaleString('en-US', { weekday: 'long' });
  return `${weekday}, ${formatDate(dateStr)}`;
};

export const formatTime24 = (timestamp) => {
  if (!isValidTimestamp(timestamp)) return '-';
  return new Date(timestamp).toLocaleTimeString('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
};

export const formatDuration = (checkInStr, checkOutStr) => {
  if (!isValidTimestamp(checkInStr) || !isValidTimestamp(checkOutStr)) return '-';
  const diffMs = new Date(checkOutStr) - new Date(checkInStr);
  if (diffMs <= 0) return '-';
  const totalMinutes = Math.round(diffMs / 60000);
  if (totalMinutes < 60) return '< 1 Hour';
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return `${hours} HR ${minutes} MINS`;
};
