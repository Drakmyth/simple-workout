document.querySelectorAll("[data-auto-submit]").forEach((form) => {
  form.querySelector("select")?.addEventListener("change", () => form.requestSubmit());
  form.querySelector("[data-submit]")?.setAttribute("hidden", "");
});
