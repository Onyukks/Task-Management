import { describe, it, expect } from "vitest";
import { toQueryString } from "./api";

describe("toQueryString", () => {
  it("omits empty params", () => {
    expect(toQueryString({})).toBe("");
    expect(toQueryString({ status: "", search: "" })).toBe("");
  });

  it("composes filter + search + sort + pagination together", () => {
    const qs = toQueryString({
      status: "in_progress",
      search: "report",
      sortBy: "due_date",
      sortDir: "asc",
      page: 2,
      pageSize: 10,
    });
    const params = new URLSearchParams(qs);
    expect(params.get("status")).toBe("in_progress");
    expect(params.get("search")).toBe("report");
    expect(params.get("sortBy")).toBe("due_date");
    expect(params.get("sortDir")).toBe("asc");
    expect(params.get("page")).toBe("2");
    expect(params.get("pageSize")).toBe("10");
  });
});
