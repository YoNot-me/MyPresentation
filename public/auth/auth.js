// Shared login logic for the brand ("Войти") and admin ("Вход админ") forms.
// The form declares its behaviour via data-* attributes, so one script serves both.
(function () {
  const form = document.getElementById("loginForm");
  if (!form) return;

  const errorEl = document.getElementById("formError");
  const userInput = document.getElementById("userField");
  const passInput = form.querySelector('input[name="password"]');
  const submitBtn = form.querySelector('button[type="submit"]');

  const endpoint = form.dataset.endpoint;
  const redirect = form.dataset.redirect;
  const userField = form.dataset.userField || "name";

  function setError(msg) {
    errorEl.textContent = msg || "";
  }

  form.addEventListener("submit", async function (e) {
    e.preventDefault();
    setError("");

    const user = userInput.value.trim();
    const password = passInput.value;

    if (!user || !password) {
      setError("Заполните все поля");
      return;
    }

    const payload = { [userField]: user, password: password };

    submitBtn.disabled = true;
    try {
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
        credentials: "same-origin",
      });

      // The server answers with a 3xx redirect on success (cookie is set),
      // which fetch follows to a JSON endpoint — we don't need its body.
      if (res.ok || res.redirected) {
        // Remember the brand name so the works page can show it as the title —
        // the server filters works by the cookie and never returns the name.
        if (userField === "name") {
          try {
            localStorage.setItem("brandName", user);
          } catch (e) {
            /* storage unavailable — the title just falls back to a placeholder */
          }
        }
        window.location.href = redirect;
        return;
      }

      if (res.status === 401 || res.status === 400) {
        setError("Неверное название бренда или пароль");
      } else {
        setError("Ошибка входа. Попробуйте ещё раз");
      }
    } catch (err) {
      setError("Не удалось связаться с сервером");
    } finally {
      submitBtn.disabled = false;
    }
  });
})();
