import { describe, expect, it } from "vitest";
import { expenseEntryTotal, parseLineAmount } from "./expense-day-totals";

describe("expenseEntryTotal", () => {
  it("uses totalAmount when set", () => {
    expect(expenseEntryTotal({ totalAmount: "100000", lines: [] })).toBe(100000);
  });

  it("sums lines when totalAmount is zero", () => {
    expect(
      expenseEntryTotal({
        totalAmount: "0",
        lines: [{ name: "a", amount: "50000" }, { name: "b", amount: "50000" }],
      }),
    ).toBe(100000);
  });

  it("parseLineAmount handles numbers", () => {
    expect(parseLineAmount(1000)).toBe(1000);
  });
});
