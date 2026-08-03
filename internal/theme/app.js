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
