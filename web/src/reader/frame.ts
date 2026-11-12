import type { ReaderBody, Theme } from "@/mygo";
import frameCSS from "./frame.css?inline";

/**
 * The body iframe's own CSP, the first element of its head: nothing loads but inline styles,
 * data: images and the godwit-img: images Go's image proxy serves (a remote image never reaches
 * the network from the page), on top of the page's CSP (ADR 0013).
 */
export const FRAME_CSP = "default-src 'none'; style-src 'unsafe-inline'; img-src data: godwit-img:";

const escapeText = (s: string) => s.replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`);

// A web address in plain text, with what may follow it (linkLength cuts that off).
const URL_RE = /\bhttps?:\/\/[^\s<>"']+/g;

/**
 * How much of a matched address is the link, as core/maillinks.go's mailLinks has it: not the
 * sentence punctuation after it, then not an unbalanced ")" ("text (url)"); a ")" that closes a
 * "(" in it is its own, so "Go_(lang)" keeps it and "a!)" its "!".
 */
function linkLength(u: string): number {
  let end = u.length;
  const open = u.split("(").length - 1;
  let closed = u.split(")").length - 1;
  while (end > 0 && ".,;:!?]}".includes(u[end - 1])) end--;
  while (end > 0 && u[end - 1] === ")" && open < closed) {
    closed--;
    end--;
  }
  return end;
}

// A single-line reply lead ("On … wrote:", "于 … 写道：") placed right above a quote by mail clients.
// Uses explicit start/whitespace/comma boundaries instead of \b so Cyrillic and accented Latin match.
const ATTR_RE =
  /(?:^|[\s,，])(?:wrote\s*:|a écrit\s*:|escribió\s*:|escreveu\s*:|ha scritto\s*:|(?:пишет|написал[аи]?)\s*:|yazdı\s*:|đã viết\s*:|(?:schrieb|schreef)(?=\s).*:)|(?:写道|寫道|書きました)\s*[:：]|(?:^|\s)작성\s*:/i;

function isAttribution(text: string): boolean {
  const trimmed = text.trim();
  if (!trimmed || trimmed.includes("\n")) return false;
  const s = trimmed.replace(/\s+/g, " ");
  return s.length <= 240 && /[:：]$/.test(s) && ATTR_RE.test(s);
}

// A quote folded behind a pill and hairline rule: <details> needs no script in the CSP-locked frame;
// the parent wires the left collapse rail and the hover tooltip.
const DOTS_SVG =
  '<svg class="godwit-quote-dots" width="16" height="10" viewBox="0 0 16 10" fill="currentColor" aria-hidden="true">' +
  '<circle cx="3" cy="5" r="1.45"/><circle cx="8" cy="5" r="1.45"/><circle cx="13" cy="5" r="1.45"/></svg>';
const UP_SVG =
  '<svg class="godwit-quote-up" width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" ' +
  'stroke-width="1.65" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">' +
  '<path d="M3 7.5L6 4.5L9 7.5"/></svg>';
const FOLD_SUMMARY =
  `<summary data-godwit-ui><span class="godwit-quote-pill">${DOTS_SVG}${UP_SVG}</span>` +
  '<span class="godwit-quote-line" aria-hidden="true"></span></summary>';
const FOLD_RAIL =
  '<span class="godwit-quote-rail" data-godwit-ui aria-hidden="true">' +
  '<span class="godwit-quote-rail-line"></span></span>';
const FOLD_OPEN = `<details class="godwit-quote-fold">${FOLD_SUMMARY}<div class="godwit-quote-body">${FOLD_RAIL}`;
const FOLD_CLOSE = "</div></details>";

/**
 * Plain text as HTML: escaped, its web addresses links (data-href, as Go makes an HTML mail's)
 * and its quoted lines ("> …") marked, which the style mutes.
 */
export function plainHTML(text: string, fold = false): string {
  const lines = text.split("\n");
  const quoted = (line: string) => /^\s*>/.test(line);
  // With fold: the quoted lines at the end (and a preceding "On … wrote:" lead) fold behind a
  // button, when the message has words of its own above them.
  let from = lines.length;
  if (fold) {
    while (from > 0 && (quoted(lines[from - 1]) || lines[from - 1].trim() === "")) from--;
    while (from < lines.length && !quoted(lines[from])) from++;
    if (from < lines.length) {
      let lead = from - 1;
      while (lead >= 0 && lines[lead].trim() === "") lead--;
      if (lead >= 0 && isAttribution(lines[lead])) from = lead;
    }
    if (from === lines.length || !lines.slice(0, from).some((l) => l.trim() !== ""))
      from = lines.length;
  }
  const renderLine = (line: string) => {
    let html = "";
    let at = 0;
    for (const m of line.matchAll(URL_RE)) {
      const link = m[0].slice(0, linkLength(m[0]));
      if (!/^https?:\/\/./.test(link)) continue; // nothing but the scheme
      const url = escapeText(link);
      html += `${escapeText(line.slice(at, m.index))}<span data-href="${url}">${url}</span>`;
      at = m.index + link.length;
    }
    html += escapeText(line.slice(at));
    return quoted(line) ? `<span class="godwit-quote">${html}</span>` : html;
  };
  if (from === lines.length) return lines.map(renderLine).join("\n");
  let bodyEnd = from;
  while (bodyEnd > 0 && lines[bodyEnd - 1].trim() === "") bodyEnd--;
  let quoteEnd = lines.length;
  while (quoteEnd > from && lines[quoteEnd - 1].trim() === "") quoteEnd--;
  return (
    lines.slice(0, bodyEnd).map(renderLine).join("\n") +
    FOLD_OPEN +
    lines.slice(from, quoteEnd).map(renderLine).join("\n") +
    FOLD_CLOSE
  );
}

// The words of a node, left out of the count when they are a stylesheet's or a script's.
const visibleWords = (doc: Document, root: Node, keep: (n: Node) => boolean): string => {
  const walker = doc.createTreeWalker(root, NodeFilter.SHOW_TEXT);
  let out = "";
  for (let n = walker.nextNode(); n; n = walker.nextNode()) {
    const tag = n.parentElement?.tagName;
    if (tag !== "STYLE" && tag !== "SCRIPT" && keep(n)) out += n.textContent;
  }
  return out.trim();
};

const BLOCK_TAGS = new Set([
  "ADDRESS",
  "ARTICLE",
  "ASIDE",
  "BLOCKQUOTE",
  "DIV",
  "DL",
  "DT",
  "DD",
  "FIELDSET",
  "FIGCAPTION",
  "FIGURE",
  "FOOTER",
  "H1",
  "H2",
  "H3",
  "H4",
  "H5",
  "H6",
  "HEADER",
  "HR",
  "LI",
  "MAIN",
  "NAV",
  "OL",
  "P",
  "PRE",
  "SECTION",
  "TABLE",
  "TR",
  "UL",
]);

// Visible text of root with <br> and block boundaries preserved as "\n" (while folding raw source
// line wraps inside text nodes to spaces, as HTML layout does), so multi-line blocks are never
// concatenated into a single synthetic attribution line.
function renderedText(root: Node, keep: (n: Node) => boolean = () => true): string {
  let out = "";
  const walk = (n: Node, pre: boolean) => {
    if (!keep(n)) return;
    if (n.nodeType === Node.TEXT_NODE) {
      const raw = n.textContent ?? "";
      out += pre ? raw : raw.replace(/[\r\n]+/g, " ");
      return;
    }
    if (n.nodeType !== Node.ELEMENT_NODE) return;
    const tag = (n as Element).tagName;
    if (tag === "STYLE" || tag === "SCRIPT") return;
    if (tag === "BR") {
      out += "\n";
      return;
    }
    const isBlock = n !== root && BLOCK_TAGS.has(tag);
    if (isBlock && out && !out.endsWith("\n")) out += "\n";
    const inPre = pre || tag === "PRE";
    for (let c = n.firstChild; c; c = c.nextSibling) walk(c, inPre);
    if (isBlock && out && !out.endsWith("\n")) out += "\n";
  };
  walk(root, false);
  return out.trim();
}

function isInlineAttrPart(n: ChildNode): boolean {
  if (n.nodeType === Node.COMMENT_NODE || n.nodeType === Node.TEXT_NODE) return true;
  if (n.nodeType !== Node.ELEMENT_NODE) return false;
  const el = n as Element;
  if (
    el.tagName === "BR" ||
    BLOCK_TAGS.has(el.tagName) ||
    ["IMG", "SVG", "VIDEO", "IFRAME", "STYLE", "SCRIPT"].includes(el.tagName) ||
    hasVisualPaint(el)
  )
    return false;
  return ![...el.querySelectorAll("*")].some(
    (c) =>
      c.tagName === "BR" ||
      BLOCK_TAGS.has(c.tagName) ||
      ["IMG", "SVG", "VIDEO", "IFRAME"].includes(c.tagName) ||
      hasVisualPaint(c),
  );
}

const QUOTE = "blockquote, div.gmail_quote, div.yahoo_quoted, div.protonmail_quote";

function hasVisualPaint(el: Element): boolean {
  return (
    el.hasAttribute("data-bg") ||
    el.hasAttribute("bgcolor") ||
    el.hasAttribute("background") ||
    /(?:^|[;\s])(?:background|border)(?:-[a-z-]+)?\s*:/i.test(el.getAttribute("style") ?? "")
  );
}

function isBlankSpacer(n: ChildNode): boolean {
  if (n.nodeType === Node.TEXT_NODE) return (n.textContent ?? "").trim() === "";
  if (n.nodeType !== Node.ELEMENT_NODE) return false;
  const el = n as Element;
  if (el.tagName === "BR") return true;
  if ((el.tagName !== "DIV" && el.tagName !== "P") || (el.textContent ?? "").trim() !== "")
    return false;
  return ![el, ...el.querySelectorAll("*")].some(
    (c) => (c !== el && c.tagName !== "BR") || hasVisualPaint(c),
  );
}

/**
 * An HTML body with the quote at its end folded behind a button: the last outermost quote block
 * (plus a single-line "On … wrote:" lead right before a bare blockquote), when the message has
 * words of its own before it and none after. An inline reply and a quote-only mail stay open.
 */
export function foldQuotes(html: string): string {
  const doc = new DOMParser().parseFromString(`<!doctype html><body>${html}`, "text/html");
  const last = [...doc.body.querySelectorAll(QUOTE)]
    .filter((q) => !q.parentElement?.closest(QUOTE))
    .at(-1);
  if (!last) return html;
  const after = visibleWords(
    doc,
    doc.body,
    (n) =>
      !last.contains(n) && !!(last.compareDocumentPosition(n) & Node.DOCUMENT_POSITION_FOLLOWING),
  );
  if (after !== "") return html;

  let first: ChildNode = last;
  if (last.tagName === "BLOCKQUOTE") {
    let cur = last.previousSibling;
    while (cur && isBlankSpacer(cur)) cur = cur.previousSibling;
    if (
      cur?.nodeType === Node.ELEMENT_NODE &&
      ["P", "DIV"].includes((cur as Element).tagName) &&
      !(cur as Element).querySelector("p, div, blockquote, img, table, ul, ol, pre, hr") &&
      !hasVisualPaint(cur as Element) &&
      isAttribution(renderedText(cur))
    ) {
      first = cur;
    } else if (cur && isInlineAttrPart(cur)) {
      let start: ChildNode = cur;
      while (start.previousSibling && isInlineAttrPart(start.previousSibling)) {
        start = start.previousSibling;
      }
      let inlineText = "";
      for (let n: ChildNode | null = start; ; n = n.nextSibling) {
        if (n) inlineText += (n.textContent ?? "").replace(/[\r\n]+/g, " ");
        if (!n || n === cur) break;
      }
      if (isAttribution(inlineText)) first = start;
    }
  }

  const before = renderedText(
    doc.body,
    (n) =>
      n === doc.body ||
      (!first.contains(n) &&
        n !== first &&
        !!(first.compareDocumentPosition(n) & Node.DOCUMENT_POSITION_PRECEDING)),
  );
  if (before === "" || isAttribution(before)) return html;

  while (first.previousSibling && isBlankSpacer(first.previousSibling)) {
    first.previousSibling.remove();
  }

  // Mark the outer quote wrapper (and any single-child inner quote that only follows an
  // attribution, as in Godwit's or Gmail's wrapper+blockquote pairs) so only the left rail shows.
  let rootQuote: Element = last;
  while (true) {
    rootQuote.setAttribute("data-quote-root", "");
    const directQuotes = [...rootQuote.querySelectorAll(QUOTE)].filter((q) => {
      let p = q.parentElement;
      while (p && p !== rootQuote) {
        if (p.matches(QUOTE)) return false;
        p = p.parentElement;
      }
      return p === rootQuote;
    });
    if (directQuotes.length !== 1) break;
    const inner = directQuotes[0];
    const leadText = renderedText(
      rootQuote,
      (n) =>
        !inner.contains(n) &&
        !!(inner.compareDocumentPosition(n) & Node.DOCUMENT_POSITION_PRECEDING),
    );
    const tailText = visibleWords(
      doc,
      rootQuote,
      (n) =>
        !inner.contains(n) &&
        !!(inner.compareDocumentPosition(n) & Node.DOCUMENT_POSITION_FOLLOWING),
    );
    if (tailText !== "" || (leadText !== "" && !isAttribution(leadText))) break;
    rootQuote = inner;
  }

  const tpl = doc.createElement("div");
  tpl.innerHTML = `${FOLD_OPEN}${FOLD_CLOSE}`;
  const fold = tpl.firstElementChild as HTMLElement;
  const quoteBody = fold.querySelector(".godwit-quote-body") as HTMLElement;
  first.before(fold);
  let node: ChildNode | null = first;
  while (node) {
    const next: ChildNode | null = node === last ? null : node.nextSibling;
    quoteBody.append(node);
    node = next;
  }
  return doc.body.innerHTML;
}

// A theme value goes into a <style> only if it can't leave its declaration.
const color = (v: string | undefined) => (v && /^[#\w\s(),.%/+-]+$/.test(v) ? v : "");

const size = (px: unknown) =>
  typeof px === "number" && Number.isFinite(px) && px > 0 && px < 200 ? px : 14;

// line is the line height: plain text reads at 1.5; HTML mail sets its own, and where it does not
// (a button, a footer) it gets the browser's normal, as in any mail client, not our reading height.
const base = (px: number, line: string) =>
  `html{font:${px}px/${line} -apple-system,BlinkMacSystemFont,"Helvetica Neue",sans-serif}`;

const htmlCSS = (t: Theme, scheme: string) => {
  const fg = (m: Record<string, string> | undefined) => color(m?.["--foreground"]);
  return (
    `html{color-scheme:${scheme};` +
    `--godwit-fg-light:${fg(t.light) || "initial"};` +
    `--godwit-fg-dark:${fg(t.dark) || "initial"}}`
  );
};

/** Whether the theme is the dark one; systemDark is the page's appearance, for "system". */
export const isDark = (t: Theme, systemDark: boolean) =>
  t.mode === "dark" || (t.mode === "system" && systemDark);

// What lays a mail out rather than writes it: a table or the like, or a painted background, in a
// tag or in its stylesheet.
const LAYOUT =
  /<(?:table|center|iframe|video|svg)\b|<[^>]*(?:background|bgcolor)[^>]*>|<style\b[^>]*>[^<]*background/i;

// A tracking pixel, hidden or a pixel wide, shows nothing.
const shown = (img: string) => !/display:\s*none|(?:^|[\s;"'])width:\s*1px|\bwidth="1"/i.test(img);

/**
 * Whether an HTML body is text in all but name (a plain message sent as text/html, or one a person
 * wrote: paragraphs, lists, links, its own type sizes, no layout and no picture shown): it reads as
 * plain text does, in the reader's text column.
 */
export const isTextHTML = (html: string) =>
  !LAYOUT.test(html) && ![...html.matchAll(/<img\b[^>]*>/gi)].some((m) => shown(m[0]));

// Go marks a stylesheet with author-provided dark rules (core/mail).
const hasDarkStyles = (b: ReaderBody) => b.html.startsWith('<style data-dark="">');

/**
 * The colour scheme of a body's document, also set on its iframe element (a scheme differing from
 * the iframe's would give it an opaque canvas). HTML follows the selected appearance; plain text
 * can follow the system directly.
 */
export function frameScheme(b: ReaderBody, t: Theme, systemDark = false): string {
  if (b.html !== "") return isDark(t, systemDark) ? "dark" : "light";
  return t.mode === "light" || t.mode === "dark" ? t.mode : "light dark";
}

// Plain text: on the page's background (transparent), in the theme's text colour by its mode.
const plainCSS = (t: Theme, scheme: string) => {
  const fg = (m: Record<string, string> | undefined) => color(m?.["--foreground"]);
  const rule = (c: string) => (c ? `:root{--godwit-plain-fg:${c}}` : "");
  const pre = `html{color-scheme:${scheme}}`;
  if (t.mode === "light") return pre + rule(fg(t.light));
  if (t.mode === "dark") return pre + rule(fg(t.dark));
  const dark = rule(fg(t.dark));
  return `${pre}${rule(fg(t.light))}${dark ? `@media (prefers-color-scheme:dark){${dark}}` : ""}`;
};

/**
 * The srcdoc of a body's iframe: the CSP meta first, then our style, then the body (already
 * sanitized by Go; plain text escaped and linked here, plainHTML) in #godwit-body. The theme gives plain text its colour
 * and both kinds the body font size.
 */
export function buildSrcdoc(b: ReaderBody, t: Theme, systemDark = false): string {
  const html = b.html !== "";
  const scheme = frameScheme(b, t, systemDark);
  const css =
    frameCSS +
    base(size(t.bodyFontSize), html ? "normal" : "1.5") +
    (html ? htmlCSS(t, scheme) : plainCSS(t, scheme));
  const dark = isDark(t, systemDark);
  const auto = html && dark && !hasDarkStyles(b);
  const cls =
    (!html
      ? "godwit-plain"
      : dark
        ? hasDarkStyles(b)
          ? "godwit-dark"
          : "godwit-light godwit-auto"
        : "godwit-light") + (html && isTextHTML(b.html) ? " godwit-text" : "");
  return (
    "<!doctype html><html><head>" +
    `<meta http-equiv="Content-Security-Policy" content="${FRAME_CSP}">` +
    '<meta charset="utf-8">' +
    `<style>${css}</style></head>` +
    `<body data-scheme="${scheme}"><div id="godwit-body" class="${cls}"${auto ? " data-auto-dark" : ""}>${html ? (b.collapseQuotes ? foldQuotes(b.html) : b.html) : plainHTML(b.plain, b.collapseQuotes)}</div></body></html>`
  );
}
