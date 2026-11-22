import { inputClass } from "@/components/primitives";

interface Props {
  id: string;
  value: string;
  disabled: boolean;
  onChange: (value: string) => void;
}

/** Masked Bot Token field; its help text must use the id `${id}-help`. */
export function TelegramTokenInput({ id, value, disabled, onChange }: Props) {
  return (
    <input
      id={id}
      type="password"
      autoComplete="off"
      spellCheck={false}
      required
      maxLength={128}
      pattern="[0-9]{1,20}:[A-Za-z0-9_\-]{30,80}"
      placeholder="123456789:AA…"
      aria-describedby={`${id}-help`}
      className={inputClass}
      disabled={disabled}
      value={value}
      onChange={(e) => onChange(e.target.value.trim())}
    />
  );
}
