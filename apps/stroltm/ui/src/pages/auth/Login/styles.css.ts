import { style } from "@vanilla-extract/css";

/* eslint-disable perfectionist/sort-objects -- property order in style() is owned by vanilla-extract/concentric-order */
export const authLayout = style({
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  minHeight: "100vh",
});
/* eslint-enable perfectionist/sort-objects */
