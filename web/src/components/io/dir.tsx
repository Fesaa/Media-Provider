import {
  ClipboardIcon,
  FolderMinusIcon,
  FolderPlusIcon,
  PlusIcon,
} from "@heroicons/react/16/solid";
import axios from "axios";
import React, { useEffect, useState } from "react";
import NotificationHandler from "../../notifications/handler";

function getDirName(s: string): string {
  const parts = s.split("/");
  if (parts.length < 2) {
    return s;
  }

  return parts[-1];
}

/**
 * Copy a string to clipboard
 * @param  {String} string         The string to be copied to clipboard
 * @return {Boolean}               returns a boolean correspondent to the success of the copy operation.
 * @see https://stackoverflow.com/a/53951634/938822
 */
function copyToClipboard(string) {
  let textarea;
  let result;

  try {
    textarea = document.createElement("textarea");
    textarea.setAttribute("readonly", true);
    textarea.setAttribute("contenteditable", true);
    textarea.style.position = "fixed"; // prevent scroll from jumping to the bottom when focus is set.
    textarea.value = string;

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
    document.body.removeChild(textarea);
  }

  // manual copy fallback using prompt
  if (!result) {
    const isMac = navigator.platform.toUpperCase().indexOf("MAC") >= 0;
    const copyHotkey = isMac ? "⌘C" : "CTRL+C";
    result = prompt(`Press ${copyHotkey}`, string); // eslint-disable-line no-alert
    if (!result) {
      return false;
    }
  }
  return true;
}

export default function Dir(props: {
  base: string;
  name: string;
  depth: number;
}) {
  const [open, setOpen] = useState<boolean>(false);
  const [subs, setSubs] = useState<string[]>([]);

  useEffect(() => {
    if (!open) {
      setSubs([]);
    }
  }, [open]);

  async function loadSubs() {
    const data = {
      dir: props.base,
    };
    axios
      .post(`${BASE_URL}/api/io/ls`, data)
      .catch((err) => {
        console.error(err);
        NotificationHandler.addErrorNotificationByTitle("Failed to load directory");
      })
      .then((res) => {
        if (res == null) {
          return;
        }

        if (res.data == null) {
          setOpen(true);
          return;
        }

        setSubs(res.data);
        setOpen(true);
      });
  }

  async function createSubDir() {
    let dirName = prompt("Directory name");
    if (dirName == null || dirName == "") {
      return;
    }

    const data = {
      baseDir: props.base,
      newDir: dirName,
    };

    axios
      .post(`${BASE_URL}/api/io/create`, data)
      .catch((err) => console.error(err))
      .then((res) => {
        setOpen(false);
      });
  }

  function iconFactory() {
    if (open) {
      return <FolderMinusIcon className="w-6 h-6" />;
    }

    return <FolderPlusIcon className="w-6 h-6" />;
  }

  function callBackFactory() {
    if (open) {
      return () => setOpen(false);
    }

    return () => loadSubs();
  }

  return (
    <div className="flex flex-col">
      <div className="flex flex-row space-x-4 text-center items-center">
        <div
          onClick={callBackFactory()}
          className="flex flex-row space-x-4 text-center items-center"
        >
          {iconFactory()}
          {getDirName(props.name)}
        </div>

        <ClipboardIcon
          className="w-4 h-4"
          onClick={() => {
            let path = props.base;
            if (path.startsWith("/")) {
              path = path.substring(1);
            }
            copyToClipboard(path);
          }}
        />
      </div>

      <div className="max-h-64 overflow-x-auto overflow-y-auto">
        {open &&
          subs.map((dir) => {
            return (
              <div
                className={`flex flex-row`}
                style={{ marginLeft: props.depth * 10 }}
                key={dir}
              >
                <Dir
                  base={props.base + "/" + dir}
                  name={dir}
                  depth={props.depth + 1}
                />
              </div>
            );
          })}
        {open && (
          <div
            className={`flex flex-row text-center items-center`}
            style={{ marginLeft: props.depth * 10 }}
            onClick={createSubDir}
          >
            <PlusIcon className="w-6 h-6" />{" "}
            <span className="text-sm hover:underline hover:cursor-pointer">
              Add new folder
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
