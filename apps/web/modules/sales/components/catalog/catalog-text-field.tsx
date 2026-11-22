import { Field, inputClass } from "@/components/primitives";

interface Props {
  label: string;
  value: string;
  onChange: (value: string) => void;
  required?: boolean;
  type?: string;
  maxLength?: number;
}

/** Labelled input; number inputs get the API's money bounds. */
export function CatalogTextField({ label, value, onChange, required = false, type = "text", maxLength }: Props) {
  return (
    <Field title={label}>
      <input
        className={inputClass}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required={required}
        type={type}
        maxLength={maxLength}
        {...(type === "number" ? { min: "0", max: "9999999999999.99", step: "0.01" } : {})}
      />
    </Field>
  );
}
