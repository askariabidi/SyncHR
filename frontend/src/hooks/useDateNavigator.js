import { useState } from 'react';
import { getTodayISO, buildDateStr } from '../utils/dateFormat';

// Drives a date-picker + month-nav + day-tabs UI: tracks a selected date plus
// the month currently being browsed, and won't allow navigating into the future.
export function useDateNavigator() {
  const todayISO = getTodayISO();
  const [selectedDate, setSelectedDate] = useState(todayISO);
  const [viewYear, setViewYear] = useState(new Date().getFullYear());
  const [viewMonth, setViewMonth] = useState(new Date().getMonth()); // 0-11

  const isCurrentMonthView = () => {
    const now = new Date();
    return viewYear === now.getFullYear() && viewMonth === now.getMonth();
  };

  const handlePrevMonth = () => {
    if (viewMonth === 0) {
      setViewYear((y) => y - 1);
      setViewMonth(11);
    } else {
      setViewMonth((m) => m - 1);
    }
  };

  const handleNextMonth = () => {
    if (isCurrentMonthView()) return; // no browsing into the future
    if (viewMonth === 11) {
      setViewYear((y) => y + 1);
      setViewMonth(0);
    } else {
      setViewMonth((m) => m + 1);
    }
  };

  const handleDayClick = (day) => {
    const dateStr = buildDateStr(viewYear, viewMonth, day);
    if (dateStr > todayISO) return;
    setSelectedDate(dateStr);
  };

  const handleDatePickerChange = (e) => {
    const value = e.target.value;
    if (!value) return;
    setSelectedDate(value);
    const [y, m] = value.split('-').map(Number);
    setViewYear(y);
    setViewMonth(m - 1);
  };

  const daysInViewMonth = new Date(viewYear, viewMonth + 1, 0).getDate();

  return {
    todayISO,
    selectedDate,
    viewYear,
    viewMonth,
    isCurrentMonthView,
    handlePrevMonth,
    handleNextMonth,
    handleDayClick,
    handleDatePickerChange,
    daysInViewMonth,
  };
}
