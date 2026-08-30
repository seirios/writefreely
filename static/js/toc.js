// http://localhost:9307/a-ytv9/implementing-notion-like-table-of-contents-in-javascript.html
// ORIGINAL AUTHOR: Krzysztof Kowalczyk
//             URL: https://gist.github.com/kjk/d9343c3f45d9f529b2b8156048254840

function getAllHeaders(optRootID = "") {
    let hdrSel = "%h1, %h2, %h3, %h4, %h5, %h6"
    if (optRootID != "") {
        hdrSel = hdrSel.replaceAll("%", "#" + optRootID + " ");
    } else {
        hdrSel = hdrSel.replaceAll("%", "")
    }
    return Array.from(document.querySelectorAll(hdrSel));
}

function removeHash(str) {
    return str.replace(/#$/, "");
}

class TocItem {
    text = "";
    hLevel = 0;
    nesting = 0;
    element;
    constructor(text, hLevel, nesting, element) {
        this.text = text;
        this.hLevel = hLevel;
        this.nesting = nesting;
        this.element = element;
    }
}

function buildTocItems(optFilter = "", optRootID = "") {
    let allHdrs = getAllHeaders(optRootID);
    let filter = [];
    let hasFilter = optFilter != "";
    if (hasFilter) {
        filter = optFilter.split(",");
    }
    let res = [];
    for (let el of allHdrs) {
        let nskip = 0;
        for (let f of filter) {
            if ((f.startsWith(".") && !el.classList.contains(f.slice(1)))
                || (f.startsWith("#") && el.id != f.slice(1))) {
                nskip++;
            }
        }
        if (hasFilter && nskip == filter.length) {
            continue;
        }
        let text = el.innerText.trim();
        text = removeHash(text);
        text = text.trim();
        let hLevel = parseInt(el.tagName[1]);
        let h = new TocItem(text, hLevel, 0, el);
        res.push(h);
    }
    return res;
}

function fixNesting(hdrs) {
    let n = hdrs.length;
    for (let i = 0; i < n; i++) {
        let h = hdrs[i];
        if (i == 0) {
            h.nesting = 0;
        } else {
            h.nesting = h.hLevel - 1;
        }
    }
}

function genTocMini(items) {
    let tmp = "";
    let t = `<div class="toc-item-mini toc-light">▃</div>`;
    for (let i = 0; i < items.length; i++) {
        tmp += t;
    }
    return `<div class="toc-mini">` + tmp + `</div>`;
}

// Modified from https://stackoverflow.com/a/9756789
function attrEscape(s) {
    return ('' + s) /* Forces the conversion to string. */
        .replace(/\\/g, '\\\\') /* This MUST be the 1st replacement. */
        .replace(/\t/g, '\\t') /* These two replacements protect whitespaces. */
        .replace(/\n/g, '\\n')
        .replace(/&/g, '&amp;') /* These five replacements protect from HTML/XML. */
        .replace(/'/g, '&apos;')
        .replace(/"/g, '&quot;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        ;
}

function genTocList(items, func = "tocGoTo") {
    let tmp = "";
    let t = `<div title="{title}" class="toc-item toc-trunc {ind}" onclick={func}({n})>{text}</div>`;
    let n = 0;
    for (let h of items) {
        let s = t;
        s = s.replace("{func}", func);
        s = s.replace("{n}", n);
        let ind = "toc-ind-" + h.nesting;
        s = s.replace("{ind}", ind);
        s = s.replace("{text}", h.text);
        s = s.replace("{title}", attrEscape(h.text));
        tmp += s;
        n++;
    }
    return `<div class="toc-list">` + tmp + `</div>`;
}

/**
 * @param {HTMLElement} el
 */
function highlightElement(el) {
    let tempBgColor = "var(--toc-flash-color)";
    let origCol = el.style.backgroundColor;
    if (origCol === tempBgColor) {
        return;
    }
    el.style.backgroundColor = tempBgColor;
    setTimeout(() => {
        el.style.backgroundColor = origCol;
    }, 500);
}

let tocItems = [];
function tocGoTo(n) {
    let el = tocItems[n].element;
    let y = el.getBoundingClientRect().top + window.scrollY;
    let offY = 12;
    y -= offY;
    window.scrollTo({
        top: y,
    });
    highlightElement(el);
    // the above scrollTo() triggers updateClosestToc() which might
    // not be accurate so we set the exact selected after a small delay
    setTimeout(() => {
        showSelectedTocItem(n);
    }, 100);
}

function genToc(optFilter = "", optRootID = "", optItems = []) {
    tocItems = buildTocItems(optFilter, optRootID);
    fixNesting(tocItems);
    tocItems = tocItems.concat(optItems);
    const container = document.createElement("div");
    container.className = "toc-wrapper";
    let s = genTocMini(tocItems);
    let s2 = genTocList(tocItems);
    container.innerHTML = s + s2;
    document.body.appendChild(container);
}

function genTocWithItems(ownItems, optFunc = "") {
    tocItems = ownItems;
    const container = document.createElement("div");
    container.className = "toc-wrapper";
    let s = genTocMini(tocItems);
    let s2 = genTocList(tocItems, optFunc);
    container.innerHTML = s + s2;
    document.body.appendChild(container);
}

function showSelectedTocItem(elIdx) {
    // make toc-mini-item black for closest element
    let els = document.querySelectorAll(".toc-item-mini");
    let cls = "toc-light";
    for (let i = 0; i < els.length; i++) {
        let el = els[i];
        if (i == elIdx) {
            el.classList.remove(cls);
        } else {
            el.classList.add(cls);
        }
    }

    // make toc-item bold for closest element
    els = document.querySelectorAll(".toc-item");
    cls = "toc-bold";
    for (let i = 0; i < els.length; i++) {
        let el = els[i];
        if (i == elIdx) {
            el.classList.add(cls);
        } else {
            el.classList.remove(cls);
        }
    }
}
function updateClosestToc() {
    let closestIdx = -1;
    let closestDistance = Infinity;

    for (let i = 0; i < tocItems.length; i++) {
        let tocItem = tocItems[i];
        const rect = tocItem.element.getBoundingClientRect();
        const distanceFromTop = Math.abs(rect.top);
        if (
            distanceFromTop < closestDistance &&
            rect.bottom > 0 &&
            rect.top < window.innerHeight
        ) {
            closestDistance = distanceFromTop;
            closestIdx = i;
        }
    }
    if (closestIdx >= 0) {
        showSelectedTocItem(closestIdx);
    }
}

window.addEventListener("scroll", updateClosestToc);

function initToc(optFilter = "", optRootID = "", optItems = []) {
    genToc(optFilter, optRootID, optItems);
    updateClosestToc();
}

function initTocWithItems(items, func = "", replace = false) {
    if (replace == true) {
        let ell = document.getElementsByClassName("toc-wrapper");
        if (ell.length > 0) {
            ell[0].remove();
        }
    }
    genTocWithItems(items, func);
    updateClosestToc();
}
