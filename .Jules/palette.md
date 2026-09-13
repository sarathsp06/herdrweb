## 2025-02-18 - Radiogroups for Settings Option Buttons
**Learning:** Custom button-grid option pickers (like theme and font size selectors) need `role="radiogroup"` on the container and `role="radio"` with `aria-checked` on each option button so screen readers convey selection state.
**Action:** Always pair visual active styling on button option groups with `role="radio"` and `aria-checked={$state}`.
