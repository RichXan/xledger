# Design System Documentation: Financial Precision & Editorial Clarity

## 1. Overview & Creative North Star: "The Architectural Ledger"
This design system moves beyond the standard "SaaS Dashboard" aesthetic to embrace a "High-End Editorial" experience. The Creative North Star, **The Architectural Ledger**, treats financial data not as a series of rows and columns, but as a structured, premium environment. 

We break the "template" look by using **intentional asymmetry** and **tonal depth**. Rather than relying on heavy borders and boxes, we use the "Bento Grid" philosophy—grouping related data into organic, logical clusters that feel like a curated gallery of information. The result is an interface that feels authoritative, secure, and profoundly calm.

---

## 2. Colors: Tonal Architecture
The palette is rooted in a deep corporate blue for authority, balanced by a vibrant "Mint" for growth and clarity.

### The "No-Line" Rule
**Explicit Instruction:** Do not use 1px solid borders to section off the UI. 
Boundaries must be defined solely through background color shifts. For example, a `surface-container-low` section should sit directly on a `surface` background. This creates a "seamless" transition that feels sophisticated rather than "boxed in."

### Surface Hierarchy & Nesting
Treat the UI as physical layers of fine paper or frosted glass. 
- **Base Layer:** `surface` (#f7f9fc)
- **Primary Content Area:** `surface-container-low` (#f2f4f7)
- **High-Focus Cards:** `surface-container-lowest` (#ffffff)
- **Active/Hover Elements:** `surface-container-high` (#e6e8eb)

### The "Glass & Gradient" Rule
To elevate the brand above generic competitors, use **Glassmorphism** for floating elements (like modals or dropdowns). Use `surface-container-lowest` at 80% opacity with a `backdrop-blur` of 12px. 
*Signature Detail:* Use a subtle linear gradient on primary CTAs, transitioning from `primary` (#00327d) to `primary-container` (#0047ab) at a 135-degree angle. This adds "visual soul" and depth.

---

## 3. Typography: The Editorial Voice
We utilize a dual-font strategy to balance corporate stability with modern elegance.

*   **Display & Headlines (Manrope):** Chosen for its geometric precision and wide apertures. It feels expensive and modern.
    *   `display-lg` (3.5rem): Use for high-level hero numbers or total assets.
    *   `headline-md` (1.75rem): Use for page titles and major section headers.
*   **Body & Labels (Inter):** A functional workhorse designed for maximum legibility in dense financial data.
    *   `body-md` (0.875rem): Standard for all ledger entries and descriptions.
    *   `label-sm` (0.6875rem): All-caps for metadata, using `letter-spacing: 0.05rem`.

---

## 4. Elevation & Depth: Tonal Layering
In this system, depth is a function of light and tone, not structure.

*   **The Layering Principle:** Place a `surface-container-lowest` card on a `surface-container-low` background. The subtle shift in hex code is enough to signal importance without visual clutter.
*   **Ambient Shadows:** Use only for elements that "float" above the page (e.g., Modals). 
    *   *Spec:* `0px 12px 32px rgba(25, 28, 30, 0.04)`. Note the 4% opacity; the shadow should be felt, not seen.
*   **The "Ghost Border" Fallback:** If accessibility requires a container edge, use the `outline-variant` token at 15% opacity. Never use a 100% opaque border.
*   **Glassmorphism Depth:** For persistent sidebars or floating action toolbars, use semi-transparent `surface` colors. This allows the financial charts to "bleed through" the UI, making the application feel like a single, integrated canvas.

---

## 5. Components: Minimalist Utility

### Buttons: The Signature Action
*   **Primary:** A 135° gradient from `primary` to `primary-container`. `border-radius: lg` (0.5rem). High-contrast white text.
*   **Secondary:** Ghost style. No background, only a `primary` text label. 
*   **States:** On hover, increase the gradient saturation. On press, scale the button to 0.98 for tactile feedback.

### The Bento Card
*   **Layout:** Use `spacing-5` (1.1rem) for internal padding. 
*   **Style:** No borders. Background: `surface-container-lowest`. 
*   **Visual Soul:** Incorporate a 4px vertical "accent stripe" of `tertiary-fixed` (#93f993) on the left edge of cards representing positive cash flow.

### Input Fields: Transparent Precision
*   **Style:** Minimalist. No background fill. Only an underline using `outline-variant` (#c3c6d5). 
*   **Focus State:** The underline transforms into a 2px `primary` line with a soft `surface-tint` glow.

### Tree-View Lists (Ledger Special)
*   **Forbid Dividers:** Do not use horizontal lines between rows.
*   **Separation:** Use `spacing-2` (0.4rem) between items. Active items should use a subtle `surface-container-high` background with `border-radius: md`.

### Label Chips
*   **Style:** Pill-shaped (`rounded-full`). 
*   **Colors:** Use `tertiary-fixed` for "Approved" and `error-container` for "Flagged." Keep text color deep (`on-tertiary-fixed` or `on-error-container`) for contrast.

---

## 6. Do's and Don'ts

### Do
*   **DO** use negative space as a separator. If you think you need a line, try adding `spacing-8` instead.
*   **DO** use "Manrope" for all numbers in data visualizations. Its geometric nature aligns perfectly with the Bento grid.
*   **DO** group related financial metrics into a single "Bento Card" to reduce cognitive load.

### Don't
*   **DON'T** use pure black (#000000) for text. Use `on-surface` (#191c1e) to maintain a soft, premium feel.
*   **DON'T** use standard "Drop Shadows" on cards. Stick to Tonal Layering (Surface-on-Surface).
*   **DON'T** use 100% opacity for backgrounds in modal overlays. Always use a backdrop-blur (Glassmorphism) to maintain the "Architectural Ledger" depth.