(function () {
  "use strict";
  var root = window.__MD2SITE_ROOT__ || "";
  var index = (window.__MD2SITE_INDEX__ && window.__MD2SITE_INDEX__.pages) || [];
  var input = document.getElementById("search");
  var results = document.getElementById("search-results");
  var nav = document.getElementById("nav");
  var sidebar = document.getElementById("sidebar");
  var toggle = document.getElementById("sidebar-toggle");

  if (toggle) {
    toggle.addEventListener("click", function () {
      sidebar.classList.toggle("open");
    });
  }

  // --- Collapsible / resizable columns (state kept in localStorage) ---
  function store(key, value) {
    try { localStorage.setItem("md2site." + key, value); } catch (e) {}
  }
  function cssPx(name) {
    return parseInt(getComputedStyle(document.documentElement).getPropertyValue(name), 10) || 0;
  }

  function initViewToggle(btn, cls, key, labels) {
    if (!btn) return;
    var sync = function () {
      var visible = !document.body.classList.contains(cls);
      btn.classList.toggle("active", visible);
      btn.setAttribute("aria-pressed", visible ? "true" : "false");
      btn.title = visible ? labels.hide : labels.show;
    };
    btn.addEventListener("click", function () {
      var collapsed = document.body.classList.toggle(cls);
      store(key, collapsed ? "1" : "0");
      sync();
    });
    sync();
  }
  initViewToggle(document.getElementById("view-nav"), "sidebar-collapsed", "sidebarCollapsed", {
    show: "Show navigation",
    hide: "Hide navigation"
  });

  function initOutlineMenu(menu, btn, dropdown) {
    if (!menu || !btn || !dropdown) return;
    var items = dropdown.querySelectorAll("[data-action]");
    var toggleItem = dropdown.querySelector('[data-action="toggle"]');

    var sync = function () {
      var visible = !document.body.classList.contains("outline-collapsed");
      var left = document.body.classList.contains("outline-left");
      btn.classList.toggle("active", visible);
      btn.setAttribute("aria-pressed", visible ? "true" : "false");
      if (toggleItem) toggleItem.textContent = visible ? "Hide" : "Show";
      items.forEach(function (item) {
        var action = item.getAttribute("data-action");
        var selected = (action === "left" && left) || (action === "right" && !left);
        item.classList.toggle("selected", selected);
        item.setAttribute("aria-checked", selected ? "true" : "false");
      });
    };

    var close = function () {
      dropdown.hidden = true;
      btn.setAttribute("aria-expanded", "false");
    };

    var open = function () {
      sync();
      dropdown.hidden = false;
      btn.setAttribute("aria-expanded", "true");
    };

    btn.addEventListener("click", function (e) {
      e.stopPropagation();
      if (dropdown.hidden) open();
      else close();
    });

    items.forEach(function (item) {
      item.addEventListener("click", function (e) {
        e.stopPropagation();
        var action = item.getAttribute("data-action");
        if (action === "left") {
          document.body.classList.add("outline-left");
          store("outlineSide", "left");
        } else if (action === "right") {
          document.body.classList.remove("outline-left");
          store("outlineSide", "right");
        } else if (action === "toggle") {
          var collapsed = document.body.classList.toggle("outline-collapsed");
          store("outlineCollapsed", collapsed ? "1" : "0");
        }
        sync();
        close();
      });
    });

    document.addEventListener("click", function () { close(); });
    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape") close();
    });
    sync();
  }
  initOutlineMenu(
    document.getElementById("view-outline-menu"),
    document.getElementById("view-outline"),
    document.getElementById("view-outline-dropdown")
  );

  function initResize(handle, varName, key, min, max, widthFn) {
    if (!handle) return;
    handle.addEventListener("mousedown", function (e) {
      e.preventDefault();
      handle.classList.add("dragging");
      document.body.classList.add("dragging");
      var onMove = function (ev) {
        var w = Math.min(max, Math.max(min, widthFn(ev)));
        document.documentElement.style.setProperty(varName, w + "px");
      };
      var onUp = function () {
        handle.classList.remove("dragging");
        document.body.classList.remove("dragging");
        store(key, cssPx(varName) + "px");
        window.removeEventListener("mousemove", onMove);
        window.removeEventListener("mouseup", onUp);
      };
      window.addEventListener("mousemove", onMove);
      window.addEventListener("mouseup", onUp);
    });
  }
  initResize(document.getElementById("sidebar-resize"), "--sidebar-width", "sidebarWidth",
    180, 520, function (ev) { return ev.clientX; });
  initResize(document.getElementById("outline-resize"), "--outline-width", "outlineWidth",
    160, 440, function (ev) {
      if (document.body.classList.contains("outline-left")) {
        var offset = document.body.classList.contains("sidebar-collapsed") ? 0 : cssPx("--sidebar-width");
        return ev.clientX - offset;
      }
      return window.innerWidth - ev.clientX;
    });

  // Outline scroll spy: highlight the heading currently at the top.
  var outline = document.getElementById("outline");
  if (outline) {
    var outlineLinks = {};
    outline.querySelectorAll('a[href^="#"]').forEach(function (a) {
      outlineLinks[decodeURIComponent(a.getAttribute("href").slice(1))] = a;
    });
    var headings = Array.prototype.filter.call(
      document.querySelectorAll("main h1[id], main h2[id], main h3[id], main h4[id]"),
      function (h) { return outlineLinks[h.id]; }
    );
    var activeLink = null;
    var updateSpy = function () {
      var cur = null;
      for (var i = 0; i < headings.length; i++) {
        if (headings[i].getBoundingClientRect().top <= 80) cur = headings[i];
        else break;
      }
      var link = cur ? outlineLinks[cur.id] : null;
      if (link === activeLink) return;
      if (activeLink) activeLink.classList.remove("current");
      if (link) link.classList.add("current");
      activeLink = link;
    };
    var ticking = false;
    window.addEventListener("scroll", function () {
      if (ticking) return;
      ticking = true;
      requestAnimationFrame(function () { ticking = false; updateSpy(); });
    }, { passive: true });
    updateSpy();
  }

  if (!input) return;

  // Substring search over normalized title+body. Works for Japanese text
  // (no tokenizer needed) and English alike. All terms must match.
  var docs = index.map(function (p) {
    return { p: p, title: p.t.toLowerCase(), body: p.b.toLowerCase() };
  });

  function search(query) {
    var terms = query.toLowerCase().split(/\s+/).filter(Boolean);
    if (!terms.length) return [];
    var hits = [];
    for (var i = 0; i < docs.length; i++) {
      var d = docs[i], score = 0, firstPos = -1, ok = true;
      for (var j = 0; j < terms.length; j++) {
        var t = terms[j];
        var inTitle = d.title.indexOf(t) >= 0;
        var pos = d.body.indexOf(t);
        if (!inTitle && pos < 0) { ok = false; break; }
        if (inTitle) score += 10;
        if (pos >= 0) {
          score += 1;
          if (firstPos < 0) firstPos = pos;
        }
      }
      if (ok) hits.push({ d: d, score: score, pos: firstPos, term: terms[0] });
    }
    hits.sort(function (a, b) { return b.score - a.score; });
    return hits.slice(0, 20);
  }

  function excerpt(body, pos, term) {
    if (pos < 0) return body.slice(0, 100);
    var start = Math.max(0, pos - 40);
    var end = Math.min(body.length, pos + term.length + 60);
    return (start > 0 ? "…" : "") + body.slice(start, end) + (end < body.length ? "…" : "");
  }

  function highlight(text, terms) {
    var frag = document.createDocumentFragment();
    var lower = text.toLowerCase();
    var i = 0;
    while (i < text.length) {
      var best = -1, bestTerm = "";
      for (var j = 0; j < terms.length; j++) {
        var p = lower.indexOf(terms[j], i);
        if (p >= 0 && (best < 0 || p < best)) { best = p; bestTerm = terms[j]; }
      }
      if (best < 0) {
        frag.appendChild(document.createTextNode(text.slice(i)));
        break;
      }
      frag.appendChild(document.createTextNode(text.slice(i, best)));
      var mark = document.createElement("mark");
      mark.textContent = text.slice(best, best + bestTerm.length);
      frag.appendChild(mark);
      i = best + bestTerm.length;
    }
    return frag;
  }

  function render(hits, terms) {
    results.textContent = "";
    if (!hits.length) {
      var empty = document.createElement("div");
      empty.className = "search-empty";
      empty.textContent = "該当なし";
      results.appendChild(empty);
      return;
    }
    hits.forEach(function (h) {
      var a = document.createElement("a");
      a.className = "search-hit";
      a.href = root + h.d.p.u;
      var title = document.createElement("span");
      title.className = "hit-title";
      title.appendChild(highlight(h.d.p.t, terms));
      var ex = document.createElement("span");
      ex.className = "hit-excerpt";
      ex.appendChild(highlight(excerpt(h.d.p.b, h.pos, h.term), terms));
      a.appendChild(title);
      a.appendChild(ex);
      results.appendChild(a);
    });
  }

  input.addEventListener("input", function () {
    var q = input.value.trim();
    if (!q) {
      results.hidden = true;
      nav.hidden = false;
      return;
    }
    var terms = q.toLowerCase().split(/\s+/).filter(Boolean);
    render(search(q), terms);
    results.hidden = false;
    nav.hidden = true;
  });

  input.addEventListener("keydown", function (e) {
    if (e.key === "Escape") {
      input.value = "";
      results.hidden = true;
      nav.hidden = false;
      input.blur();
    }
  });
})();
