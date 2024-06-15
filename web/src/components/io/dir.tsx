import {
  ClipboardIcon, FolderIcon,
  PlusIcon,
} from "@heroicons/react/16/solid";
import React, {ReactNode, useEffect, useState} from "react";
import NotificationHandler from "../../notifications/handler";
import {createNewDir, getSubDirs} from "../../utils/http";
import {NewDirRequest} from "../../utils/types";
import {copyToClipboard} from "../../utils/copy";

function getDirName(s: string): string {
  const parts = s.split("/");
  if (parts.length < 2) {
    return s;
  }

  return parts[parts.length - 1];
}

function getDirUp(s: string) {
    const parts = s.split("/");
    if (parts.length < 2) {
        return s;
    }

    return parts.slice(0, parts.length - 1).join("/");

}

function dirLine(s: string, callBack: () => void): ReactNode {
  return <div
      key={s}
      className="px-2 py-2 border-2 border-solid border-gray-200 bg-white flex flex-row justify-between items-center align-text-bottom"
  >
    <div className="space-x-2 flex flex-row items-center">
      <FolderIcon className="w-6 h-6 text-blue-500" />
      <span
          className="hover:cursor-pointer hover:underline"
          onClick={() => callBack()}
      >
        {getDirName(s)}
      </span>
    </div>
    <ClipboardIcon className="w-4 h-4 hover:cursor-pointer" onClick={() => copyToClipboard(s)} />
  </div>
}

export default function Dir(props: {
  base: string;
  root?: boolean;
}) {
  const [subs, setSubs] = useState<string[]>([]);
  const [curRoot, setCurRoot] = useState<string>(props.base);
  const [root, setRoot] = useState<boolean>(props.root || true);

  useEffect(() => {
    loadSubs(curRoot)
    setRoot(curRoot == props.base)
  }, [curRoot]);

  function loadSubs(dir: string) {
    getSubDirs({dir}).then(setSubs)
        .catch(err => {
          console.debug(err)
          NotificationHandler.addErrorNotificationByTitle("Failed to load subdirectories")
        })
  }

  async function createSubDir() {
    let dirName = prompt("Directory name");
    if (dirName == null || dirName == "") {
      return;
    }
    const req: NewDirRequest = {
      baseDir: curRoot,
      newDir: dirName,
    };
    createNewDir(req).catch(err => {
        console.debug(err)
        NotificationHandler.addErrorNotificationByTitle("Failed to create new directory")
    }).then(() => (
        setSubs([...subs, dirName])
    ))
  }

  return (
      <div className="flex flex-col">
        <span className="text-xl mb-5 flex flex-grow text-center">{props.base}</span>
        <div className="flex flex-col">
          <div className="text-left text-xl"></div>
          {!root && dirLine('...', () => {
            setCurRoot(getDirUp(curRoot))
          })}
          {subs.map(dir => {
            return dirLine(curRoot + "/" + dir, () => {
              setCurRoot(curRoot + "/" + dir)
            })
          })}
        </div>
        {<div className="px-2 py-2 border-2 border-solid border-gray-200 bg-white flex flex-row justify-between items-center align-text-bottom">
          <div className={`flex flex-row text-center items-center`} onClick={createSubDir}>
            <PlusIcon className="w-6 h-6 text-green-500"/>{" "}
            <span className="text-sm hover:underline hover:cursor-pointer">
              Add new folder
            </span>
          </div>
        </div>}
      </div>);
}
