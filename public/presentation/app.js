// Brand works page. Data comes from GET /works (protected by the Pres-Access cookie).
// Preview/work files are served by GET /presentation/:name/*filepath.
(function () {
  const grid = document.getElementById("grid");
  const brandTitle = document.getElementById("brandTitle");
  const emptyState = document.getElementById("emptyState");
  const emptyInner = document.getElementById("emptyInner");

  // The API filters works by the login cookie and never returns the brand name,
  // so show the name the user logged in with (saved by auth.js). This works
  // even when the brand has no works yet.
  try {
    const savedBrand = localStorage.getItem("brandName");
    if (savedBrand) brandTitle.textContent = savedBrand;
  } catch (e) {
    /* storage unavailable — keep the placeholder title */
  }

  // The cover lives in works/<brand>/<work>/preview/. The brand file route
  // /presentation/<work>/preview resolves that single-file folder to the image
  // itself, so we don't need to know its name or extension.
  function previewURL(work) {
    if (!work.work_name) return "";
    return "/presentation/" + encodeURIComponent(work.work_name) + "/preview";
  }

  // Opens the presentation viewer for a work (guest mode — brand comes from the
  // Pres-Access cookie server-side).
  function viewerURL(work) {
    return (
      "/static/presentation/view.html?work=" +
      encodeURIComponent(work.work_name || "")
    );
  }

  // Maps a stored status string to its badge CSS class. Rows always carry a
  // status (the DB column defaults to "в работе"), but fall back just in case.
  function statusClass(status) {
    switch ((status || "").trim().toLowerCase()) {
      case "сдан":
        return "status-done";
      case "на согласовании":
        return "status-review";
      case "правка":
        return "status-revision";
      default:
        return "status-progress";
    }
  }

  function statusLabel(status) {
    const s = (status || "").trim() || "в работе";
    return s.charAt(0).toUpperCase() + s.slice(1);
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

      const thumb = document.createElement("a");
      thumb.className = "thumb";
      thumb.href = viewerURL(work);
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

      const status = document.createElement("span");
      status.className = "work-status " + statusClass(work.status);
      status.textContent = statusLabel(work.status);

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

      card.append(thumb, name, status, desc, link);
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
