// Brand works page. Data comes from GET /works (protected by the Pres-Access cookie).
// Preview/work files are served by GET /presentation/:name/*filepath.
(function () {
  const grid = document.getElementById("grid");
  const brandTitle = document.getElementById("brandTitle");
  const emptyState = document.getElementById("emptyState");
  const emptyInner = document.getElementById("emptyInner");

  // The `preview` column has no json tag on the server, so it may arrive as
  // either "Preview" or "preview" — accept both.
  function previewOf(work) {
    return work.Preview || work.preview || "";
  }

  // Stored preview looks like "works/<brand>/<work>/preview/preview.png".
  // The brand-facing file route is /presentation/<work_name>/<rel>, so we only
  // need the file name and rebuild a safe URL from it.
  function previewURL(work) {
    const stored = previewOf(work);
    if (!stored) return "";
    const parts = stored.split("/").filter(Boolean);
    const file = parts[parts.length - 1];
    if (!file) return "";
    return (
      "/presentation/" +
      encodeURIComponent(work.work_name) +
      "/preview/" +
      encodeURIComponent(file)
    );
  }

  function showMessage(msg) {
    grid.hidden = true;
    emptyState.hidden = false;
    emptyInner.textContent = msg;
  }

  function render(works) {
    if (!Array.isArray(works) || works.length === 0) {
      showMessage("Пока нет работ");
      return;
    }

    if (works[0] && works[0].brand) {
      brandTitle.textContent = works[0].brand;
    }

    emptyState.hidden = true;
    grid.hidden = false;
    grid.innerHTML = "";

    for (const work of works) {
      const card = document.createElement("article");
      card.className = "work-card";

      const thumb = document.createElement("div");
      thumb.className = "thumb";
      const src = previewURL(work);
      if (src) {
        const img = document.createElement("img");
        img.src = src;
        img.alt = work.work_name || "";
        img.loading = "lazy";
        img.onerror = () => img.remove();
        thumb.appendChild(img);
      }

      const name = document.createElement("h3");
      name.className = "work-name";
      name.textContent = work.work_name || "";

      const desc = document.createElement("p");
      desc.className = "work-desc";
      desc.textContent = work.description || "";

      const link = document.createElement("a");
      link.className = "btn-block";
      link.textContent = "Ссылка на диске";
      if (work.url) {
        link.href = work.url;
        link.target = "_blank";
        link.rel = "noopener noreferrer";
      } else {
        link.classList.add("is-disabled");
        link.href = "#";
      }

      card.append(thumb, name, desc, link);
      grid.appendChild(card);
    }
  }

  async function load() {
    try {
      const res = await fetch("/works", { credentials: "same-origin" });
      if (res.status === 401 || res.redirected) {
        window.location.href = "/auth";
        return;
      }
      if (!res.ok) {
        showMessage("Не удалось загрузить работы");
        return;
      }
      const data = await res.json();
      render(data);
    } catch (err) {
      showMessage("Ошибка соединения с сервером");
    }
  }

  load();
})();
