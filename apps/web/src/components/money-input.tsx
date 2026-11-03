"use client";

import TextField, { TextFieldProps } from "@mui/material/TextField";
import { digitsOnly, formatMoneyTyping } from "@/lib/money-input";

type MoneyInputProps = Omit<TextFieldProps, "value" | "onChange"> & {
  value: string;
  onValueChange: (digits: string) => void;
};

export function MoneyInput({ value, onValueChange, ...props }: MoneyInputProps) {
  const display = formatMoneyTyping(value);

  return (
    <TextField
      {...props}
      value={display}
      onChange={(e) => onValueChange(digitsOnly(e.target.value))}
      slotProps={{
        htmlInput: { inputMode: "numeric", autoComplete: "off" },
        ...props.slotProps,
      }}
    />
  );
}
