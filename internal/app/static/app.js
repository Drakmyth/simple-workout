document.querySelectorAll("[data-auto-submit]").forEach((form) => {
  form.querySelector("select")?.addEventListener("change", () => form.requestSubmit());
  form.querySelector("[data-submit]")?.setAttribute("hidden", "");
});

const playback = document.querySelector("[data-playback]");
if (playback) {
  const slides = [...playback.querySelectorAll("[data-slide]")];
  const completion = playback.querySelector("[data-complete]");
  const action = document.querySelector("[data-playback-action]");
  const nextButton = action.querySelector("[data-playback-next]");
  const choices = new Map();
  let index = 0;

  const chooseVariant = (exercise) => {
    if (choices.has(exercise)) return choices.get(exercise);
    const preference = playback.dataset.imagePreference;
    let variant = preference;
    if (preference === "random") variant = Math.random() < 0.5 ? "male" : "female";
    if (preference === "alternating") {
      let next = "male";
      try { next = localStorage.getItem("playback-image-next") || "male"; } catch {}
      variant = next === "female" ? "female" : "male";
      try { localStorage.setItem("playback-image-next", variant === "male" ? "female" : "male"); } catch {}
    }
    choices.set(exercise, variant);
    return variant;
  };

  const show = () => {
    slides.forEach((slide) => slide.hidden = true);
    if (index >= slides.length) {
      completion.hidden = false;
      action.hidden = true;
      window.scrollTo({ top: 0 });
      return;
    }
    const slide = slides[index];
    slide.hidden = false;
    if (slide.dataset.kind === "set") {
      const variant = chooseVariant(slide.dataset.exercise);
      const image = slide.querySelector(".playback-image");
      image.src = image.dataset[`${variant}Src`];
      image.alt = image.dataset[`${variant}Alt`];
    }
    nextButton.textContent = index === slides.length - 1 ? "Finish" : "Next";
    action.hidden = false;
    window.scrollTo({ top: 0 });
  };

  nextButton.addEventListener("click", () => { index += 1; show(); });
  show();
}
