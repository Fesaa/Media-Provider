import React from "react";
import NotificationHandler from "../notifications/handler";
import { TrashIcon } from "@heroicons/react/16/solid";
import {InfoStat, StopRequest} from "../utils/types";
import {stopDownload} from "../utils/http";


export default function InfoLine(props: {
  TKey: string;
  infoStat: InfoStat;
  refreshFunc: (repeat: boolean) => void;
}) {
  const infoStat = props.infoStat;

  async function remove(id: string) {
    const stopRequest: StopRequest = {
      provider: infoStat.Provider,
      id: id,
      delete_files: true,
    }
    stopDownload(stopRequest).then(() => {
      NotificationHandler.addSuccesNotificationByTitle("Successfully stop content download");
    }).catch((err) => {
        NotificationHandler.addErrorNotificationByTitle(err.message);
    })
  }

  return (
    <tr
      className="even:bg-white border-gray-300 hover:bg-gray-300"
      key={infoStat.InfoHash}
    >
      <td className="p-2 text-sm">
        <div className="">{infoStat.Name}</div>
      </td>
      <td className="p-2 text-sm text-center hidden md:table-cell">
        {props.infoStat.Size}
      </td>
      <td className="p-2 text-sm text-center">
        {props.infoStat.Completed} % {props.infoStat.Speed && `@ ${props.infoStat.Speed}`}
        <div className="h-2.5 w-full rounded-full bg-gray-200 dark:bg-gray-700  md:block">
          <div
            className="h-2.5 rounded-full bg-blue-600"
            style={{ width: `${props.infoStat.Completed}%` }}
          ></div>
        </div>
      </td>
      <td className="p-2 flex flex-row md:flex-row justify-center">
        <TrashIcon
          className="h-8 w-8"
          type="button"
          onClick={(e) => remove(infoStat.InfoHash)}
          style={{ cursor: "pointer" }}
        />
      </td>
    </tr>
  );
}
