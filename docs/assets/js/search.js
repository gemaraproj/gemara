(function () {
  const input = document.getElementById('search-input');
  const results = document.getElementById('search-results');
  if (!input || !results) return;

  const indexUrl = input.dataset.index || '/search.json';
  let index = null;
  let pending = null;

  function loadIndex() {
    if (index) return Promise.resolve();
    if (pending) return pending;
    pending = fetch(indexUrl)
      .then(function (r) { return r.json(); })
      .then(function (data) { index = data; })
      .catch(function (err) {
        console.error('Search index failed to load', err);
        index = [];
      });
    return pending;
  }

  function escapeHTML(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function search(query) {
    if (!index || !query) return [];
    const q = query.toLowerCase();
    const matches = [];
    for (let i = 0; i < index.length; i++) {
      const entry = index[i];
      const title = (entry.title || '').toLowerCase();
      const content = (entry.content || '').toLowerCase();
      const titleMatch = title.indexOf(q) !== -1;
      const contentMatch = content.indexOf(q) !== -1;
      if (!titleMatch && !contentMatch) continue;
      matches.push({
        entry: entry,
        score: (titleMatch ? 10 : 0) + (contentMatch ? 1 : 0),
      });
    }
    matches.sort(function (a, b) { return b.score - a.score; });
    return matches.slice(0, 10).map(function (m) { return m.entry; });
  }

  function render(matches, query) {
    if (!query) {
      results.innerHTML = '';
      results.hidden = true;
      return;
    }
    if (matches.length === 0) {
      results.innerHTML = '<li class="search-empty">No matches for &ldquo;' + escapeHTML(query) + '&rdquo;.</li>';
      results.hidden = false;
      return;
    }
    results.innerHTML = matches
      .map(function (m) {
        return '<li><a href="' + escapeHTML(m.url) + '">' + escapeHTML(m.title) + '</a></li>';
      })
      .join('');
    results.hidden = false;
  }

  input.addEventListener('focus', function () {
    loadIndex().then(function () {
      const q = input.value.trim();
      if (q) render(search(q), q);
    });
  });

  input.addEventListener('input', function () {
    loadIndex().then(function () {
      const q = input.value.trim();
      render(search(q), q);
    });
  });

  // Hide results when focus leaves the search area, but allow link clicks first.
  input.addEventListener('blur', function () {
    setTimeout(function () {
      if (!results.contains(document.activeElement)) {
        results.hidden = true;
      }
    }, 150);
  });

  input.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') {
      input.value = '';
      results.hidden = true;
      input.blur();
    }
  });
})();
