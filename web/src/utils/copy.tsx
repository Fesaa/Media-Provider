/**
 * Copy a string to clipboard
 * @param  {String} s         The string to be copied to clipboard
 * @return {Boolean}               returns a boolean correspondent to the success of the copy operation.
 * @see https://stackoverflow.com/a/53951634/938822
 */
export function copyToClipboard(s: string) {
    let textarea;
    let result;

    try {
        textarea = document.createElement("textarea");
        textarea.setAttribute("readonly", "true");
        textarea.setAttribute("contenteditable", "true");
        textarea.style.position = "fixed"; // prevent scroll from jumping to the bottom when focus is set.
        textarea.value = s;

        document.body.appendChild(textarea);

        textarea.focus();
        textarea.select();

        const range = document.createRange();
        range.selectNodeContents(textarea);

        const sel = window.getSelection();
        if (sel != null) {
            sel.removeAllRanges();
            sel.addRange(range);
        }

        textarea.setSelectionRange(0, textarea.value.length);
        result = document.execCommand("copy");
    } catch (err) {
        console.error(err);
        result = null;
    } finally {
        if (textarea) {
            document.body.removeChild(textarea);
        }
    }

    // manual copy fallback using prompt
    if (!result) {
        const isMac = navigator.platform.toUpperCase().indexOf("MAC") >= 0;
        const copyHotkey = isMac ? "⌘C" : "CTRL+C";
        result = prompt(`Press ${copyHotkey}`, s); // eslint-disable-line no-alert
        if (!result) {
            return false;
        }
    }
    return true;
}