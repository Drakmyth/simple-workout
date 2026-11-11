document.querySelectorAll("[data-auto-submit]").forEach((form) => {
  form.querySelector("select")?.addEventListener("change", () => form.requestSubmit());
  form.querySelector("[data-submit]")?.setAttribute("hidden", "");
});

const playback = document.querySelector("[data-playback]");
if (playback) {
  playback.querySelector("h1")?.focus();
  const display = playback.querySelector("[data-rest-timer]");
  if (display) {
    let remaining = Number(display.dataset.restSeconds);
    if (remaining > 0) {
      const timer = setInterval(() => {
        remaining -= 1;
        display.textContent = `${remaining} ${remaining === 1 ? "second" : "seconds"}`;
        if (remaining <= 0) clearInterval(timer);
      }, 1000);
    }
  }
}
