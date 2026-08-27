"use client";

import {
  useState,
  useEffect,
  useRef,
  type ChangeEvent,
  type ClipboardEvent,
  type KeyboardEvent,
} from "react";

import {
  clearOtpDigit,
  createOtpCells,
  removeOtpDigit,
  replaceOtpDigits,
  sanitizeOtp,
  serializeOtpCells,
  type OtpCells,
} from "./otp-input-model";

export interface SegmentedOtpInputProps {
  readonly value: string;
  readonly onChange: (value: string) => void;
  readonly label: string;
  readonly describedBy?: string;
  readonly disabled?: boolean;
  readonly autoFocus?: boolean;
  readonly required?: boolean;
  readonly length?: number;
  readonly className?: string;
}

export function SegmentedOtpInput({
  value,
  onChange,
  label,
  describedBy,
  disabled = false,
  autoFocus = false,
  required = false,
  length = 6,
  className = "",
}: Readonly<SegmentedOtpInputProps>) {
  const inputs = useRef<Array<HTMLInputElement | null>>([]);
  const autofocusRef = useRef(true);
  const cleanValue = sanitizeOtp(value, length);
  const [interaction, setInteraction] = useState<{
    readonly cells: OtpCells;
    readonly emittedValue: string;
  }>(() => ({
    cells: createOtpCells(cleanValue, length),
    emittedValue: cleanValue,
  }));
  const cells =
    cleanValue === interaction.emittedValue
      ? interaction.cells
      : createOtpCells(cleanValue, length);

  useEffect(() => {
    // Focus once, the first time the field is actually usable — not on every
    // re-enable. Re-focusing after a rejected code would drag the caret back
    // to cell one and past the error alert that just rendered, which is
    // exactly where a screen-reader user needs to stay. Keeping `disabled` in
    // the dependencies (rather than firing on mount alone) means a field that
    // mounts disabled still gets its one focus when it opens up.
    if (autoFocus && !disabled && autofocusRef.current) {
      autofocusRef.current = false;
      inputs.current[0]?.focus();
    }
  }, [autoFocus, disabled]);

  function focus(index: number) {
    inputs.current[Math.max(0, Math.min(index, length - 1))]?.focus();
  }

  function update(nextCells: OtpCells) {
    const nextValue = serializeOtpCells(nextCells);
    setInteraction({ cells: nextCells, emittedValue: nextValue });
    onChange(nextValue);
  }

  function insert(event: ChangeEvent<HTMLInputElement>, index: number) {
    const insertion = sanitizeOtp(event.target.value, length);
    if (!insertion) return;
    const nextCells = replaceOtpDigits(cells, insertion, index, length);
    update(nextCells);
    const insertionStart = insertion.length === length ? 0 : index;
    focus(Math.min(insertionStart + insertion.length, length - 1));
  }

  function handleKeyDown(
    event: KeyboardEvent<HTMLInputElement>,
    index: number,
  ) {
    if (event.key === "Backspace") {
      event.preventDefault();
      const next = removeOtpDigit(cells, index);
      update(next.cells);
      focus(next.focusIndex);
    } else if (event.key === "Delete") {
      event.preventDefault();
      update(clearOtpDigit(cells, index));
    } else if (event.key === "ArrowLeft") {
      event.preventDefault();
      focus(index - 1);
    } else if (event.key === "ArrowRight") {
      event.preventDefault();
      focus(index + 1);
    } else if (event.key === "Home") {
      event.preventDefault();
      focus(0);
    } else if (event.key === "End") {
      event.preventDefault();
      focus(length - 1);
    }
  }

  function handlePaste(event: ClipboardEvent<HTMLInputElement>, index: number) {
    const pasted = sanitizeOtp(event.clipboardData.getData("text"), length);
    if (!pasted) return;
    event.preventDefault();
    const nextCells = replaceOtpDigits(cells, pasted, index, length);
    update(nextCells);
    const insertionStart = pasted.length === length ? 0 : index;
    focus(Math.min(insertionStart + pasted.length, length - 1));
  }

  return (
    <fieldset
      aria-describedby={describedBy}
      className={`obiara-otp-fieldset ${className}`.trim()}
      disabled={disabled}
    >
      <legend>{label}</legend>
      <div className="obiara-otp-inputs">
        {Array.from({ length }, (_, index) => (
          <input
            aria-label={`${label}, digit ${index + 1} of ${length}`}
            autoComplete={index === 0 ? "one-time-code" : "off"}
            autoFocus={autoFocus && index === 0}
            enterKeyHint={index === length - 1 ? "done" : "next"}
            inputMode="numeric"
            key={index}
            maxLength={index === 0 ? length : 1}
            onChange={(event) => insert(event, index)}
            onFocus={(event) => event.currentTarget.select()}
            onKeyDown={(event) => handleKeyDown(event, index)}
            onPaste={(event) => handlePaste(event, index)}
            pattern="[0-9]*"
            ref={(element) => {
              inputs.current[index] = element;
            }}
            required={required}
            type="text"
            value={cells[index] ?? ""}
          />
        ))}
      </div>
    </fieldset>
  );
}
