// Admin panel logic for both brands.html (grid of brands) and works.html
// (per-brand works management). The <script> tag declares which page via
// data-page. All data comes from the live admin JSON endpoints.
(function () {
  const page = document.currentScript.dataset.page;
  const stateMsg = document.getElementById("stateMsg");

  function setState(msg) {
    if (!stateMsg) return;
    stateMsg.textContent = msg || "";
    stateMsg.style.display = msg ? "block" : "none";
  }

  async function guardedFetch(url, opts) {
    const res = await fetch(url, Object.assign({ credentials: "same-origin" }, opts));
    // fetch follows redirects automatically, so res.redirected is true both
    // for a successful 303 (e.g. AddNewBrand -> /admin/brands) and for an
    // auth failure (ProtectedAdmin -> /auth/admin.html). Only the latter
    // should log the user out, so check where we actually ended up.
    if (res.status === 401 || res.url.indexOf("/auth/admin.html") !== -1) {
      window.location.href = "/auth/admin.html";
      throw new Error("unauthorized");
    }
    return res;
  }

  // POST /admin/brands/add — creates a brand with a name and password.
  async function addBrand(name, password) {
    const url = "/admin/brands/add";
    return guardedFetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: name, password: password }),
    });
  }

  // POST /admin/:brandName/works/add — multipart form: JSON "data" field,
  // optional single "preview" file, and required "files" (1..50) — this
  // mirrors what AdminService.AddNewWork reads from the multipart form.
  async function addWork(brand, fields, previewFile, files) {
    const fd = new FormData();
    fd.append("data", JSON.stringify(fields || {}));
    if (previewFile) fd.append("preview", previewFile);
    for (const file of files || []) {
      fd.append("files", file);
    }

    const url = "/admin/" + encodeURIComponent(brand) + "/works/add";
    return guardedFetch(url, { method: "POST", body: fd });
  }

  // ---------- Shared "add work" modal (used on brands.html and works.html) ----------
  // Wires up the modal once per page and returns openAddWork(brand). Pass
  // onSuccess to run something (e.g. refresh a list) after a successful add.
  function setupAddWorkModal(onSuccess) {
    const addWorkOverlay = document.getElementById("addWorkOverlay");
    if (!addWorkOverlay) return function () {};

    const addWorkBrandLabel = document.getElementById("addWorkBrandLabel");
    const newWorkName = document.getElementById("newWorkName");
    const newWorkUrl = document.getElementById("newWorkUrl");
    const newWorkDesc = document.getElementById("newWorkDesc");
    const newWorkPreview = document.getElementById("newWorkPreview");
    const newWorkPreviewLabel = document.getElementById("newWorkPreviewLabel");
    const newWorkPreviewText = document.getElementById("newWorkPreviewText");
    const newWorkFiles = document.getElementById("newWorkFiles");
    const newWorkFilesLabel = document.getElementById("newWorkFilesLabel");
    const newWorkFilesText = document.getElementById("newWorkFilesText");
    const MAX_WORK_FILES = 50;
    const addWorkError = document.getElementById("addWorkError");
    const addWorkSubmit = document.getElementById("addWorkSubmit");
    const addWorkCancel = document.getElementById("addWorkCancel");
    let addWorkBrand = null;

    function openAddWork(brand) {
      addWorkBrand = brand;
      addWorkBrandLabel.textContent = brand;
      newWorkName.value = "";
      newWorkUrl.value = "";
      newWorkDesc.value = "";
      newWorkPreview.value = "";
      newWorkPreviewText.textContent = "Загрузить превью";
      newWorkPreviewLabel.classList.remove("has-file");
      newWorkFiles.value = "";
      newWorkFilesText.textContent = "Загрузить файлы работы (до 50)";
      newWorkFilesLabel.classList.remove("has-file");
      addWorkError.textContent = "";
      addWorkOverlay.hidden = false;
      newWorkName.focus();
    }
    function closeAddWork() {
      addWorkOverlay.hidden = true;
      addWorkBrand = null;
    }

    addWorkCancel.addEventListener("click", closeAddWork);
    addWorkOverlay.addEventListener("click", (e) => {
      if (e.target === addWorkOverlay) closeAddWork();
    });
    newWorkPreview.addEventListener("change", () => {
      const file = newWorkPreview.files[0];
      newWorkPreviewText.textContent = file ? file.name : "Загрузить превью";
      newWorkPreviewLabel.classList.toggle("has-file", !!file);
    });
    newWorkFiles.addEventListener("change", () => {
      const count = newWorkFiles.files.length;
      if (count > MAX_WORK_FILES) {
        addWorkError.textContent =
            "Можно загрузить не более " + MAX_WORK_FILES + " файлов";
        newWorkFiles.value = "";
        newWorkFilesText.textContent = "Загрузить файлы работы (до 50)";
        newWorkFilesLabel.classList.remove("has-file");
        return;
      }
      addWorkError.textContent = "";
      newWorkFilesText.textContent = count
          ? "Выбрано файлов: " + count
          : "Загрузить файлы работы (до 50)";
      newWorkFilesLabel.classList.toggle("has-file", count > 0);
    });

    addWorkSubmit.addEventListener("click", async () => {
      if (!addWorkBrand) return;
      const name = newWorkName.value.trim();
      if (!name) {
        addWorkError.textContent = "Введите название работы";
        return;
      }
      if (newWorkFiles.files.length === 0) {
        addWorkError.textContent = "Добавьте хотя бы один файл работы";
        return;
      }

      addWorkSubmit.disabled = true;
      try {
        const res = await addWork(
            addWorkBrand,
            {
              work_name: name,
              url: newWorkUrl.value.trim(),
              description: newWorkDesc.value,
            },
            newWorkPreview.files[0],
            newWorkFiles.files
        );
        if (!res.ok && !res.redirected) {
          addWorkError.textContent = "Не удалось добавить работу";
          return;
        }
        closeAddWork();
        if (onSuccess) onSuccess();
      } catch (e) {
        if (e.message !== "unauthorized") {
          addWorkError.textContent = "Ошибка соединения";
        }
      } finally {
        addWorkSubmit.disabled = false;
      }
    });

    return openAddWork;
  }

  // ---------- Brands grid ----------
  async function initBrands() {
    const gridEl = document.getElementById("brandsGrid");

    // ---- "Добавить бренд" modal ----
    const addBrandBtn = document.getElementById("addBrandBtn");
    const addBrandOverlay = document.getElementById("addBrandOverlay");
    const newBrandName = document.getElementById("newBrandName");
    const newBrandPassword = document.getElementById("newBrandPassword");
    const addBrandError = document.getElementById("addBrandError");
    const addBrandSubmit = document.getElementById("addBrandSubmit");
    const addBrandCancel = document.getElementById("addBrandCancel");

    function openAddBrand() {
      newBrandName.value = "";
      newBrandPassword.value = "";
      addBrandError.textContent = "";
      addBrandOverlay.hidden = false;
      newBrandName.focus();
    }
    function closeAddBrand() {
      addBrandOverlay.hidden = true;
    }

    addBrandBtn.addEventListener("click", openAddBrand);
    addBrandCancel.addEventListener("click", closeAddBrand);
    addBrandOverlay.addEventListener("click", (e) => {
      if (e.target === addBrandOverlay) closeAddBrand();
    });

    addBrandSubmit.addEventListener("click", async () => {
      const name = newBrandName.value.trim();
      const password = newBrandPassword.value;
      if (!name) {
        addBrandError.textContent = "Введите название бренда";
        return;
      }
      if (!password) {
        addBrandError.textContent = "Введите пароль";
        return;
      }

      addBrandSubmit.disabled = true;
      try {
        const res = await addBrand(name, password);
        if (!res.ok && !res.redirected) {
          addBrandError.textContent = "Не удалось создать бренд";
          return;
        }
        closeAddBrand();
        loadBrands();
      } catch (e) {
        if (e.message !== "unauthorized") {
          addBrandError.textContent = "Ошибка соединения";
        }
      } finally {
        addBrandSubmit.disabled = false;
      }
    });

    // ---- "Добавить работу" modal (opened via "+" on a brand bubble) ----
    const openAddWork = setupAddWorkModal();

    // ---- brands list ----
    async function loadBrands() {
      try {
        const res = await guardedFetch("/admin/brands");
        if (!res.ok) return setState("Не удалось загрузить бренды");
        const brands = await res.json();
        if (!Array.isArray(brands) || brands.length === 0) {
          gridEl.innerHTML = "";
          return setState("Брендов пока нет");
        }
        setState("");
        gridEl.innerHTML = "";
        for (const b of brands) {
          const wrap = document.createElement("div");
          wrap.className = "brand-item-wrap";

          const a = document.createElement("a");
          a.className = "brand-item";
          a.textContent = b.name;
          a.href =
              "/admin/panel/works.html?brand=" + encodeURIComponent(b.name);

          const plus = document.createElement("button");
          plus.type = "button";
          plus.className = "brand-add-work";
          plus.textContent = "+";
          plus.title = "Добавить работу";
          plus.addEventListener("click", () => openAddWork(b.name));

          wrap.append(a, plus);
          gridEl.appendChild(wrap);
        }
      } catch (e) {
        if (e.message !== "unauthorized") setState("Ошибка соединения");
      }
    }

    loadBrands();
  }

  // ---------- Works management ----------
  // Preview is always at .../preview — the folder holds exactly one file,
  // and ServingWork resolves it server-side regardless of name/extension.
  function previewURL(brand, work) {
    return (
        "/admin/" +
        encodeURIComponent(brand) +
        "/serve/" +
        encodeURIComponent(work.work_name) +
        "/preview"
    );
  }

  // Send a partial update to PUT /admin/:brand/:work/change.
  // The server expects a multipart form with an optional JSON "data" field and
  // an optional "preview" file — FormData produces exactly that. When only the
  // cover is replaced (no field changes) we omit "data" entirely, so the server
  // sees an empty PostForm("data") and skips the needless DB update.
  async function changeWork(brand, work, fields, file) {
    const fd = new FormData();
    if (fields && Object.keys(fields).length > 0) {
      fd.append("data", JSON.stringify(fields));
    }
    if (file) fd.append("preview", file);

    const url =
        "/admin/" +
        encodeURIComponent(brand) +
        "/" +
        encodeURIComponent(work) +
        "/change";

    const res = await guardedFetch(url, { method: "PUT", body: fd });
    if (!res.ok && !res.redirected) {
      alert("Не удалось сохранить изменения");
      return false;
    }
    return true;
  }

  function initWorks() {
    const params = new URLSearchParams(window.location.search);
    const brand = params.get("brand");
    const gridEl = document.getElementById("grid");
    const headingEl = document.getElementById("worksHeading");
    const renameBtn = document.getElementById("renameBrandBtn");
    const addWorkBtn = document.getElementById("addWorkBtn");
    const coverInput = document.getElementById("coverInput");

    if (!brand) {
      setState("Не указан бренд");
      return;
    }

    headingEl.textContent = brand + " / Работы";

    const openAddWork = setupAddWorkModal(() => load());
    addWorkBtn.addEventListener("click", () => openAddWork(brand));

    renameBtn.addEventListener("click", async function () {
      const name = prompt("Новое название бренда:", brand);
      if (!name || name.trim() === "" || name === brand) return;
      const res = await guardedFetch(
          "/admin/" + encodeURIComponent(brand) + "/rename",
          {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name: name.trim() }),
          }
      );
      if (res.ok || res.redirected) {
        window.location.href =
            "/admin/panel/works.html?brand=" + encodeURIComponent(name.trim());
      } else {
        alert("Не удалось переименовать бренд");
      }
    });

    // one shared file picker for "Редактировать обложку"
    let pendingCoverWork = null;
    coverInput.addEventListener("change", async function () {
      const file = coverInput.files[0];
      const work = pendingCoverWork;
      coverInput.value = "";
      pendingCoverWork = null;
      if (!file || !work) return;
      if (await changeWork(brand, work, {}, file)) load();
    });

    async function load() {
      try {
        const res = await guardedFetch(
            "/admin/" + encodeURIComponent(brand) + "/works"
        );
        if (res.status === 404) {
          setState("У бренда пока нет работ");
          gridEl.innerHTML = "";
          return;
        }
        if (!res.ok) return setState("Не удалось загрузить работы");
        const works = await res.json();
        if (!Array.isArray(works) || works.length === 0) {
          setState("У бренда пока нет работ");
          return;
        }
        setState("");
        render(works);
      } catch (e) {
        if (e.message !== "unauthorized") setState("Ошибка соединения");
      }
    }

    function render(works) {
      gridEl.innerHTML = "";
      for (const work of works) {
        const name = work.work_name;

        const card = document.createElement("article");
        card.className = "work-card";

        const thumb = document.createElement("a");
        thumb.className = "thumb";
        thumb.href =
            "/static/presentation/view.html?brand=" +
            encodeURIComponent(brand) +
            "&work=" +
            encodeURIComponent(name || "");
        const src = previewURL(brand, work);
        if (src) {
          const img = document.createElement("img");
          img.src = src;
          img.alt = name || "";
          img.loading = "lazy";
          img.onerror = () => img.remove();
          thumb.appendChild(img);
        }

        const title = document.createElement("h3");
        title.className = "work-name";
        title.textContent = name || "";

        const desc = document.createElement("p");
        desc.className = "work-desc";
        desc.textContent = work.description || "";

        const actions = document.createElement("div");
        actions.className = "card-actions";

        actions.appendChild(
            makeAction("Редактировать название", async () => {
              const val = prompt("Новое название работы:", name);
              if (!val || val.trim() === "" || val === name) return;
              if (await changeWork(brand, name, { work_name: val.trim() })) load();
            })
        );
        actions.appendChild(
            makeAction("Редактировать обложку", () => {
              pendingCoverWork = name;
              coverInput.click();
            })
        );
        actions.appendChild(
            makeAction("Редактировать описание", async () => {
              const val = prompt("Новое описание:", work.description || "");
              if (val === null) return;
              if (await changeWork(brand, name, { description: val })) load();
            })
        );
        actions.appendChild(
            makeAction("Редактировать ссылку", async () => {
              const val = prompt("Новая ссылка на диск:", work.url || "");
              if (val === null) return;
              if (await changeWork(brand, name, { url: val })) load();
            })
        );

        card.append(thumb, title, desc, actions);
        gridEl.appendChild(card);
      }
    }

    function makeAction(label, handler) {
      const btn = document.createElement("button");
      btn.className = "action";
      btn.type = "button";
      btn.textContent = label;
      btn.addEventListener("click", handler);
      return btn;
    }

    load();
  }

  // Logout button (present on every admin page). POST /logout/admin clears the
  // admin session server-side; we then send the user to the admin login page
  // regardless of the POST result so the panel is never left half-open.
  const logoutBtn = document.getElementById("logoutBtn");
  if (logoutBtn) {
    logoutBtn.addEventListener("click", async function () {
      logoutBtn.disabled = true;
      try {
        await fetch("/logout/admin", {
          method: "POST",
          credentials: "same-origin",
        });
      } catch (e) {
        /* ignore — redirect to login regardless */
      } finally {
        window.location.href = "/auth/admin.html";
      }
    });
  }

  if (page === "brands") initBrands();
  else if (page === "works") initWorks();
})();