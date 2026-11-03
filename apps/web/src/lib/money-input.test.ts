import { describe, expect, it } from "vitest";
import { digitsOnly, formatMoneyTyping, hasPositiveAmount } from "./money-input";

describe("money-input", () => {
  it("digitsOnly strips commas", () => {
    expect(digitsOnly("12,222,222")).toBe("12222222");
  });

  it("formatMoneyTyping groups vi-VN", () => {
    expect(formatMoneyTyping("12222222")).toBe("12.222.222");
  });

  it("hasPositiveAmount", () => {
    expect(hasPositiveAmount("1000")).toBe(true);
    expect(hasPositiveAmount("")).toBe(false);
    expect(hasPositiveAmount("0")).toBe(false);
  });
});
