import { describe, expect, it } from "vitest";

import {
  getAdminPageTitle,
  getWrappedFocusIndex,
  isFocusCandidateState,
  isRailGroupVisible,
} from "./admin-shell-model";

describe("admin shell behavior", () => {
  it("keeps the parent desk title on nested detail routes", () => {
    expect(getAdminPageTitle("/incidents/INC-1042")).toBe("Incidents");
    expect(getAdminPageTitle("/operators/role/admin")).toBe(
      "Operators & roles",
    );
    expect(getAdminPageTitle("/account/security")).toBe("Operator account");
  });

  it("creates a readable fallback for routes outside the rail", () => {
    expect(getAdminPageTitle("/")).toBe("Command centre");
    expect(getAdminPageTitle("/audit-history")).toBe("Audit history");
  });

  it("wraps keyboard focus only at drawer boundaries", () => {
    expect(getWrappedFocusIndex(0, 4, true)).toBe(3);
    expect(getWrappedFocusIndex(3, 4, false)).toBe(0);
    expect(getWrappedFocusIndex(1, 4, false)).toBeNull();
    expect(getWrappedFocusIndex(-1, 4, false)).toBe(0);
    expect(getWrappedFocusIndex(-1, 4, true)).toBe(3);
    expect(getWrappedFocusIndex(0, 0, true)).toBeNull();
  });

  it("excludes hidden, inert, and collapsed-group controls from focus containment", () => {
    const visible = {
      hidden: false,
      inert: false,
      collapsedGroup: false,
      display: "block",
      visibility: "visible",
    };
    expect(isFocusCandidateState(visible)).toBe(true);
    expect(isFocusCandidateState({ ...visible, hidden: true })).toBe(false);
    expect(isFocusCandidateState({ ...visible, inert: true })).toBe(false);
    expect(isFocusCandidateState({ ...visible, collapsedGroup: true })).toBe(
      false,
    );
    expect(isFocusCandidateState({ ...visible, display: "none" })).toBe(false);
    expect(isFocusCandidateState({ ...visible, visibility: "hidden" })).toBe(
      false,
    );
  });

  it("never strands links in a collapsed rail group", () => {
    expect(isRailGroupVisible(false, true)).toBe(true);
    expect(isRailGroupVisible(true, true)).toBe(true);
    expect(isRailGroupVisible(false, false)).toBe(false);
  });
});
