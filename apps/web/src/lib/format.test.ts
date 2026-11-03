import { describe, expect, it } from "vitest";
import { formatShortVND, formatVND } from "./format";

describe("formatVND", () => {
  it("formats numbers in vi-VN", () => {
    expect(formatVND(18500000)).toContain("đ");
    expect(formatVND("1000")).toContain("1");
  });
});

describe("formatShortVND", () => {
  it("abbreviates millions", () => {
    expect(formatShortVND(18_500_000)).toBe("18.5M đ");
  });
});
