// Lightweight visual effects for your board.
// Import this in index.html: <script src="/static/effects.js" defer></script>

(function () {
  // Helper: find cell by its value (expects data-val="N" on each .cell)
  function findCell(cellValue) {
    return document.querySelector(`.cell[data-val="${cellValue}"]`);
  }

  // Trigger micro pop + optional ripple (for teleports)
  window.fxLand = function (cellValue, teleported = false) {
    const el = findCell(cellValue);
    if (!el) return;

    // retrigger CSS animation by toggling class
    el.classList.remove('pop');
    // force reflow
    void el.offsetWidth;
    el.classList.add('pop');

    if (teleported) {
      el.classList.add('ripple');
      setTimeout(() => el.classList.remove('ripple'), 600);
    }
  };

  // Glow the whole grid for big moments (first finish, etc.)
  window.fxGlowBoard = function () {
    const grid = document.querySelector('.grid');
    if (!grid) return;
    grid.classList.remove('glow');
    void grid.offsetWidth;
    grid.classList.add('glow');
    setTimeout(() => grid.classList.remove('glow'), 700);
  };

  // Roll animation for a dice element (pass element or selector)
  window.rollDice = function (elOrSelector) {
    const el = (typeof elOrSelector === 'string') ? document.querySelector(elOrSelector) : elOrSelector;
    if (!el) return;
    el.classList.remove('roll');
    void el.offsetWidth;
    el.classList.add('roll');
    setTimeout(() => el.classList.remove('roll'), 620);
  };

  // Update a progress pill width (0–100)
  window.setProgress = function (selector, pct) {
    const bar = document.querySelector(selector);
    if (!bar) return;
    const clamped = Math.max(0, Math.min(100, pct));
    bar.style.width = clamped + '%';
  };

  // Optional: quick SSE hook examples (uncomment & adapt to your events)
  /*
  const es = new EventSource('/events');
  es.addEventListener('moved', (e) => {
    const d = JSON.parse(e.data); // { cell_value, teleported }
    fxLand(d.cell_value, d.teleported);
  });
  es.addEventListener('big-moment', () => fxGlowBoard());
  // When rolling:
  // rollDice('#dice');
  */
})();
