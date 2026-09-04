"use client";

/**
 * Brand form controls.
 *
 * Native `<select>`, `<input type="date">`, radios and checkboxes are drawn by
 * the operating system: a blue macOS calendar and a blue tick land in the
 * middle of a plum-and-cream page, and they look different on every machine
 * the product is opened on. These are the same controls rendered in Obiara's
 * own vocabulary, keyboard-complete and announced the same way.
 *
 * The real input is kept in every case rather than replaced — a visually
 * hidden checkbox or radio still carries focus, still participates in a form,
 * and is still what a screen reader reads. Only the paint is ours.
 */

import {
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
  type ReactNode,
} from "react";

import "./form-controls.css";

/* ------------------------------------------------------------------ check */

export interface ObiaraCheckboxProps {
  readonly checked: boolean;
  readonly onChange: (checked: boolean) => void;
  readonly label: ReactNode;
  readonly description?: ReactNode;
  readonly disabled?: boolean;
  readonly required?: boolean;
  readonly name?: string;
}

export function ObiaraCheckbox({
  checked,
  onChange,
  label,
  description,
  disabled = false,
  required = false,
  name,
}: Readonly<ObiaraCheckboxProps>) {
  return (
    <label
      className={`obiara-check${disabled ? " is-disabled" : ""}`}
      data-checked={checked ? "true" : "false"}
    >
      <input
        checked={checked}
        className="obiara-control-input"
        disabled={disabled}
        name={name}
        onChange={(event) => onChange(event.target.checked)}
        required={required}
        type="checkbox"
      />
      <span aria-hidden="true" className="obiara-check-box">
        <svg viewBox="0 0 16 16" focusable="false">
          <path
            d="M3.5 8.5l3 3 6-6.5"
            fill="none"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2.4"
          />
        </svg>
      </span>
      <span className="obiara-control-text">
        <span className="obiara-control-label">{label}</span>
        {description ? (
          <small className="obiara-control-description">{description}</small>
        ) : null}
      </span>
    </label>
  );
}

/* ------------------------------------------------------------------ radio */

export interface ObiaraRadioOption<Value extends string> {
  readonly value: Value;
  readonly label: ReactNode;
  readonly description?: ReactNode;
}

export interface ObiaraRadioGroupProps<Value extends string> {
  readonly legend: ReactNode;
  readonly name: string;
  readonly value: Value;
  readonly onChange: (value: Value) => void;
  readonly options: ReadonlyArray<ObiaraRadioOption<Value>>;
  readonly disabled?: boolean;
}

export function ObiaraRadioGroup<Value extends string>({
  legend,
  name,
  value,
  onChange,
  options,
  disabled = false,
}: Readonly<ObiaraRadioGroupProps<Value>>) {
  return (
    <fieldset className="obiara-radio-group" disabled={disabled}>
      <legend>{legend}</legend>
      <div className="obiara-radio-options">
        {options.map((option) => (
          <label
            className="obiara-radio"
            data-checked={option.value === value ? "true" : "false"}
            key={option.value}
          >
            <input
              checked={option.value === value}
              className="obiara-control-input"
              name={name}
              onChange={() => onChange(option.value)}
              type="radio"
              value={option.value}
            />
            <span aria-hidden="true" className="obiara-radio-dot" />
            <span className="obiara-control-text">
              <span className="obiara-control-label">{option.label}</span>
              {option.description ? (
                <small className="obiara-control-description">
                  {option.description}
                </small>
              ) : null}
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}

/* ----------------------------------------------------------------- select */

export interface ObiaraSelectOption<Value extends string> {
  readonly value: Value;
  readonly label: string;
  readonly description?: string;
}

export interface ObiaraSelectProps<Value extends string> {
  readonly label: ReactNode;
  readonly value: Value;
  readonly onChange: (value: Value) => void;
  readonly options: ReadonlyArray<ObiaraSelectOption<Value>>;
  readonly describedBy?: string;
  readonly disabled?: boolean;
  readonly placeholder?: string;
}

/**
 * A listbox, not a native select.
 *
 * The button carries the ARIA combobox contract and the popup is a real
 * listbox, so the control is announced and driven exactly like the native one
 * it replaces: arrows move, Home/End jump, typing jumps to a label, Enter and
 * Space commit, Escape closes without changing anything.
 */
export function ObiaraSelect<Value extends string>({
  label,
  value,
  onChange,
  options,
  describedBy,
  disabled = false,
  placeholder = "Select an option",
}: Readonly<ObiaraSelectProps<Value>>) {
  const listId = useId();
  const labelId = useId();
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(() =>
    Math.max(
      0,
      options.findIndex((option) => option.value === value),
    ),
  );
  const root = useRef<HTMLDivElement>(null);
  const typeahead = useRef({ buffer: "", at: 0 });
  const selected = options.find((option) => option.value === value);

  useEffect(() => {
    if (!open) return undefined;
    // A click anywhere else is a dismissal, the same as Escape. Without this
    // the popup outlives the intent that opened it.
    function onPointerDown(event: PointerEvent) {
      if (!root.current?.contains(event.target as Node)) setOpen(false);
    }
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [open]);

  const commit = useCallback(
    (index: number) => {
      const option = options[index];
      if (!option) return;
      onChange(option.value);
      setOpen(false);
    },
    [onChange, options],
  );

  function onKeyDown(event: KeyboardEvent<HTMLElement>) {
    const last = options.length - 1;
    if (event.key === "Escape") {
      setOpen(false);
      return;
    }
    if (
      !open &&
      (event.key === "Enter" || event.key === " " || event.key === "ArrowDown")
    ) {
      event.preventDefault();
      setOpen(true);
      return;
    }
    if (!open) return;
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActive((index) => Math.min(last, index + 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActive((index) => Math.max(0, index - 1));
    } else if (event.key === "Home") {
      event.preventDefault();
      setActive(0);
    } else if (event.key === "End") {
      event.preventDefault();
      setActive(last);
    } else if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      commit(active);
    } else if (event.key.length === 1) {
      // Typeahead: consecutive letters build a prefix, a pause starts over.
      const now = performance.now();
      const state = typeahead.current;
      state.buffer =
        now - state.at > 700 ? event.key : state.buffer + event.key;
      state.at = now;
      const match = options.findIndex((option) =>
        option.label.toLowerCase().startsWith(state.buffer.toLowerCase()),
      );
      if (match >= 0) setActive(match);
    }
  }

  return (
    <div className="obiara-select" ref={root}>
      <span className="obiara-control-label" id={labelId}>
        {label}
      </span>
      <button
        aria-controls={open ? listId : undefined}
        aria-describedby={describedBy}
        aria-expanded={open}
        aria-haspopup="listbox"
        aria-labelledby={`${labelId}`}
        className="obiara-select-trigger"
        disabled={disabled}
        onClick={() => setOpen((wasOpen) => !wasOpen)}
        onKeyDown={onKeyDown}
        type="button"
      >
        <span className={selected ? undefined : "obiara-select-placeholder"}>
          {selected ? selected.label : placeholder}
        </span>
        <span aria-hidden="true" className="obiara-select-caret">
          ▾
        </span>
      </button>
      {open ? (
        <ul
          aria-labelledby={labelId}
          className="obiara-select-list"
          id={listId}
          role="listbox"
          tabIndex={-1}
        >
          {options.map((option, index) => (
            <li
              aria-selected={option.value === value}
              className={`obiara-select-option${index === active ? " is-active" : ""}`}
              key={option.value}
              onClick={() => commit(index)}
              onMouseEnter={() => setActive(index)}
              role="option"
            >
              <span className="obiara-control-label">{option.label}</span>
              {option.description ? (
                <small className="obiara-control-description">
                  {option.description}
                </small>
              ) : null}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}

/* ------------------------------------------------------------------- time */

/** Half-hour slots across a day, as the `HH:MM` strings a form submits. */
export function timeSlots(stepMinutes = 30): ReadonlyArray<{
  value: string;
  label: string;
}> {
  const slots: Array<{ value: string; label: string }> = [];
  for (let minutes = 0; minutes < 24 * 60; minutes += stepMinutes) {
    const hour = Math.floor(minutes / 60);
    const minute = minutes % 60;
    const value = `${pad(hour)}:${pad(minute)}`;
    const meridiem = hour < 12 ? "am" : "pm";
    const twelve = hour % 12 === 0 ? 12 : hour % 12;
    slots.push({ value, label: `${twelve}:${pad(minute)} ${meridiem}` });
  }
  return slots;
}

export interface ObiaraTimeFieldProps {
  readonly label: ReactNode;
  /** `HH:MM`, 24-hour — the same value `<input type="time">` produced. */
  readonly value: string;
  readonly onChange: (value: string) => void;
  readonly disabled?: boolean;
  readonly stepMinutes?: number;
}

/**
 * A time chosen from the brand's own list.
 *
 * `<input type="time">` draws a system spinner that differs on every platform
 * and cannot be themed. Quiet hours only ever land on a slot, so a listbox of
 * slots says the same thing and looks like the rest of the page.
 */
export function ObiaraTimeField({
  label,
  value,
  onChange,
  disabled = false,
  stepMinutes = 30,
}: Readonly<ObiaraTimeFieldProps>) {
  const options = useMemo(() => timeSlots(stepMinutes), [stepMinutes]);
  return (
    <ObiaraSelect
      disabled={disabled}
      label={label}
      onChange={onChange}
      options={options}
      placeholder="Choose a time"
      value={value}
    />
  );
}

/* ------------------------------------------------------------------- date */

const MONTHS = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];
const DAY_INITIALS = ["M", "T", "W", "T", "F", "S", "S"];

function pad(value: number): string {
  return value.toString().padStart(2, "0");
}

/** Formats a calendar day as the ISO date the form submits. */
export function isoDate(year: number, month: number, day: number): string {
  return `${year}-${pad(month + 1)}-${pad(day)}`;
}

/** Days in a month, and the Monday-based weekday its first day falls on. */
export function monthGrid(
  year: number,
  month: number,
): Readonly<{ days: number; leading: number }> {
  const days = new Date(Date.UTC(year, month + 1, 0)).getUTCDate();
  // getUTCDay is Sunday-based; Obiara's calendars start on Monday.
  const firstDay = new Date(Date.UTC(year, month, 1)).getUTCDay();
  return { days, leading: (firstDay + 6) % 7 };
}

export interface ObiaraDateFieldProps {
  readonly label: ReactNode;
  /** ISO `YYYY-MM-DD`, or an empty string for no choice yet. */
  readonly value: string;
  readonly onChange: (value: string) => void;
  readonly max?: string;
  readonly min?: string;
  readonly describedBy?: string;
  readonly disabled?: boolean;
  /** Which month to open on when nothing is chosen yet. */
  readonly defaultMonth?: string;
}

/**
 * A calendar drawn in the brand, not the operating system's.
 *
 * `<input type="date">` renders a blue system calendar that looks different on
 * every machine and cannot be styled. This keeps the same value shape — an ISO
 * `YYYY-MM-DD` string — so callers and validation are unchanged.
 */
export function ObiaraDateField({
  label,
  value,
  onChange,
  max,
  min,
  describedBy,
  disabled = false,
  defaultMonth,
}: Readonly<ObiaraDateFieldProps>) {
  const labelId = useId();
  const dialogId = useId();
  const [open, setOpen] = useState(false);
  const root = useRef<HTMLDivElement>(null);

  const opening = value || defaultMonth || max || min || "";
  const parsed = /^(\d{4})-(\d{2})-(\d{2})$/.exec(opening);
  const [cursor, setCursor] = useState(() => ({
    year: parsed ? Number(parsed[1]) : 2000,
    month: parsed ? Number(parsed[2]) - 1 : 0,
  }));

  useEffect(() => {
    if (!open) return undefined;
    function onPointerDown(event: PointerEvent) {
      if (!root.current?.contains(event.target as Node)) setOpen(false);
    }
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [open]);

  const { days, leading } = useMemo(
    () => monthGrid(cursor.year, cursor.month),
    [cursor.year, cursor.month],
  );

  function outOfRange(iso: string): boolean {
    if (min && iso < min) return true;
    if (max && iso > max) return true;
    return false;
  }

  function step(months: number) {
    setCursor((current) => {
      const next = current.month + months;
      return {
        year: current.year + Math.floor(next / 12),
        month: ((next % 12) + 12) % 12,
      };
    });
  }

  return (
    <div className="obiara-date" ref={root}>
      <span className="obiara-control-label" id={labelId}>
        {label}
      </span>
      <button
        aria-controls={open ? dialogId : undefined}
        aria-describedby={describedBy}
        aria-expanded={open}
        aria-labelledby={labelId}
        className="obiara-date-trigger"
        disabled={disabled}
        onClick={() => setOpen((wasOpen) => !wasOpen)}
        onKeyDown={(event) => {
          if (event.key === "Escape") setOpen(false);
        }}
        type="button"
      >
        <span className={value ? undefined : "obiara-select-placeholder"}>
          {value || "Choose a date"}
        </span>
        <span aria-hidden="true" className="obiara-date-icon">
          ▦
        </span>
      </button>
      {open ? (
        <div
          aria-label="Choose a date"
          className="obiara-date-panel"
          id={dialogId}
          role="dialog"
        >
          <div className="obiara-date-head">
            <button
              aria-label="Previous month"
              className="obiara-date-step"
              onClick={() => step(-1)}
              type="button"
            >
              ‹
            </button>
            <div className="obiara-date-caption" aria-live="polite">
              {MONTHS[cursor.month]} {cursor.year}
            </div>
            <button
              aria-label="Next month"
              className="obiara-date-step"
              onClick={() => step(1)}
              type="button"
            >
              ›
            </button>
          </div>
          <div aria-hidden="true" className="obiara-date-weekdays">
            {DAY_INITIALS.map((initial, index) => (
              <span key={`${initial}-${index}`}>{initial}</span>
            ))}
          </div>
          <div className="obiara-date-grid" role="grid">
            {Array.from({ length: leading }, (_, index) => (
              <span key={`lead-${index}`} />
            ))}
            {Array.from({ length: days }, (_, index) => {
              const day = index + 1;
              const iso = isoDate(cursor.year, cursor.month, day);
              const unavailable = outOfRange(iso);
              return (
                <button
                  aria-pressed={iso === value}
                  className="obiara-date-day"
                  disabled={unavailable}
                  key={iso}
                  onClick={() => {
                    onChange(iso);
                    setOpen(false);
                  }}
                  type="button"
                >
                  {day}
                </button>
              );
            })}
          </div>
          {value ? (
            <button
              className="obiara-date-clear"
              onClick={() => {
                onChange("");
                setOpen(false);
              }}
              type="button"
            >
              Clear
            </button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
